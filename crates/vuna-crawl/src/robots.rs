//! `robots.txt` politeness. The parse-and-check step is a pure function over already-fetched
//! bytes ([`is_allowed`], [`crawl_delay_secs`]) so it is unit-testable without a network round
//! trip; [`RobotsCache`] is the thin stateful wrapper [`crate::HttpFetcher`] uses so it only ever
//! downloads a given host's `robots.txt` once per process lifetime.

use std::collections::HashMap;
use std::sync::Mutex;
use texting_robots::Robot;

/// Parses `robots_txt` for `user_agent` and reports whether `url` may be fetched.
///
/// A malformed `robots.txt` fails **open** (allowed): the file is meant to restrict a well-behaved
/// crawler, and a broken file on the site operator's end must not be able to block the whole site.
/// `texting_robots` itself is deliberately lenient, so this only matters for genuinely unparsable
/// input; a node that wants to fail closed instead can check `Robot::new(..).is_err()` itself.
pub fn is_allowed(robots_txt: &[u8], user_agent: &str, url: &str) -> bool {
    match Robot::new(user_agent, robots_txt) {
        Ok(robot) => robot.allowed(url),
        Err(_) => true,
    }
}

/// The `Crawl-Delay` (in whole seconds, rounded up) `robots_txt` declares for `user_agent`'s
/// group, if any. `None` if there is no directive or the file doesn't parse.
pub fn crawl_delay_secs(robots_txt: &[u8], user_agent: &str) -> Option<u64> {
    let robot = Robot::new(user_agent, robots_txt).ok()?;
    robot.delay.map(|secs| secs.ceil() as u64)
}

/// Caches each host's fetched `robots.txt` body (or the fact that it has none) so a fetcher never
/// downloads it twice. Fetching the body itself is [`crate::HttpFetcher`]'s job — this type only
/// stores what it's given and answers allow/deny + crawl-delay queries against it.
///
/// TODO(agent): entries never expire within a process; a long-lived node should re-fetch
/// `robots.txt` periodically (e.g. once a day) rather than trusting the first fetch forever.
#[derive(Default)]
pub struct RobotsCache {
    /// host -> raw robots.txt bytes, once fetched (or `None` if the host has none / the fetch
    /// errored, in which case everything is allowed and there's no crawl-delay).
    bodies: Mutex<HashMap<String, Option<Vec<u8>>>>,
}

impl RobotsCache {
    pub fn new() -> Self {
        Self::default()
    }

    /// True if we've already fetched (or already tried and failed to fetch) `host`'s `robots.txt`.
    pub fn has(&self, host: &str) -> bool {
        self.bodies.lock().unwrap().contains_key(host)
    }

    /// Store the fetched `robots.txt` body for `host` (or `None` if it doesn't exist / errored).
    pub fn store(&self, host: &str, body: Option<Vec<u8>>) {
        self.bodies.lock().unwrap().insert(host.to_string(), body);
    }

    /// Whether `user_agent` may fetch `url`, using the cached body for `host` if we have one.
    /// Returns `true` (allowed) if `host` isn't cached yet — callers must fetch and [`Self::store`]
    /// a new host's `robots.txt` before this reflects real restrictions for it.
    pub fn allowed(&self, host: &str, user_agent: &str, url: &str) -> bool {
        match self.bodies.lock().unwrap().get(host) {
            Some(Some(body)) => is_allowed(body, user_agent, url),
            Some(None) | None => true,
        }
    }

    /// The cached `Crawl-Delay` for `host`, if any. `None` if there's no cached body, no host
    /// entry, or no directive.
    pub fn crawl_delay(&self, host: &str, user_agent: &str) -> Option<u64> {
        match self.bodies.lock().unwrap().get(host) {
            Some(Some(body)) => crawl_delay_secs(body, user_agent),
            _ => None,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    const TXT: &str = "User-Agent: VunaBot\n\
                        Disallow: /private\n\
                        Allow: /private/ok\n\
                        Crawl-Delay: 3\n\
                        User-Agent: *\n\
                        Disallow: /\n";

    #[test]
    fn disallowed_path_is_blocked() {
        assert!(!is_allowed(TXT.as_bytes(), "VunaBot", "https://example.com/private/secret"));
    }

    #[test]
    fn more_specific_allow_overrides_broader_disallow() {
        assert!(is_allowed(TXT.as_bytes(), "VunaBot", "https://example.com/private/ok"));
    }

    #[test]
    fn unmatched_path_is_allowed() {
        assert!(is_allowed(TXT.as_bytes(), "VunaBot", "https://example.com/public"));
    }

    #[test]
    fn other_user_agents_fall_back_to_the_wildcard_group() {
        // The wildcard group disallows everything; VunaBot has its own group so is unaffected.
        assert!(!is_allowed(TXT.as_bytes(), "SomeOtherBot", "https://example.com/anything"));
    }

    #[test]
    fn malformed_robots_txt_fails_open() {
        let garbage = [0xff, 0xfe, 0x00, 0xff];
        assert!(is_allowed(&garbage, "VunaBot", "https://example.com/x"));
    }

    #[test]
    fn crawl_delay_is_read_for_the_matching_group() {
        assert_eq!(crawl_delay_secs(TXT.as_bytes(), "VunaBot"), Some(3));
    }

    #[test]
    fn crawl_delay_is_none_when_absent() {
        assert_eq!(crawl_delay_secs(b"User-Agent: *\nDisallow: /x\n", "VunaBot"), None);
    }

    #[test]
    fn cache_defaults_to_allowed_before_any_fetch() {
        let cache = RobotsCache::new();
        assert!(!cache.has("example.com"));
        assert!(cache.allowed("example.com", "VunaBot", "https://example.com/private"));
        assert_eq!(cache.crawl_delay("example.com", "VunaBot"), None);
    }

    #[test]
    fn cache_enforces_rules_once_a_body_is_stored() {
        let cache = RobotsCache::new();
        cache.store("example.com", Some(TXT.as_bytes().to_vec()));
        assert!(cache.has("example.com"));
        assert!(!cache.allowed("example.com", "VunaBot", "https://example.com/private/secret"));
        assert!(cache.allowed("example.com", "VunaBot", "https://example.com/public"));
        assert_eq!(cache.crawl_delay("example.com", "VunaBot"), Some(3));
    }

    #[test]
    fn cache_records_hosts_with_no_robots_txt_as_wide_open() {
        let cache = RobotsCache::new();
        cache.store("open.example", None);
        assert!(cache.has("open.example"));
        assert!(cache.allowed("open.example", "VunaBot", "https://open.example/anything"));
    }
}
