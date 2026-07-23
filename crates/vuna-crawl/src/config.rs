//! Fetcher configuration. One struct, one `Default` — every knob a polite crawler needs and
//! nothing it doesn't. `HttpFetcher::new` takes this by value so a node can build a distinct
//! config per extractor kind or per list without touching this crate.

use std::time::Duration;

/// Tunables for [`crate::HttpFetcher`]. All fields are public; construct with `..Default::default()`
/// to override just what you need.
#[derive(Clone, Debug, PartialEq)]
pub struct FetchConfig {
    /// Sent as the `User-Agent` header and as the identity `texting_robots` matches against
    /// `robots.txt` `User-Agent` groups. Should name the bot and link back to an about page —
    /// see the crate's own advice on being a good web citizen.
    pub user_agent: String,
    /// Per-request network timeout (connect + read), passed straight to `reqwest`.
    pub timeout: Duration,
    /// Redirect hop cap. `0` means "don't follow redirects at all".
    pub max_redirects: usize,
    /// Hard cap on response body size. A body that doesn't fit is a [`vuna_core::Error::Fetch`],
    /// never a silent truncation (see [`crate::body::read_capped`]).
    pub max_body_bytes: usize,
    /// Minimum seconds between two requests to the same host, absent a stricter `Crawl-Delay` in
    /// that host's `robots.txt` (the larger of the two always wins — see [`crate::HttpFetcher`]).
    pub min_delay_per_host_secs: u64,
}

impl Default for FetchConfig {
    fn default() -> Self {
        Self {
            user_agent: "VunaBot/0.1 (+https://vulos.org/vuna)".to_string(),
            timeout: Duration::from_secs(20),
            max_redirects: 5,
            max_body_bytes: 8 * 1024 * 1024, // 8 MiB — generous for HTML/JSON, not for media dumps.
            min_delay_per_host_secs: 2,
        }
    }
}
