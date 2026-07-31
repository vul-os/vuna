//! # The three-tier trust discipline
//!
//! Vuna is an open network in which **anyone may contribute index entries about pages you never
//! fetched yourself**. That is the whole value proposition and also the whole attack surface: a
//! poisoned index entry is invisible and symptomless. It does not crash anything, it does not look
//! malformed, and the page it lies about is not retained anywhere to compare against. It simply
//! ranks, and you cannot tell.
//!
//! The defence is not one mechanism, it is a discipline of **three**, and the rule is that every
//! derived artifact must be assigned to exactly one of them and say so. This module exists because
//! the codebase was already applying all three without naming any of them — which is exactly how
//! [`extract::WebDoc`](crate::extract::WebDoc) came to sit in the unprotected middle, checkable by
//! neither, for as long as it did.
//!
//! | Tier | How a claim is checked | What it costs the checker |
//! |---|---|---|
//! | [`Recomputable`](TrustTier::Recomputable) | re-run a deterministic function over bytes you hold, compare the result | you must hold the bytes |
//! | [`Corroborable`](TrustTier::Corroborable) | `k` independent observers agree ([`crate::quorum`]) | you must wait for `k`, and accept `None` when they disagree |
//! | [`LocalOnly`](TrustTier::LocalOnly) | never accepted from a peer at all; rebuilt locally | you pay the compute yourself |
//!
//! ## Scope: derived artifacts only
//!
//! These tiers describe artifacts Vuna *derives* from someone else's content. They do **not**
//! describe published objects that their author signs — [`crate::frontier::UrlList`],
//! [`crate::space::EmbeddingSpace`], [`crate::node::NodeDescriptor`]. Those are checked by
//! signature against a publisher you chose to subscribe to, which is a solved problem and a
//! different one: the question there is *who said this*, not *is this a faithful reading of a page*.
//!
//! ## Why the crawled web is the hard case
//!
//! For a signed object you can defer to its author. For a crawled page **the author signs nothing**
//! — the open web has no participating publisher, no feed, and no key. So a web extraction gets no
//! signature check, and the two remaining tiers must carry it:
//!
//! - **Recomputable** — given the bytes, chunk boundaries, links and the snippet are pure functions
//!   of those bytes, so any holder can re-derive them exactly. Before
//!   [`WebDoc::source_hash`](crate::extract::WebDoc::source_hash) existed this was *true in
//!   principle and unusable in practice*: the body is deliberately not retained, so there was
//!   nothing to recompute against and no way to tell which bytes a peer claimed to have read.
//!   The anchor is what converts the tier from a property into a check.
//! - **Corroborable** — `k` independent crawlers that fetched the same URL and got the same bytes
//!   agree on `(url, source_hash)` ([`crate::quorum::WebObservation`]).
//!
//! ## What this does not solve
//!
//! Stated plainly, because the gap this module closes was itself created by leaving an assumption
//! unwritten:
//!
//! 1. **The live web is not deterministic.** Two honest crawlers fetching the same URL minutes
//!    apart routinely receive different bytes — timestamps, rotating ads, CSRF tokens, A/B splits,
//!    personalisation. For those pages corroboration on an exact byte hash will *fail*, and the
//!    correct behaviour is to report "not corroborated" rather than to loosen the comparison until
//!    something passes. Byte-exact quorum works for stable documents and honestly declines
//!    elsewhere. A fuzzy near-duplicate anchor is a real future design question and is not answered
//!    here.
//! 2. **An anchor is not provenance.** `source_hash` proves an extraction is a faithful reading of
//!    *some* bytes. It does not prove those bytes were ever served by that URL. Only corroboration
//!    by observers who fetched it themselves speaks to that.
//! 3. **Nothing yet enforces the tiers at a network boundary.** There is no admission path that
//!    rejects an unanchored contribution today, because there is no network: Vuna has zero live
//!    nodes and zero bytes of real index. This module is the contract that path will be written
//!    against, and the point of writing it first is that the read path has not been built yet.

use crate::ContentId;

/// Which of the three checking disciplines an artifact is subject to.
///
/// The tier is a property of the *artifact type*, declared on [`TrustAnchored`], not something a
/// sender gets to assert about an individual message. A peer cannot promote its contribution to a
/// weaker tier by labelling it.
#[derive(Clone, Copy, Debug, PartialEq, Eq, PartialOrd, Ord, Hash)]
pub enum TrustTier {
    /// Re-run a deterministic function over bytes you hold and compare the result byte-for-byte.
    /// The strongest tier: it needs no peers, no votes, and no trust in anyone — but it is only
    /// available to a checker that actually has the source bytes.
    Recomputable,
    /// Accept a value when `k` independent, distinct-identity observers agree on it, and accept
    /// *nothing* when they do not. Used where no one authoritative signs the underlying fact.
    Corroborable,
    /// Never accepted from a peer under any circumstances. The artifact is rebuilt locally or not
    /// held. Chosen where a wrong value would be undetectable *and* the rebuild is affordable —
    /// per-space HNSW vector indexes are the case in point: an embedding is a float blob with no
    /// structure to validate, so a poisoned one cannot be spotted by inspection at any price.
    LocalOnly,
}

impl TrustTier {
    /// Every tier, so exhaustiveness can be asserted in tests rather than assumed.
    pub const ALL: [TrustTier; 3] =
        [TrustTier::Recomputable, TrustTier::Corroborable, TrustTier::LocalOnly];

    /// Whether an artifact at this tier may be accepted from a peer **at all**, before any
    /// check is applied. `false` is not "check it harder", it is "do not take it".
    pub const fn accepts_peer_contribution(&self) -> bool {
        match self {
            TrustTier::Recomputable | TrustTier::Corroborable => true,
            TrustTier::LocalOnly => false,
        }
    }

    /// Whether checking at this tier requires more than one observer. Recomputation is a solitary
    /// act; corroboration is not.
    pub const fn requires_peers(&self) -> bool {
        matches!(self, TrustTier::Corroborable)
    }

    /// One line naming what "checking" concretely means here — so the contract states it rather
    /// than leaving each call site to infer it.
    pub const fn check_means(&self) -> &'static str {
        match self {
            TrustTier::Recomputable => {
                "re-derive from bytes you hold and compare; any difference is a rejection"
            }
            TrustTier::Corroborable => {
                "k distinct observers agree, or the answer is None — never a fabricated value"
            }
            TrustTier::LocalOnly => "rebuild it yourself; a peer's copy is not accepted",
        }
    }
}

/// Declares which tier an artifact is checked at, and exposes the bytes-derived anchor it carries
/// (if any).
///
/// Implementing this is how a type states its place in the discipline. A new artifact that cannot
/// answer these two questions has not been thought through yet.
pub trait TrustAnchored {
    /// The tier this artifact is checked at when it arrives from a peer.
    const TIER: TrustTier;

    /// The bytes-derived anchor a recomputation check runs against, or `None` when this artifact
    /// carries none.
    ///
    /// Implementations MUST return `None` for [`ContentId::ZERO`]. The all-zero id is the "unset"
    /// sentinel, and treating it as a value would let an artifact with no anchor compare equal to
    /// another artifact with no anchor — manufacturing agreement out of two absences. `None` and
    /// "a hash that happens to be zero" must never be confusable, which is why
    /// [`crate::extract::WebDoc`] is documented as refusing to deserialize without the field at
    /// all rather than defaulting it.
    fn anchor(&self) -> Option<ContentId>;

    /// Convenience: does this artifact carry a usable anchor?
    fn is_anchor_present(&self) -> bool {
        self.anchor().is_some()
    }
}

/// Normalizes the ZERO sentinel to `None` so no implementation has to remember to.
const fn non_zero(id: ContentId) -> Option<ContentId> {
    // `ContentId::ZERO` means "unset". A const-friendly byte comparison keeps this usable in any
    // context and avoids depending on `PartialEq` being const.
    let mut i = 0;
    while i < 32 {
        if id.0[i] != 0 {
            return Some(id);
        }
        i += 1;
    }
    None
}

impl TrustAnchored for crate::extract::WebDoc {
    /// A crawled page's author signs nothing, so a web extraction cannot be checked by deferring to
    /// anyone. It is **corroborable** as its network-facing tier: `k` observers agreeing on
    /// `(url, source_hash)`. It is *also* recomputable for any checker that holds the bytes — the
    /// anchor is what makes that possible — but a node receiving a contribution generally does not
    /// hold them, so the tier that governs admission is the weaker of the two.
    const TIER: TrustTier = TrustTier::Corroborable;

    fn anchor(&self) -> Option<ContentId> {
        non_zero(self.source_hash)
    }
}

impl TrustAnchored for crate::index::IndexedDoc {
    /// The durable per-URL record carries the extraction's anchor forward, and is checked the same
    /// way for the same reason.
    const TIER: TrustTier = TrustTier::Corroborable;

    fn anchor(&self) -> Option<ContentId> {
        non_zero(self.source_hash)
    }
}

impl TrustAnchored for crate::extract::RetailObservation {
    /// The observed store is a non-participant: it signs nothing, and its stock level is a fact
    /// about the world that no page byte definitively encodes. Corroboration is the only ground
    /// truth available.
    const TIER: TrustTier = TrustTier::Corroborable;

    /// Deliberately `None`. A retail observation is an *interpretation* of a page (a price parsed,
    /// an availability flag read), not a reproduction of it, so no byte hash would let a checker
    /// re-derive the claim. The distinction matters: web extractions and retail observations sit at
    /// the same tier but only one of them has a recomputation path underneath it.
    fn anchor(&self) -> Option<ContentId> {
        None
    }
}

impl TrustAnchored for crate::space::Vector {
    /// Embeddings are never accepted from a peer. A vector is an opaque float blob: a subtly
    /// poisoned one is indistinguishable from a correct one by any local inspection, so there is no
    /// check to perform at any price — and because chunk *text* is retained, re-embedding is local
    /// recompute rather than a re-crawl. That is what makes refusing peer vectors affordable.
    const TIER: TrustTier = TrustTier::LocalOnly;

    fn anchor(&self) -> Option<ContentId> {
        None
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::extract::{Availability, Chunk, Link, RetailMethod, RetailObservation, WebDoc};
    use crate::index::IndexedDoc;
    use crate::space::Vector;

    fn doc(source_hash: ContentId) -> WebDoc {
        WebDoc {
            url: "https://example.com/a".into(),
            title: Some("A".into()),
            snippet: "s".into(),
            chunks: vec![Chunk { ordinal: 0, text: "hello".into() }],
            links: vec![Link { to_url: "https://example.com/b".into(), anchor: "b".into(), rel: None }],
            discovered: vec!["https://example.com/b".into()],
            source_hash,
        }
    }

    /// Coverage assertion: every artifact that crosses the wire has declared a tier, and the
    /// declarations are the ones the contract documents. If a new artifact is added without a
    /// `TrustAnchored` impl it will not appear here — and if someone silently re-tiers an existing
    /// one, this fails.
    #[test]
    fn every_derived_artifact_declares_its_tier() {
        assert_eq!(<WebDoc as TrustAnchored>::TIER, TrustTier::Corroborable);
        assert_eq!(<IndexedDoc as TrustAnchored>::TIER, TrustTier::Corroborable);
        assert_eq!(<RetailObservation as TrustAnchored>::TIER, TrustTier::Corroborable);
        assert_eq!(<Vector as TrustAnchored>::TIER, TrustTier::LocalOnly);
    }

    /// The discipline has exactly three tiers and each means something different. A fourth added
    /// without thought, or two collapsed into one, breaks this.
    #[test]
    fn the_discipline_has_three_distinct_tiers() {
        assert_eq!(TrustTier::ALL.len(), 3);
        let mut seen: Vec<&str> = TrustTier::ALL.iter().map(|t| t.check_means()).collect();
        seen.sort_unstable();
        seen.dedup();
        assert_eq!(seen.len(), 3, "two tiers describe checking identically");
        assert!(TrustTier::ALL.iter().all(|t| !t.check_means().is_empty()));
    }

    /// The defining property of `LocalOnly`: it is refused before any check runs. This is the one
    /// tier where "we validated it" is not a defence.
    #[test]
    fn local_only_artifacts_are_refused_from_peers_outright() {
        assert!(!TrustTier::LocalOnly.accepts_peer_contribution());
        assert!(TrustTier::Recomputable.accepts_peer_contribution());
        assert!(TrustTier::Corroborable.accepts_peer_contribution());
        // Vectors are the artifact that sits there, so a peer's vector is never admissible.
        assert!(!<Vector as TrustAnchored>::TIER.accepts_peer_contribution());
    }

    /// Only corroboration needs other people. Recomputation is checkable alone, which is why it is
    /// the stronger tier even though it is available less often.
    #[test]
    fn only_corroboration_requires_peers() {
        assert!(TrustTier::Corroborable.requires_peers());
        assert!(!TrustTier::Recomputable.requires_peers());
        assert!(!TrustTier::LocalOnly.requires_peers());
    }

    /// The ZERO sentinel is "no anchor", never a matchable value. Two unanchored documents must not
    /// be able to agree with each other by both being empty — that is manufacturing corroboration
    /// out of two absences.
    #[test]
    fn the_zero_sentinel_is_absence_not_a_matchable_anchor() {
        let unanchored = doc(ContentId::ZERO);
        assert_eq!(unanchored.anchor(), None);
        assert!(!unanchored.is_anchor_present());
        assert!(!unanchored.is_anchored());

        // Two unanchored docs have equal `source_hash` fields...
        let other = doc(ContentId::ZERO);
        assert_eq!(unanchored.source_hash, other.source_hash);
        // ...and yet neither yields an anchor, so nothing can be concluded from that equality.
        assert_eq!(other.anchor(), None);
    }

    /// A real anchor — including one that is *nearly* all zeros — passes through untouched. The
    /// sentinel check must be exact, not a "looks mostly empty" heuristic.
    #[test]
    fn a_real_anchor_is_returned_verbatim_even_when_nearly_all_zero() {
        let mut bytes = [0u8; 32];
        bytes[31] = 1; // differs from ZERO in the very last byte only
        let id = ContentId(bytes);
        let d = doc(id);
        assert_eq!(d.anchor(), Some(id));
        assert!(d.is_anchored());

        let mut first = [0u8; 32];
        first[0] = 1;
        assert_eq!(doc(ContentId(first)).anchor(), Some(ContentId(first)));
    }

    /// Carrying the anchor from extraction into the durable record is the point of adding it to
    /// `IndexedDoc`; if `from_web` dropped it, the replicated artifact would be unanchored again.
    #[test]
    fn the_anchor_survives_the_extract_to_index_boundary() {
        let id = ContentId([7u8; 32]);
        let indexed = IndexedDoc::from_web(&doc(id), ContentId([9u8; 32]), 1_000);
        assert_eq!(indexed.source_hash, id);
        assert_eq!(indexed.anchor(), Some(id));

        // And an unanchored extraction cannot launder itself into an anchored index entry.
        let indexed = IndexedDoc::from_web(&doc(ContentId::ZERO), ContentId([9u8; 32]), 1_000);
        assert_eq!(indexed.anchor(), None);
    }

    /// Retail sits at the same tier as web but deliberately offers no anchor: its claim is an
    /// interpretation of the world, not a reproduction of bytes. Asserting this keeps someone from
    /// "fixing" it later by hashing the page and implying a recomputation path that does not exist.
    #[test]
    fn a_retail_observation_offers_no_recomputation_anchor() {
        let obs = RetailObservation {
            store: "shop.example".into(),
            sku: "SKU1".into(),
            availability: Availability::InStock,
            quantity: Some(3),
            price_minor: None,
            currency: None,
            method: RetailMethod::StructuredData,
            observed_at: 100,
        };
        assert_eq!(obs.anchor(), None);
        assert!(!obs.is_anchor_present());
        assert_eq!(<RetailObservation as TrustAnchored>::TIER, TrustTier::Corroborable);
    }

    /// Vectors are local-only and anchorless — the tier exists precisely because there is nothing
    /// to check them against.
    #[test]
    fn vectors_are_local_only_and_carry_no_anchor() {
        let v = Vector { space: "m@4/f32".into(), values: vec![0.1, 0.2, 0.3, 0.4] };
        assert_eq!(v.anchor(), None);
        assert!(!<Vector as TrustAnchored>::TIER.accepts_peer_contribution());
    }
}
