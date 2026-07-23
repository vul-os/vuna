//! Per-host politeness delay. This is pure bookkeeping — a host → last-fetch-time map and a
//! subtraction — with **no sleeping and no network**. [`crate::HttpFetcher`] is the only caller
//! that turns a nonzero [`RateLimiter::wait_secs`] into an actual `thread::sleep`, which is what
//! keeps this module deterministically unit-testable (see the tests below: real clocks and real
//! sleeps never enter the picture).

use std::collections::HashMap;
use std::sync::Mutex;
use vuna_core::UnixSecs;

/// Tracks the last time each host was fetched and decides whether a new request must still wait
/// out that host's minimum delay.
pub struct RateLimiter {
    min_delay: u64,
    last_fetch: Mutex<HashMap<String, UnixSecs>>,
}

impl RateLimiter {
    /// `min_delay_secs` is the floor applied to every host unless a caller passes a larger,
    /// per-request delay to [`Self::wait_secs`] (e.g. a stricter `robots.txt` `Crawl-Delay`).
    pub fn new(min_delay_secs: u64) -> Self {
        Self { min_delay: min_delay_secs, last_fetch: Mutex::new(HashMap::new()) }
    }

    /// How many seconds the caller must still wait before it may fetch `host` at `now`. `0` means
    /// go ahead immediately. `extra_delay` lets a caller enforce a per-host delay stricter than
    /// this limiter's floor (e.g. `robots.txt`'s `Crawl-Delay`) without a second map.
    ///
    /// This does **not** record the fetch — call [`Self::record`] once it actually happens, so a
    /// request that errors before hitting the network doesn't consume the host's delay window.
    pub fn wait_secs(&self, host: &str, now: UnixSecs, extra_delay: u64) -> u64 {
        let delay = self.min_delay.max(extra_delay);
        let guard = self.last_fetch.lock().unwrap();
        match guard.get(host) {
            Some(&last) => last.saturating_add(delay).saturating_sub(now),
            None => 0,
        }
    }

    /// Record that `host` was just fetched at `now`, starting its delay window over.
    pub fn record(&self, host: &str, now: UnixSecs) {
        self.last_fetch.lock().unwrap().insert(host.to_string(), now);
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn first_request_never_waits() {
        let rl = RateLimiter::new(5);
        assert_eq!(rl.wait_secs("example.com", 100, 0), 0);
    }

    #[test]
    fn second_request_before_delay_must_wait_the_remainder() {
        let rl = RateLimiter::new(5);
        rl.record("example.com", 100);
        assert_eq!(rl.wait_secs("example.com", 103, 0), 2);
        assert_eq!(rl.wait_secs("example.com", 105, 0), 0);
        assert_eq!(rl.wait_secs("example.com", 110, 0), 0);
    }

    #[test]
    fn hosts_are_tracked_independently() {
        let rl = RateLimiter::new(5);
        rl.record("a.example", 100);
        assert_eq!(rl.wait_secs("b.example", 100, 0), 0);
    }

    #[test]
    fn extra_delay_overrides_the_floor_when_stricter() {
        let rl = RateLimiter::new(2);
        rl.record("slow.example", 100);
        // robots.txt Crawl-Delay of 10s beats our 2s floor.
        assert_eq!(rl.wait_secs("slow.example", 105, 10), 5);
    }

    #[test]
    fn floor_wins_when_it_is_stricter_than_extra_delay() {
        let rl = RateLimiter::new(10);
        rl.record("slow.example", 100);
        assert_eq!(rl.wait_secs("slow.example", 105, 2), 5);
    }

    #[test]
    fn a_failed_request_does_not_consume_the_delay_window() {
        let rl = RateLimiter::new(5);
        rl.record("example.com", 100);
        // Caller checked wait_secs, decided to proceed, but the fetch itself failed — it must not
        // call record() again, so the original window (100..105) still applies.
        assert_eq!(rl.wait_secs("example.com", 104, 0), 1);
    }
}
