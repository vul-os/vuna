//! Pluggable **extractors** — the heart of "one engine, many verticals". A fetched page goes in;
//! typed extractions come out. The web extractor emits chunks+links (RAG/search); the retail
//! extractor emits price/stock observations (radar). New verticals are new [`Extractor`] impls,
//! adopted by opt-in exactly like embedding spaces.

use crate::{ContentId, UnixSecs};
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
///
/// `Ord` is derived **only** so this can be a deterministic tally key in
/// [`crate::quorum`] (a `BTreeMap` needs it). The ordering carries no meaning — it is never used
/// to rank, prefer, or break a tie between values, and `quorum` refuses ties outright rather than
/// letting declaration order decide. See `quorum::reconcile`'s tie tests.
#[derive(Clone, Copy, Debug, PartialEq, Eq, PartialOrd, Ord, Serialize, Deserialize)]
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
///
/// Because the body is discarded, every other field here is an *assertion by whoever extracted it*.
/// [`source_hash`](Self::source_hash) is the one field that ties those assertions back to something
/// outside the extractor's control — see [`crate::trust`] for the tier this artifact sits at and
/// what checking it does and does not buy.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct WebDoc {
    pub url: String,
    pub title: Option<String>,
    pub snippet: String,
    pub chunks: Vec<Chunk>,
    pub links: Vec<Link>,
    /// URLs discovered here to feed back into the frontier.
    pub discovered: Vec<String>,
    /// **The trust anchor.** `BLAKE3-256` over the exact [`FetchedPage::body`] bytes this
    /// extraction was derived from — no charset decoding, no whitespace normalization, no framing,
    /// so two nodes that fetch this URL and receive the same bytes derive the same value. It is
    /// deliberately a hash of the *bytes alone* and not of `(url, bytes)`: identical bytes served
    /// at two URLs (a mirror, a canonical/alias pair) must hash the same, and the URL is carried
    /// beside it in [`WebDoc::url`] anyway.
    ///
    /// This is **not** [`crate::frontier::UrlEntry::content_hash`]. That one is node-local crawl
    /// state used to skip re-extraction when a page has not changed; it never leaves the node and
    /// nothing checks it. This one travels with the shared artifact and exists to be checked.
    ///
    /// There is no `#[serde(default)]` on purpose. A peer's index entry that omits the field must
    /// fail to deserialize rather than quietly arrive as [`ContentId::ZERO`], because a defaulted
    /// zero would be an artifact with no anchor that still *looks* anchored — precisely the
    /// symptomless poisoning this field exists to prevent. [`ContentId::ZERO`] is likewise refused
    /// as an anchor at the [`crate::trust`] seam, so it cannot be used as a wildcard.
    ///
    /// **What it does not do:** it does not prove the bytes were ever served by that URL, and it
    /// does not make a stale extraction detectable. It makes the extraction *recomputable by
    /// anyone who holds the same bytes*, and it gives independent observers something exact to
    /// agree on ([`crate::quorum::WebObservation`]).
    pub source_hash: ContentId,
}

impl WebDoc {
    /// True when this document carries a usable trust anchor. [`ContentId::ZERO`] is treated as
    /// *absent*, never as a value that could match.
    pub fn is_anchored(&self) -> bool {
        self.source_hash != ContentId::ZERO
    }
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

#[cfg(test)]
mod tests {
    use super::*;

    fn anchored_json(extra: &str) -> String {
        format!(
            r#"{{"url":"https://example.com/a","title":null,"snippet":"s","chunks":[],
                 "links":[],"discovered":[]{extra}}}"#
        )
    }

    /// The wire guard. A contributed document that simply **omits** `source_hash` must be rejected
    /// at parse time. If the field were `#[serde(default)]` it would arrive as `ContentId::ZERO`
    /// and present as a well-formed, anchorless document — an artifact that looks checkable and is
    /// not. Rejecting it here is what keeps "unanchored" from being a silently reachable state.
    #[test]
    fn a_webdoc_without_a_source_hash_is_refused_at_the_wire() {
        let missing = anchored_json("");
        let err = serde_json::from_str::<WebDoc>(&missing).unwrap_err();
        assert!(
            err.to_string().contains("source_hash"),
            "expected a missing-field error naming source_hash, got: {err}"
        );

        // The same payload *with* the field parses, so the rejection is about the anchor and not
        // about the rest of the document being malformed.
        let present = anchored_json(&format!(",\"source_hash\":{:?}", [1u8; 32]));
        let doc: WebDoc = serde_json::from_str(&present).expect("anchored doc must parse");
        assert_eq!(doc.source_hash, crate::ContentId([1u8; 32]));
        assert!(doc.is_anchored());
    }

    /// A doc round-trips through serde with its anchor intact — the anchor must survive
    /// replication, which is the only reason it exists.
    #[test]
    fn the_anchor_round_trips_through_serde() {
        let doc = WebDoc {
            url: "https://example.com/a".into(),
            title: None,
            snippet: "s".into(),
            chunks: vec![Chunk { ordinal: 0, text: "t".into() }],
            links: vec![],
            discovered: vec![],
            source_hash: crate::ContentId([0xab; 32]),
        };
        let back: WebDoc = serde_json::from_str(&serde_json::to_string(&doc).unwrap()).unwrap();
        assert_eq!(back, doc);
        assert_eq!(back.source_hash, crate::ContentId([0xab; 32]));
    }

    /// An explicit all-zero anchor parses (it is well-formed JSON) but must not count as anchored.
    /// Refusing it only at the wire would leave the sentinel reachable by an explicit sender.
    #[test]
    fn an_explicitly_zero_anchor_parses_but_is_not_anchored() {
        let json = anchored_json(&format!(",\"source_hash\":{:?}", [0u8; 32]));
        let doc: WebDoc = serde_json::from_str(&json).unwrap();
        assert!(!doc.is_anchored(), "ZERO must never count as an anchor");
    }
}
