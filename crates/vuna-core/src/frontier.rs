//! The distributed URL frontier: subscribable signed **lists** of URLs, and the per-URL state a
//! node tracks. Rides KOTVA PUB feeds + the DHT (in `vuna-frontier`); core defines only the shapes
//! and the [`Frontier`] trait.

use crate::{space::SpaceId, ContentId, IdentityKey, UnixSecs};
use serde::{Deserialize, Serialize};
use std::collections::BTreeMap;

/// Stable id of a URL: `BLAKE3` of its canonical form. Used as the DHT key for crawl-assignment
/// and replication, so two nodes independently derive the same id for the same URL.
pub type UrlId = ContentId;

/// One entry in the frontier — what to crawl and what we already did to it.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct UrlEntry {
    pub url: String,
    pub id: UrlId,
    /// Last time any node fetched it (None = never crawled).
    pub last_crawled: Option<UnixSecs>,
    /// Content hash of the last fetch — lets a re-crawl skip re-extraction when unchanged.
    pub content_hash: Option<ContentId>,
    /// Per-embedding-space watermark: when this URL's content was last embedded into that space.
    /// This is what lets a *new* space back-fill without a re-crawl (recompute over stored text).
    #[serde(default)]
    pub last_embedded: BTreeMap<SpaceId, UnixSecs>,
    /// Extractor kinds already applied (so a retail-only node skips web-only URLs, etc.).
    #[serde(default)]
    pub applied: Vec<String>,
}

impl UrlEntry {
    pub fn new(url: impl Into<String>, id: UrlId) -> Self {
        Self {
            url: url.into(),
            id,
            last_crawled: None,
            content_hash: None,
            last_embedded: BTreeMap::new(),
            applied: Vec::new(),
        }
    }
}

/// A subscribable list — the ad-filter-list model (à la uBlock / Mwmbl). Signed by its publisher;
/// you choose which to trust. Anyone may publish one; there is no central frontier authority.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct UrlList {
    pub id: ContentId,
    pub name: String,
    pub publisher: IdentityKey,
    /// Monotonic — anti-rollback on the publisher's feed.
    pub version: u64,
    /// The URLs (or, for large lists, a content-address the body is fetched by — see `body`).
    #[serde(default)]
    pub urls: Vec<String>,
    /// For large lists, the list body lives in a separate PUB manifest addressed here.
    #[serde(default)]
    pub body: Option<ContentId>,
    pub updated_at: UnixSecs,
}

/// The node-local view over the frontier. Backed by the DHT + local store in `vuna-frontier`.
pub trait Frontier {
    /// Subscribe to a signed list; its URLs enter the local frontier.
    fn subscribe(&mut self, list: &UrlList) -> crate::Result<()>;

    /// The slice of URLs THIS node is responsible for crawling next, DHT-assigned, staleness-first.
    fn due(&self, now: UnixSecs, limit: usize) -> crate::Result<Vec<UrlEntry>>;

    /// Record the outcome of a crawl so peers (via replication) and this node avoid redundant work.
    fn record(&mut self, entry: &UrlEntry) -> crate::Result<()>;

    /// Insert a newly-discovered URL (from an extractor's links, a feed, or the roadmap visit-app).
    fn discover(&mut self, url: &str, now: UnixSecs) -> crate::Result<UrlId>;
}
