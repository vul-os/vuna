//! The [`Fetcher`] seam and its `reqwest`-backed implementation, [`HttpFetcher`].
//!
//! `HttpFetcher` is the only place in this crate that touches the network. Everything it needs to
//! decide (allowed? how long to wait? how much body to keep?) is pushed into the pure, unit-tested
//! modules [`crate::robots`], [`crate::ratelimit`], and [`crate::body`] — this file is glue plus
//! I/O, deliberately thin so those decisions stay testable without a live server.

use crate::body::read_capped;
use crate::config::FetchConfig;
use crate::ratelimit::RateLimiter;
use crate::robots::RobotsCache;
use reqwest::blocking::Client;
use reqwest::redirect::Policy;
use std::thread;
use std::time::Duration;
use url::Url;
use vuna_core::extract::FetchedPage;
use vuna_core::{Error, Result, UnixSecs};

/// The polite-fetch seam. `now` is supplied by the caller rather than read from the system clock,
/// so a node's crawl loop stays deterministic and replay-safe, and so tests never need real time
/// or a real sleep — see the crate-level doc and the pure modules' own tests for the parts of
/// politeness that ARE covered without network access.
pub trait Fetcher: Send + Sync {
    /// Fetch `url`. `now` is stamped onto the returned [`FetchedPage::fetched_at`] and used for
    /// rate-limiter bookkeeping (and, for a live fetch, to decide whether to sleep first).
    fn fetch(&self, url: &str, now: UnixSecs) -> Result<FetchedPage>;
}

/// A `reqwest`-backed [`Fetcher`]: robots.txt-aware, per-host rate-limited, redirect-hop-capped,
/// body-size-capped.
pub struct HttpFetcher {
    client: Client,
    config: FetchConfig,
    robots: RobotsCache,
    limiter: RateLimiter,
}

impl HttpFetcher {
    /// Builds the underlying `reqwest` client from `config` up front (connection pooling, TLS
    /// setup) — construct one `HttpFetcher` per node/config and reuse it across fetches.
    pub fn new(config: FetchConfig) -> Result<Self> {
        let client = Client::builder()
            .user_agent(config.user_agent.clone())
            .timeout(config.timeout)
            .redirect(Policy::limited(config.max_redirects))
            .build()
            .map_err(|e| Error::Fetch { url: String::new(), reason: format!("client build: {e}") })?;
        let limiter = RateLimiter::new(config.min_delay_per_host_secs);
        Ok(Self { client, config, robots: RobotsCache::new(), limiter })
    }

    /// Host key used for both the robots cache and the rate limiter — scheme is deliberately
    /// excluded (http/https on the same host share one politeness budget) but the port is kept,
    /// since a different port is plausibly a different service.
    fn host_key(url: &Url) -> Option<String> {
        let host = url.host_str()?;
        Some(match url.port() {
            Some(port) => format!("{host}:{port}"),
            None => host.to_string(),
        })
    }

    /// Ensures `self.robots` has (or has tried and failed to get) `host`'s `robots.txt`, fetching
    /// it over the network the first time a host is seen. Best-effort: any failure to fetch
    /// `robots.txt` itself is treated as "no restrictions" (see [`crate::robots::is_allowed`]'s
    /// own fail-open rationale — a broken or absent robots.txt must not block a site outright).
    fn ensure_robots(&self, base: &Url, host: &str) {
        if self.robots.has(host) {
            return;
        }
        let robots_url = match base.join("/robots.txt") {
            Ok(u) => u,
            Err(_) => {
                self.robots.store(host, None);
                return;
            }
        };
        let body = self
            .client
            .get(robots_url)
            .send()
            .ok()
            .filter(|resp| resp.status().is_success())
            .and_then(|resp| read_capped(resp, self.config.max_body_bytes).ok());
        self.robots.store(host, body);
    }
}

impl Fetcher for HttpFetcher {
    fn fetch(&self, url: &str, now: UnixSecs) -> Result<FetchedPage> {
        let parsed = Url::parse(url).map_err(|e| Error::Fetch { url: url.to_string(), reason: format!("invalid URL: {e}") })?;
        let host = Self::host_key(&parsed)
            .ok_or_else(|| Error::Fetch { url: url.to_string(), reason: "URL has no host".to_string() })?;

        // robots.txt: fetch-if-unseen, then enforce.
        self.ensure_robots(&parsed, &host);
        if !self.robots.allowed(&host, &self.config.user_agent, url) {
            return Err(Error::Fetch { url: url.to_string(), reason: "disallowed by robots.txt".to_string() });
        }

        // Politeness delay: the stricter of our configured floor and this host's Crawl-Delay.
        let crawl_delay = self.robots.crawl_delay(&host, &self.config.user_agent).unwrap_or(0);
        let wait = self.limiter.wait_secs(&host, now, crawl_delay);
        if wait > 0 {
            thread::sleep(Duration::from_secs(wait));
        }

        let response = self
            .client
            .get(parsed.clone())
            .send()
            .map_err(|e| Error::Fetch { url: url.to_string(), reason: e.to_string() })?;

        // Record the fetch only once the request actually went out, so an error above (robots
        // disallow, bad URL) never consumes the host's delay window — see ratelimit.rs's tests.
        self.limiter.record(&host, now);

        let status = response.status().as_u16();
        let content_type = response
            .headers()
            .get(reqwest::header::CONTENT_TYPE)
            .and_then(|v| v.to_str().ok())
            .map(|s| s.to_string());

        if !response.status().is_success() {
            return Err(Error::Fetch {
                url: url.to_string(),
                reason: format!("non-2xx status {status}"),
            });
        }

        let body = read_capped(response, self.config.max_body_bytes)
            .map_err(|e| Error::Fetch { url: url.to_string(), reason: e.to_string() })?;

        Ok(FetchedPage { url: url.to_string(), status, content_type, body, fetched_at: now })
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn host_key_includes_nonstandard_port_but_not_scheme() {
        let a = Url::parse("https://example.com/x").unwrap();
        let b = Url::parse("http://example.com/y").unwrap();
        assert_eq!(HttpFetcher::host_key(&a), Some("example.com".to_string()));
        assert_eq!(HttpFetcher::host_key(&a), HttpFetcher::host_key(&b));

        let c = Url::parse("https://example.com:8443/z").unwrap();
        assert_eq!(HttpFetcher::host_key(&c), Some("example.com:8443".to_string()));
        assert_ne!(HttpFetcher::host_key(&a), HttpFetcher::host_key(&c));
    }

    #[test]
    fn host_key_is_none_for_hostless_urls() {
        // `data:` and similar URLs have no host at all.
        let u = Url::parse("data:text/plain,hello").unwrap();
        assert_eq!(HttpFetcher::host_key(&u), None);
    }
}
