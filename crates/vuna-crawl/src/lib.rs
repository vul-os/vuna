//! vuna-crawl — the polite fetch layer.
//!
//! Turns a URL into a [`vuna_core::extract::FetchedPage`] the way a good web citizen should:
//! robots.txt-respecting, per-host rate-limited, timeout- and redirect-hop-capped, with a hard
//! ceiling on body size. Implements [`Fetcher`] (defined here — `vuna-core` has no fetch trait of
//! its own, per the frozen contract) via `reqwest` blocking + `url`.
//!
//! ## Layout
//! - [`fetcher`] — the [`Fetcher`] trait and the `reqwest`-backed [`HttpFetcher`]; the only I/O.
//! - [`config`] — [`config::FetchConfig`], every tunable in one place.
//! - [`robots`] — robots.txt parse/allow/crawl-delay, pure + unit-tested, plus a small cache.
//! - [`ratelimit`] — per-host minimum-delay bookkeeping, pure + unit-tested.
//! - [`body`] — the body-size cap, pure + unit-tested.
//!
//! The politeness *decisions* (robots parsing, rate-limit math, body-cap arithmetic) are kept as
//! pure functions/structs with no network or clock dependency, so they're unit-tested directly;
//! `HttpFetcher` is thin glue over them plus the actual HTTP calls, which is why this crate's
//! tests never touch the network (see each module's own tests).
#![allow(dead_code)]

pub mod body;
pub mod config;
pub mod fetcher;
pub mod ratelimit;
pub mod robots;

pub use config::FetchConfig;
pub use fetcher::{Fetcher, HttpFetcher};
