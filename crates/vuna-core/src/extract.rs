//! Pluggable **extractors** — the heart of "one engine, many verticals". A fetched page goes in;
//! typed extractions come out. The web extractor emits chunks+links (RAG/search); the retail
//! extractor emits price/stock observations (radar). New verticals are new [`Extractor`] impls,
//! adopted by opt-in exactly like embedding spaces.

use crate::UnixSecs;
use serde::{Deserialize, Serialize};

/// A raw fetched page from `vuna-crawl`. Body is bytes so the extractor decides charset/parse.
#[derive(Clone, Debug)]
pub struct FetchedPage {
    pub url: String,
    pub status: u16,
    pub content_type: Option<String>,
    pub body: Vec<u8>,
    pub fetched_at: UnixSecs,
}

/// Which vertical an extractor implements. String-typed so new kinds need no core change.
pub type ExtractorKind = String;

/// A text chunk for embedding + RAG. `text` is stored so re-embedding into a new space is local.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct Chunk {
    /// 0-based index of this chunk within the document.
    pub ordinal: u32,
    pub text: String,
}

/// An outbound link — the edge of the link graph (feeds Min-PPR ranking + graph queries).
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct Link {
    pub to_url: String,
    #[serde(default)]
    pub anchor: String,
    #[serde(default)]
    pub rel: Option<String>,
}

/// Availability signal for the retail vertical — often readable without a hard count.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum Availability {
    InStock,
    OutOfStock,
    LowStock,
    Unknown,
}

/// A retail observation of a *non-participant* store — the subject never signs, so consensus of
/// observers ([`crate::quorum`]) is the only ground truth.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct RetailObservation {
    pub store: String,
    pub sku: String,
    pub availability: Availability,
    pub quantity: Option<u64>,
    pub price_minor: Option<i64>,
    pub currency: Option<String>,
    /// How it was read — JSON-LD/OpenGraph is near-free and preferred over cart-probing.
    pub method: RetailMethod,
    pub observed_at: UnixSecs,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum RetailMethod {
    /// schema.org / Open Graph markup already in the page — parsing, not scraping heuristics.
    StructuredData,
    AvailabilityFlag,
    JsonEndpoint,
    CartProbe,
    Html,
}

/// A web document extraction — chunks for RAG, links for the graph, plus display metadata. The
/// page body itself is NOT retained; `snippet` + `chunks` are the only text kept.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct WebDoc {
    pub url: String,
    pub title: Option<String>,
    pub snippet: String,
    pub chunks: Vec<Chunk>,
    pub links: Vec<Link>,
    /// URLs discovered here to feed back into the frontier.
    pub discovered: Vec<String>,
}

/// The typed output of any extractor. `extend` with new variants as verticals are added.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub enum Extraction {
    Web(WebDoc),
    Retail(Vec<RetailObservation>),
}

/// The pluggable extractor seam. Implementations live in `vuna-extract`.
pub trait Extractor: Send + Sync {
    fn kind(&self) -> ExtractorKind;
    /// True if this extractor should run on the page (by content-type, URL pattern, or markup).
    fn applies(&self, page: &FetchedPage) -> bool;
    fn extract(&self, page: &FetchedPage) -> crate::Result<Extraction>;
}
