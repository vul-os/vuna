//! Quorum reconciliation — the [`Corroborable`](crate::trust::TrustTier::Corroborable) tier of the
//! [three-tier trust discipline](crate::trust).
//!
//! Consensus of `k` independent, anchored observers is the ground truth wherever **nobody
//! authoritative signs the underlying fact**. This module is pure, deterministic logic (no deps)
//! and is fully unit-tested.
//!
//! ## Which verticals need this, and why (corrected)
//!
//! This module used to state that quorum was a retail concern because the observed store is a
//! non-participant that signs nothing, "unlike the web vertical \[where\] there is an authoritative
//! feed to defer to". **That reason does not hold.** On the crawled open web the author signs
//! nothing either: there is no participating publisher, no feed, and no key behind an arbitrary
//! URL. The web vertical was therefore not the well-protected case the comment implied — it had
//! neither a signature to check nor corroboration to fall back on, and sat in the unprotected
//! middle. Both verticals observe non-participants; both need this.
//!
//! The real distinction between them is not *whether* they need corroboration but *what a claim is
//! made of:
//!
//! - A [`RetailObservation`] is an **interpretation** — a stock level is a fact about the world
//!   that no page byte definitively encodes. Corroboration is the only check there is, which is why
//!   [`TrustAnchored::anchor`](crate::trust::TrustAnchored::anchor) is `None` for it.
//! - A [`WebObservation`] is a **reproduction** — an extraction is a deterministic function of
//!   bytes. Observers corroborate `(url, source_hash)`, and because the claim is byte-exact,
//!   agreement here also certifies a *recomputable* artifact: anyone who later obtains those bytes
//!   can re-derive the extraction and detect a lie unilaterally.
//!
//! ## The honest limit of byte-exact web corroboration
//!
//! Two honest crawlers fetching the same URL minutes apart routinely receive *different* bytes —
//! timestamps, rotating ads, CSRF tokens, A/B splits, personalisation. For such pages this
//! reconciler will return `None`, and that is the correct outcome, not a bug to tune away. The
//! caller MUST surface "not corroborated" rather than relax the comparison until something passes.
//! Byte-exact quorum serves stable documents and declines elsewhere; a near-duplicate anchor
//! (shingling, SimHash) is a real and unanswered design question, not something to fake here.

use crate::{
    extract::{Availability, RetailObservation, WebDoc},
    ContentId, IdentityKey, UnixSecs,
};
use serde::{Deserialize, Serialize};

/// Reconciliation policy.
#[derive(Clone, Copy, Debug)]
pub struct QuorumParams {
    /// Minimum distinct anchored observers that must agree to accept a value.
    pub k: usize,
    /// Observations older than this (relative to `now`) are ignored.
    pub window_secs: u64,
    /// Two quantities within this absolute tolerance count as agreeing.
    pub qty_tolerance: u64,
}

impl Default for QuorumParams {
    fn default() -> Self {
        Self { k: 3, window_secs: 3600, qty_tolerance: 2 }
    }
}

/// Anything `k` observers can be asked to agree about.
///
/// One trait method matters more than the others: [`claim`](Observation::claim) returns the
/// **discrete value votes are counted over**. Equality of claims is agreement; there is no
/// similarity, no distance, no fuzz. That is what keeps the reconciler auditable — the answer is
/// either something `k` observers said identically, or nothing.
pub trait Observation {
    /// The value observers vote on. `Ord` is required only so the tally is a deterministic
    /// `BTreeMap`; the ordering is **never** used to rank claims or to break a tie, because a tie
    /// is refused outright rather than resolved.
    type Claim: Clone + Ord;

    /// The discrete claim this observation votes for.
    fn claim(&self) -> Self::Claim;

    /// When this observation was made, for the freshness window.
    fn observed_at(&self) -> UnixSecs;

    /// An optional continuous quantity refined separately by median-within-tolerance. Defaults to
    /// `None` for observation kinds that have no such scalar (web extractions do not).
    fn quantity(&self) -> Option<u64> {
        None
    }

    /// Whether this observation may be counted **at all**, before any tallying.
    ///
    /// Defaults to `true`. Override it where an observation can be structurally void — a
    /// [`WebObservation`] with no anchor is the case in point: `k` peers all sending
    /// [`ContentId::ZERO`] would otherwise "agree" and manufacture a quorum out of `k` absences.
    /// Inadmissible observations are discarded before the distinct-observer count, so they cannot
    /// help reach `k` and cannot appear as dissent either.
    fn is_admissible(&self) -> bool {
        true
    }
}

/// The reconciled result for any [`Observation`] kind.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct AgreedClaim<C> {
    /// The claim `support` distinct observers agreed on.
    pub claim: C,
    /// Present only when enough observers agreed on a count within tolerance. Always `None` for
    /// observation kinds that report no quantity.
    pub quantity: Option<u64>,
    /// How many distinct observers backed the winning claim.
    pub support: usize,
    /// Distinct admissible observers who claimed something else.
    pub dissent: usize,
}

/// The reconciled result for one (store, sku).
#[derive(Clone, Debug, PartialEq)]
pub struct Agreed {
    pub availability: Availability,
    /// Present only when enough observers agreed on a count within tolerance.
    pub quantity: Option<u64>,
    /// How many distinct observers backed the winning availability.
    pub support: usize,
    /// Distinct observers who disagreed with the winning availability.
    pub dissent: usize,
}

/// Reconcile observations of one subject at time `now` — the generic reconciler every vertical
/// uses. [`reconcile`] is the retail-shaped wrapper over it.
///
/// - Inadmissible observations ([`Observation::is_admissible`]) are discarded first, so a
///   structurally void observation can neither help reach `k` nor register as dissent.
/// - Only observations within `window_secs` count.
/// - Each **distinct observer** votes once (its latest in-window observation) — this is the Sybil
///   floor: `k` must be distinct `IdentityKey`s, so ballot-stuffing from one key can't reach quorum.
///   (Anchoring those keys — proof-of-personhood / stake — is the caller's job via KOTVA ATTEST.)
/// - The claim is decided by plurality; a value is only *accepted* if its support ≥ `k`, and a
///   **tie for the plurality is not a decision** — see below.
/// - Quantity is the median of agreeing observers' counts, accepted only if ≥ `k` of them fall
///   within `qty_tolerance` of that median. With an even number of counts the **upper** median is
///   taken, so the value returned is always one an observer actually reported rather than an
///   average of two that nobody saw.
///
/// Returns `None` when the network does not (yet) know, which the caller MUST surface rather than
/// fabricate a value. That is the answer in three cases:
///
/// 1. fewer than `k` distinct admissible observers reported inside the window;
/// 2. no single claim reached `k` distinct backers;
/// 3. two or more claims tied for the most support. A tie is disagreement, not a majority, and
///    breaking it — by enum order, insertion order, or a "conservative" preference for
///    out-of-stock — would be inventing a winner the observations do not contain.
pub fn reconcile_observations<O: Observation>(
    observers: &[(IdentityKey, O)],
    now: UnixSecs,
    params: QuorumParams,
) -> Option<AgreedClaim<O::Claim>> {
    // 1. admissibility + window filter + one latest vote per distinct observer.
    let mut latest: std::collections::BTreeMap<IdentityKey, &O> = Default::default();
    for (ik, obs) in observers {
        // Dropped before anything else: a void observation must not be able to reach quorum with
        // other void observations, nor be counted as a dissenting voice it never earned.
        if !obs.is_admissible() {
            continue;
        }
        if now.saturating_sub(obs.observed_at()) > params.window_secs {
            continue;
        }
        latest
            .entry(*ik)
            .and_modify(|cur| {
                if obs.observed_at() > cur.observed_at() {
                    *cur = obs;
                }
            })
            .or_insert(obs);
    }
    // A cheap early-out, NOT an independent guard: `support` can never exceed `latest.len()`, so
    // the `support < k` check below already implies this one. Removing this line changes no
    // behaviour (it is an equivalent mutant, and no test can distinguish it). It stays because it
    // states the Sybil floor at the point the distinct-observer set is built, where a reader looks
    // for it.
    if latest.len() < params.k {
        return None;
    }

    // 2. claim plurality across distinct observers.
    let mut tally: std::collections::BTreeMap<O::Claim, usize> = Default::default();
    for obs in latest.values() {
        *tally.entry(obs.claim()).or_default() += 1;
    }
    let (winner, support) =
        tally.iter().map(|(c, n)| (c.clone(), *n)).max_by_key(|(_, n)| *n)?;
    if support < params.k {
        return None;
    }
    // A tie is not a plurality. Whichever value an ordering happened to favour would be an
    // artifact of the implementation, not something the observers said.
    if tally.values().filter(|n| **n == support).count() > 1 {
        return None;
    }
    let dissent = latest.len() - support;

    // 3. quantity: median over observers who backed the winning claim AND reported a count.
    let mut qtys: Vec<u64> = latest
        .values()
        .filter(|o| o.claim() == winner)
        .filter_map(|o| o.quantity())
        .collect();
    let quantity = if qtys.len() >= params.k {
        qtys.sort_unstable();
        let median = qtys[qtys.len() / 2];
        let agree = qtys
            .iter()
            .filter(|&&q| q.abs_diff(median) <= params.qty_tolerance)
            .count();
        (agree >= params.k).then_some(median)
    } else {
        None
    };

    Some(AgreedClaim { claim: winner, quantity, support, dissent })
}

impl Observation for RetailObservation {
    /// The claim is the availability value. `Availability`'s `Ord` is a tally key only — see its
    /// type docs; ties are refused, never broken by ordering.
    type Claim = Availability;

    fn claim(&self) -> Availability {
        self.availability
    }
    fn observed_at(&self) -> UnixSecs {
        self.observed_at
    }
    fn quantity(&self) -> Option<u64> {
        self.quantity
    }
}

/// Reconcile retail observations for a single (store, sku) at time `now`.
///
/// A thin projection of [`reconcile_observations`] into retail's shape; see that function for the
/// rules. Every retail observation is admissible, so behaviour here is exactly what it always was.
///
/// A `None` result does **not** mean the network says "unavailable"; [`Availability::Unknown`] is a
/// value observers can themselves report and reach quorum on. The two are deliberately distinct:
/// "nobody agrees" and "everybody agrees they can't tell" are different facts.
pub fn reconcile(
    observers: &[(IdentityKey, RetailObservation)],
    now: UnixSecs,
    params: QuorumParams,
) -> Option<Agreed> {
    reconcile_observations(observers, now, params).map(|a| Agreed {
        availability: a.claim,
        quantity: a.quantity,
        support: a.support,
        dissent: a.dissent,
    })
}

/// One observer's report that fetching `url` yielded bytes hashing to `source_hash`.
///
/// This is the web vertical's unit of corroboration. It carries **no extracted content on
/// purpose**: what observers are asked to agree about is which bytes the URL served, not whose
/// chunker they prefer. Everything downstream of the bytes is a deterministic function of them, so
/// agreement on `(url, source_hash)` is agreement on the extraction — while keeping the ballot
/// small enough that a peer cannot smuggle a differing opinion into it.
///
/// Because `url` is part of the claim, mixing observations of different URLs is safe: they split
/// the tally rather than pooling, so no cross-URL agreement can be manufactured. Callers should
/// still group by URL, since a mixed set makes reaching `k` for any one URL less likely.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct WebObservation {
    pub url: String,
    /// The [`WebDoc::source_hash`] this observer computed. [`ContentId::ZERO`] makes the
    /// observation inadmissible — it is an absence, not a value.
    pub source_hash: ContentId,
    pub observed_at: UnixSecs,
}

impl WebObservation {
    /// The observation an extractor's own output supports.
    pub fn from_doc(doc: &WebDoc, observed_at: UnixSecs) -> Self {
        Self { url: doc.url.clone(), source_hash: doc.source_hash, observed_at }
    }
}

impl Observation for WebObservation {
    /// `(url, source_hash)` — byte-exact. There is no near-match: either two crawlers received the
    /// same bytes from the same URL or they did not.
    type Claim = (String, ContentId);

    fn claim(&self) -> Self::Claim {
        (self.url.clone(), self.source_hash)
    }
    fn observed_at(&self) -> UnixSecs {
        self.observed_at
    }
    /// A web extraction has no continuous scalar to reconcile; the median-within-tolerance path is
    /// unused here and `quantity` on the result is always `None`.
    fn quantity(&self) -> Option<u64> {
        None
    }
    /// An unanchored observation is void. Without this, `k` peers each sending
    /// [`ContentId::ZERO`] would agree with one another and reach quorum on nothing at all —
    /// corroboration manufactured out of `k` absences.
    fn is_admissible(&self) -> bool {
        self.source_hash != ContentId::ZERO
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::extract::RetailMethod;

    fn obs(av: Availability, qty: Option<u64>, at: UnixSecs) -> RetailObservation {
        RetailObservation {
            store: "shop.example".into(),
            sku: "SKU1".into(),
            availability: av,
            quantity: qty,
            price_minor: None,
            currency: None,
            method: RetailMethod::StructuredData,
            observed_at: at,
        }
    }
    fn ik(n: u8) -> IdentityKey {
        IdentityKey([n; 32])
    }

    #[test]
    fn below_k_distinct_is_none() {
        let v = vec![(ik(1), obs(Availability::InStock, Some(5), 100))];
        assert!(reconcile(&v, 100, QuorumParams { k: 3, ..Default::default() }).is_none());
    }

    #[test]
    fn one_observer_cannot_stuff_the_ballot() {
        // Same key voting many times still counts once → never reaches k=3.
        let v: Vec<_> = (0..9).map(|i| (ik(1), obs(Availability::InStock, Some(5), 100 + i))).collect();
        assert!(reconcile(&v, 200, QuorumParams { k: 3, window_secs: 3600, qty_tolerance: 2 }).is_none());
    }

    #[test]
    fn three_distinct_agree_on_availability_and_qty() {
        let v = vec![
            (ik(1), obs(Availability::InStock, Some(10), 100)),
            (ik(2), obs(Availability::InStock, Some(11), 100)),
            (ik(3), obs(Availability::InStock, Some(9), 100)),
        ];
        let a = reconcile(&v, 100, QuorumParams { k: 3, window_secs: 3600, qty_tolerance: 2 }).unwrap();
        assert_eq!(a.availability, Availability::InStock);
        assert_eq!(a.quantity, Some(10));
        assert_eq!(a.support, 3);
        assert_eq!(a.dissent, 0);
    }

    #[test]
    fn stale_observations_are_ignored() {
        let v = vec![
            (ik(1), obs(Availability::InStock, Some(10), 0)),
            (ik(2), obs(Availability::InStock, Some(10), 0)),
            (ik(3), obs(Availability::InStock, Some(10), 5000)),
        ];
        // window 3600, now 5000 → first two are stale, only 1 fresh → None.
        assert!(reconcile(&v, 5000, QuorumParams { k: 3, window_secs: 3600, qty_tolerance: 2 }).is_none());
    }

    #[test]
    fn disagreement_lowers_confidence_but_plurality_wins() {
        let v = vec![
            (ik(1), obs(Availability::InStock, None, 100)),
            (ik(2), obs(Availability::InStock, None, 100)),
            (ik(3), obs(Availability::InStock, None, 100)),
            (ik(4), obs(Availability::OutOfStock, None, 100)),
        ];
        let a = reconcile(&v, 100, QuorumParams { k: 3, window_secs: 3600, qty_tolerance: 2 }).unwrap();
        assert_eq!(a.availability, Availability::InStock);
        assert_eq!(a.support, 3);
        assert_eq!(a.dissent, 1);
        assert_eq!(a.quantity, None); // no counts reported
    }

    // ---- edge cases of the median-within-tolerance rule ----------------------------------------
    //
    // This is the part of the reconciler that is genuinely its own mathematics rather than a
    // restatement of a signature threshold: an estimator over a *continuous* quantity, from an
    // open observer set of unknown size, with a robustness band. The cases below are the ones that
    // separate it from "count the signatures and compare to a threshold".

    fn params(k: usize, qty_tolerance: u64) -> QuorumParams {
        QuorumParams { k, window_secs: 3600, qty_tolerance }
    }

    fn votes(entries: &[(u8, Availability, Option<u64>)]) -> Vec<(IdentityKey, RetailObservation)> {
        entries.iter().map(|(id, av, qty)| (ik(*id), obs(*av, *qty, 100))).collect()
    }

    /// A tie for the plurality is disagreement, not a decision. Two-two is the smallest case and
    /// the one an enum-ordering or insertion-ordering tiebreak would silently "resolve".
    #[test]
    fn an_even_split_has_no_plurality_and_reconciles_to_nothing() {
        let v = votes(&[
            (1, Availability::InStock, None),
            (2, Availability::InStock, None),
            (3, Availability::OutOfStock, None),
            (4, Availability::OutOfStock, None),
        ]);
        assert!(reconcile(&v, 100, params(2, 2)).is_none(), "2-2 is a tie, not a majority for either value");

        // Nudging one vote across breaks the tie and the same input now decides.
        let v = votes(&[
            (1, Availability::InStock, None),
            (2, Availability::InStock, None),
            (3, Availability::InStock, None),
            (4, Availability::OutOfStock, None),
        ]);
        let a = reconcile(&v, 100, params(2, 2)).unwrap();
        assert_eq!(a.availability, Availability::InStock);
        assert_eq!((a.support, a.dissent), (3, 1));
    }

    /// A three-way tie is refused for the same reason, and specifically must not fall through to
    /// whichever `Availability` variant happens to sort last.
    #[test]
    fn a_three_way_tie_is_refused_not_resolved_by_enum_order() {
        let v = votes(&[
            (1, Availability::InStock, None),
            (2, Availability::OutOfStock, None),
            (3, Availability::Unknown, None),
        ]);
        assert!(reconcile(&v, 100, params(1, 2)).is_none());
    }

    /// "Everyone agrees they can't tell" is a real, reportable outcome — distinct from `None`,
    /// which means "the observers do not agree at all".
    #[test]
    fn unanimous_unknown_is_an_agreed_value_not_a_failure() {
        let v = votes(&[
            (1, Availability::Unknown, None),
            (2, Availability::Unknown, None),
            (3, Availability::Unknown, None),
        ]);
        let a = reconcile(&v, 100, params(3, 2)).unwrap();
        assert_eq!(a.availability, Availability::Unknown);
        assert_eq!(a.support, 3);
    }

    /// With an even number of counts the upper median is taken, so the reported number is one an
    /// observer actually saw rather than an interpolation nobody reported.
    #[test]
    fn even_numbered_counts_take_the_upper_median() {
        let v = votes(&[
            (1, Availability::InStock, Some(10)),
            (2, Availability::InStock, Some(11)),
            (3, Availability::InStock, Some(12)),
            (4, Availability::InStock, Some(13)),
        ]);
        let a = reconcile(&v, 100, params(4, 2)).unwrap();
        assert_eq!(a.quantity, Some(12), "upper median of [10,11,12,13]");
        // …and it is a value an observer reported, which an average (11.5) would not be.
        assert!([10, 11, 12, 13].contains(&a.quantity.unwrap()));
    }

    /// The robustness the median buys: one wildly wrong observer cannot move the answer, where a
    /// mean would be dragged a long way by it.
    #[test]
    fn a_single_outlier_cannot_move_the_median() {
        let counts = [10u64, 10, 11, 9_000_000];
        let v = votes(&[
            (1, Availability::InStock, Some(counts[0])),
            (2, Availability::InStock, Some(counts[1])),
            (3, Availability::InStock, Some(counts[2])),
            (4, Availability::InStock, Some(counts[3])),
        ]);
        let a = reconcile(&v, 100, params(3, 2)).unwrap();

        let mean = counts.iter().sum::<u64>() / counts.len() as u64;
        assert_eq!(mean, 2_250_007, "a mean would be dragged three orders of magnitude by one liar");
        assert_eq!(a.quantity, Some(11), "the median stays with the cluster the honest observers reported");
        assert_eq!(a.support, 4, "the outlier still backed the winning availability — only its count is ignored");
    }

    /// The tolerance band is a floor on agreement, not a filter: if the counts are simply too
    /// scattered for `k` of them to sit within tolerance of the median, no quantity is reported at
    /// all — even though availability is agreed and the median exists.
    #[test]
    fn scattered_counts_outside_tolerance_report_no_quantity() {
        let v = votes(&[
            (1, Availability::InStock, Some(3)),
            (2, Availability::InStock, Some(40)),
            (3, Availability::InStock, Some(900)),
        ]);
        let a = reconcile(&v, 100, params(3, 2)).unwrap();
        assert_eq!(a.availability, Availability::InStock, "availability is still agreed");
        assert_eq!(a.quantity, None, "no k counts agree within tolerance");

        // Widening the tolerance to cover the spread accepts the same observations.
        let a = reconcile(&v, 100, params(3, 1_000)).unwrap();
        assert_eq!(a.quantity, Some(40));
    }

    /// Exactly `k` counts within tolerance is enough; one fewer is not. The boundary is checked
    /// from both sides so an off-by-one in either direction fails this test.
    #[test]
    fn quantity_agreement_boundary_is_exactly_k() {
        // Median of [9,10,11,99] (upper median) is 11; 9, 10 and 11 are all within 2 of it.
        let three_agree = votes(&[
            (1, Availability::InStock, Some(9)),
            (2, Availability::InStock, Some(10)),
            (3, Availability::InStock, Some(11)),
            (4, Availability::InStock, Some(99)),
        ]);
        assert_eq!(reconcile(&three_agree, 100, params(3, 2)).unwrap().quantity, Some(11));
        assert_eq!(reconcile(&three_agree, 100, params(4, 2)).unwrap().quantity, None, "k=4 needs a fourth");

        // Tolerance 1 leaves only 10 and 11 within the band: two, one short of k=3.
        assert_eq!(reconcile(&three_agree, 100, params(3, 1)).unwrap().quantity, None);
    }

    /// Counts are pooled only from observers who backed the *winning* availability. A dissenter's
    /// number is not evidence about a state it did not report.
    #[test]
    fn dissenting_observers_counts_are_not_pooled_into_the_median() {
        let v = votes(&[
            (1, Availability::InStock, Some(10)),
            (2, Availability::InStock, Some(10)),
            (3, Availability::InStock, Some(10)),
            (4, Availability::OutOfStock, Some(0)),
            (5, Availability::OutOfStock, Some(0)),
        ]);
        let a = reconcile(&v, 100, params(3, 2)).unwrap();
        assert_eq!(a.availability, Availability::InStock);
        assert_eq!((a.support, a.dissent), (3, 2));
        assert_eq!(a.quantity, Some(10), "the two zeros belong to a different availability");
    }

    /// Availability can reach quorum while quantity does not: too few of the agreeing observers
    /// reported a count at all. Partial knowledge is reported as partial, never padded out.
    #[test]
    fn insufficient_quantity_reports_leave_availability_agreed_and_quantity_unknown() {
        let v = votes(&[
            (1, Availability::InStock, Some(10)),
            (2, Availability::InStock, Some(10)),
            (3, Availability::InStock, None),
            (4, Availability::InStock, None),
        ]);
        let a = reconcile(&v, 100, params(3, 2)).unwrap();
        assert_eq!(a.availability, Availability::InStock);
        assert_eq!(a.support, 4);
        assert_eq!(a.quantity, None, "only 2 of 4 reported a count; k is 3");
    }

    /// A reported zero is a fact ("we looked, there are none"), not a missing value, and must not
    /// be conflated with `None`.
    #[test]
    fn a_reported_zero_count_is_distinct_from_no_count() {
        let v = votes(&[
            (1, Availability::OutOfStock, Some(0)),
            (2, Availability::OutOfStock, Some(0)),
            (3, Availability::OutOfStock, Some(0)),
        ]);
        assert_eq!(reconcile(&v, 100, params(3, 0)).unwrap().quantity, Some(0));

        let v = votes(&[
            (1, Availability::OutOfStock, None),
            (2, Availability::OutOfStock, None),
            (3, Availability::OutOfStock, None),
        ]);
        assert_eq!(reconcile(&v, 100, params(3, 0)).unwrap().quantity, None);
    }

    /// One observer's later reading replaces its earlier one — it never gets two votes by
    /// observing twice, which is the same Sybil floor viewed from the time axis.
    #[test]
    fn an_observers_latest_reading_supersedes_its_earlier_one() {
        let v = vec![
            (ik(1), obs(Availability::InStock, Some(50), 100)),
            (ik(1), obs(Availability::OutOfStock, Some(0), 200)),
            (ik(2), obs(Availability::OutOfStock, Some(0), 200)),
            (ik(3), obs(Availability::OutOfStock, Some(0), 200)),
        ];
        let a = reconcile(&v, 200, params(3, 2)).unwrap();
        assert_eq!(a.availability, Availability::OutOfStock);
        assert_eq!(a.support, 3);
        assert_eq!(a.dissent, 0, "observer 1 has one vote, its latest");
        assert_eq!(a.quantity, Some(0));
    }

    /// The window is a hard cut-off at exactly `window_secs`, inclusive.
    #[test]
    fn the_freshness_window_boundary_is_inclusive() {
        let at_edge = vec![
            (ik(1), obs(Availability::InStock, None, 1_400)),
            (ik(2), obs(Availability::InStock, None, 1_400)),
            (ik(3), obs(Availability::InStock, None, 1_400)),
        ];
        // now - observed_at == 3600 exactly: still in.
        assert!(reconcile(&at_edge, 5_000, params(3, 2)).is_some());
        // One second older: out, and quorum collapses.
        assert!(reconcile(&at_edge, 5_001, params(3, 2)).is_none());
    }

    /// Reconciliation is a pure function of the observation *set*: the order they arrived in
    /// cannot change the answer.
    #[test]
    fn the_result_does_not_depend_on_observation_order() {
        let mut v = votes(&[
            (1, Availability::InStock, Some(10)),
            (2, Availability::InStock, Some(12)),
            (3, Availability::InStock, Some(11)),
            (4, Availability::OutOfStock, Some(0)),
        ]);
        let expected = reconcile(&v, 100, params(3, 2)).unwrap();
        v.reverse();
        assert_eq!(reconcile(&v, 100, params(3, 2)).unwrap(), expected);
        v.swap(0, 2);
        assert_eq!(reconcile(&v, 100, params(3, 2)).unwrap(), expected);
    }

    #[test]
    fn no_observations_at_all_is_none() {
        assert!(reconcile(&[], 100, params(3, 2)).is_none());
    }
}

// ================================================================================================
// The web vertical, corroborated through the SAME reconciler. Everything above is retail and is
// unchanged; these tests exercise the widening.
// ================================================================================================
#[cfg(test)]
mod web_tests {
    use super::*;
    use crate::extract::{Chunk, WebDoc};

    fn ik(n: u8) -> IdentityKey {
        IdentityKey([n; 32])
    }
    fn cid(n: u8) -> ContentId {
        ContentId([n; 32])
    }
    fn params(k: usize) -> QuorumParams {
        QuorumParams { k, window_secs: 3600, qty_tolerance: 2 }
    }
    fn obs(url: &str, hash: ContentId, at: UnixSecs) -> WebObservation {
        WebObservation { url: url.into(), source_hash: hash, observed_at: at }
    }
    const URL: &str = "https://example.com/stable";

    /// The headline case the web vertical previously could not express at all: independent
    /// crawlers that fetched the same URL and received the same bytes corroborate the extraction.
    #[test]
    fn observers_seeing_the_same_bytes_corroborate_a_web_extraction() {
        let v = vec![
            (ik(1), obs(URL, cid(0xAA), 100)),
            (ik(2), obs(URL, cid(0xAA), 110)),
            (ik(3), obs(URL, cid(0xAA), 120)),
        ];
        let a = reconcile_observations(&v, 200, params(3)).unwrap();
        assert_eq!(a.claim, (URL.to_string(), cid(0xAA)));
        assert_eq!(a.support, 3);
        assert_eq!(a.dissent, 0);
        assert_eq!(a.quantity, None, "a web extraction has no scalar to reconcile");
    }

    /// One node cannot corroborate its own extraction, however many times it reports it. The Sybil
    /// floor is the same one retail relies on, now covering the web.
    #[test]
    fn a_node_cannot_corroborate_its_own_extraction() {
        let v: Vec<_> = (0..9).map(|i| (ik(1), obs(URL, cid(0xAA), 100 + i))).collect();
        assert!(reconcile_observations(&v, 200, params(3)).is_none());

        // Two real peers still is not three.
        let v = vec![
            (ik(1), obs(URL, cid(0xAA), 100)),
            (ik(2), obs(URL, cid(0xAA), 100)),
        ];
        assert!(reconcile_observations(&v, 200, params(3)).is_none());
    }

    /// The poisoning case this widening exists to stop: a peer contributing an extraction whose
    /// bytes nobody else saw gets no quorum, so it cannot enter the index as corroborated.
    #[test]
    fn a_lone_peer_reporting_different_bytes_is_outvoted_and_flagged_as_dissent() {
        let v = vec![
            (ik(1), obs(URL, cid(0xAA), 100)),
            (ik(2), obs(URL, cid(0xAA), 100)),
            (ik(3), obs(URL, cid(0xAA), 100)),
            (ik(4), obs(URL, cid(0xBB), 100)), // the liar (or a genuinely different fetch)
        ];
        let a = reconcile_observations(&v, 100, params(3)).unwrap();
        assert_eq!(a.claim.1, cid(0xAA), "the forged bytes did not win");
        assert_eq!((a.support, a.dissent), (3, 1), "the disagreement is reported, not hidden");
    }

    /// **A plurality is not a quorum.** Enough observers reported, and one claim clearly leads with
    /// no tie — but its support is still below `k`, so there is no answer. Found by mutation
    /// testing: deleting the `support >= k` floor passed the entire suite, retail tests included,
    /// because every existing case was already stopped by the distinct-observer floor or by the
    /// tie rule. Nothing pinned the floor itself.
    #[test]
    fn a_plurality_that_does_not_reach_k_is_refused() {
        let v = vec![
            (ik(1), obs(URL, cid(0xAA), 100)),
            (ik(2), obs(URL, cid(0xAA), 100)),
            (ik(3), obs(URL, cid(0xBB), 100)),
            (ik(4), obs(URL, cid(0xCC), 100)),
        ];
        // 4 distinct observers clear the observer floor; 0xAA leads 2-1-1 with no tie for the
        // plurality; but 2 < k = 3, so the network does not know.
        assert!(
            reconcile_observations(&v, 100, params(3)).is_none(),
            "a 2-of-4 lead is a plurality, not a quorum of 3"
        );
        // Lowering k to what the observations actually support accepts the very same input.
        let a = reconcile_observations(&v, 100, params(2)).unwrap();
        assert_eq!(a.claim.1, cid(0xAA));
        assert_eq!((a.support, a.dissent), (2, 2));
    }

    /// The same floor, through the retail wrapper — the widening must not have left retail's
    /// quorum threshold unpinned either.
    #[test]
    fn a_retail_plurality_that_does_not_reach_k_is_also_refused() {
        use crate::extract::{Availability, RetailMethod, RetailObservation};
        let mk = |av| RetailObservation {
            store: "shop.example".into(),
            sku: "SKU1".into(),
            availability: av,
            quantity: None,
            price_minor: None,
            currency: None,
            method: RetailMethod::StructuredData,
            observed_at: 100,
        };
        let v = vec![
            (ik(1), mk(Availability::InStock)),
            (ik(2), mk(Availability::InStock)),
            (ik(3), mk(Availability::OutOfStock)),
            (ik(4), mk(Availability::LowStock)),
        ];
        assert!(reconcile(&v, 100, params(3)).is_none(), "InStock leads 2-1-1; 2 < k = 3");
        let a = reconcile(&v, 100, params(2)).unwrap();
        assert_eq!(a.availability, Availability::InStock);
        assert_eq!((a.support, a.dissent), (2, 2));
    }

    /// **The honest limitation, asserted rather than glossed.** On a volatile page every honest
    /// observer sees different bytes, so byte-exact corroboration yields nothing. `None` here is
    /// correct behaviour: the caller must report "not corroborated" rather than loosen the
    /// comparison until some answer appears.
    #[test]
    fn a_volatile_page_reaches_no_quorum_and_that_is_the_correct_answer() {
        let v = vec![
            (ik(1), obs(URL, cid(0x01), 100)),
            (ik(2), obs(URL, cid(0x02), 100)),
            (ik(3), obs(URL, cid(0x03), 100)),
            (ik(4), obs(URL, cid(0x04), 100)),
            (ik(5), obs(URL, cid(0x05), 100)),
        ];
        assert!(
            reconcile_observations(&v, 100, params(3)).is_none(),
            "five observers, five different byte streams: nothing is corroborated"
        );
    }

    /// Unanchored observations are void. Without the admissibility filter, `k` peers each sending
    /// ZERO would agree with each other and manufacture a quorum out of `k` absences — the exact
    /// failure the ZERO sentinel exists to prevent.
    #[test]
    fn unanchored_observations_cannot_manufacture_agreement() {
        let all_zero = vec![
            (ik(1), obs(URL, ContentId::ZERO, 100)),
            (ik(2), obs(URL, ContentId::ZERO, 100)),
            (ik(3), obs(URL, ContentId::ZERO, 100)),
            (ik(4), obs(URL, ContentId::ZERO, 100)),
        ];
        assert!(
            reconcile_observations(&all_zero, 100, params(3)).is_none(),
            "four unanchored observers agreed on nothing and must reach no quorum"
        );

        // Nor can void observations pad a real minority up to k.
        let padded = vec![
            (ik(1), obs(URL, cid(0xAA), 100)),
            (ik(2), obs(URL, cid(0xAA), 100)),
            (ik(3), obs(URL, ContentId::ZERO, 100)),
            (ik(4), obs(URL, ContentId::ZERO, 100)),
        ];
        assert!(
            reconcile_observations(&padded, 100, params(3)).is_none(),
            "two real observers plus two absences is not three observers"
        );

        // …and a void observation is not counted as dissent either: with three real agreeing
        // observers the result is unanimous, not 3-against-1.
        let with_void = vec![
            (ik(1), obs(URL, cid(0xAA), 100)),
            (ik(2), obs(URL, cid(0xAA), 100)),
            (ik(3), obs(URL, cid(0xAA), 100)),
            (ik(4), obs(URL, ContentId::ZERO, 100)),
        ];
        let a = reconcile_observations(&with_void, 100, params(3)).unwrap();
        assert_eq!((a.support, a.dissent), (3, 0));
    }

    /// The URL is part of the claim, so observations of different pages split the tally instead of
    /// pooling into a false agreement.
    #[test]
    fn observations_of_different_urls_do_not_pool_into_one_claim() {
        let v = vec![
            (ik(1), obs("https://example.com/a", cid(0xAA), 100)),
            (ik(2), obs("https://example.com/b", cid(0xAA), 100)),
            (ik(3), obs("https://example.com/c", cid(0xAA), 100)),
        ];
        assert!(
            reconcile_observations(&v, 100, params(2)).is_none(),
            "same bytes at three different URLs is three separate claims, none with k=2 support"
        );

        // Two of the three agreeing on one URL does reach k=2, and the third is dissent.
        let v = vec![
            (ik(1), obs("https://example.com/a", cid(0xAA), 100)),
            (ik(2), obs("https://example.com/a", cid(0xAA), 100)),
            (ik(3), obs("https://example.com/c", cid(0xAA), 100)),
        ];
        let a = reconcile_observations(&v, 100, params(2)).unwrap();
        assert_eq!(a.claim.0, "https://example.com/a");
        assert_eq!((a.support, a.dissent), (2, 1));
    }

    /// A tie between two byte versions is refused for the web exactly as it is for retail: an even
    /// split is disagreement, and picking the lexicographically smaller hash would be inventing a
    /// winner out of an implementation detail.
    #[test]
    fn a_tie_between_two_byte_versions_is_refused_not_broken_by_hash_order() {
        let v = vec![
            (ik(1), obs(URL, cid(0x01), 100)),
            (ik(2), obs(URL, cid(0x01), 100)),
            (ik(3), obs(URL, cid(0xFF), 100)),
            (ik(4), obs(URL, cid(0xFF), 100)),
        ];
        assert!(reconcile_observations(&v, 100, params(2)).is_none());

        // One more vote breaks the tie and the same input now decides.
        let mut v = v;
        v.push((ik(5), obs(URL, cid(0x01), 100)));
        assert_eq!(reconcile_observations(&v, 100, params(2)).unwrap().claim.1, cid(0x01));
    }

    /// Freshness applies to the web too: a page corroborated last year is not corroborated now.
    #[test]
    fn stale_web_observations_fall_out_of_the_window() {
        let v = vec![
            (ik(1), obs(URL, cid(0xAA), 1_400)),
            (ik(2), obs(URL, cid(0xAA), 1_400)),
            (ik(3), obs(URL, cid(0xAA), 1_400)),
        ];
        assert!(reconcile_observations(&v, 5_000, params(3)).is_some(), "exactly at the edge");
        assert!(reconcile_observations(&v, 5_001, params(3)).is_none(), "one second past it");
    }

    /// A re-crawl replaces that observer's earlier reading rather than adding a second vote — the
    /// Sybil floor along the time axis, now covering the web.
    #[test]
    fn an_observers_recrawl_supersedes_its_earlier_reading() {
        let v = vec![
            (ik(1), obs(URL, cid(0x11), 100)), // saw the old bytes
            (ik(1), obs(URL, cid(0xAA), 200)), // re-crawled, sees what everyone else sees
            (ik(2), obs(URL, cid(0xAA), 200)),
            (ik(3), obs(URL, cid(0xAA), 200)),
        ];
        let a = reconcile_observations(&v, 200, params(3)).unwrap();
        assert_eq!(a.claim.1, cid(0xAA));
        assert_eq!((a.support, a.dissent), (3, 0), "observer 1 votes once, with its latest reading");
    }

    /// Reconciliation is a pure function of the observation *set* — arrival order cannot change it.
    #[test]
    fn the_web_result_does_not_depend_on_observation_order() {
        let mut v = vec![
            (ik(1), obs(URL, cid(0xAA), 100)),
            (ik(2), obs(URL, cid(0xAA), 100)),
            (ik(3), obs(URL, cid(0xAA), 100)),
            (ik(4), obs(URL, cid(0xBB), 100)),
        ];
        let expected = reconcile_observations(&v, 100, params(3)).unwrap();
        v.reverse();
        assert_eq!(reconcile_observations(&v, 100, params(3)).unwrap(), expected);
        v.swap(0, 2);
        assert_eq!(reconcile_observations(&v, 100, params(3)).unwrap(), expected);
    }

    /// The observation an extractor produces is derived from its own `WebDoc`, so the thing voted
    /// on is the thing indexed — they cannot drift apart.
    #[test]
    fn an_observation_is_derived_from_the_document_it_corroborates() {
        let doc = WebDoc {
            url: URL.into(),
            title: Some("T".into()),
            snippet: "s".into(),
            chunks: vec![Chunk { ordinal: 0, text: "hello".into() }],
            links: vec![],
            discovered: vec![],
            source_hash: cid(0xAA),
        };
        let o = WebObservation::from_doc(&doc, 100);
        assert_eq!(o.url, doc.url);
        assert_eq!(o.source_hash, doc.source_hash);
        assert!(o.is_admissible());

        // An unanchored document yields an inadmissible observation — it cannot vote at all.
        let unanchored = WebDoc { source_hash: ContentId::ZERO, ..doc };
        assert!(!WebObservation::from_doc(&unanchored, 100).is_admissible());
    }

    /// The widening must not have changed retail. Driving `RetailObservation` through the generic
    /// reconciler directly and through the retail wrapper must give the same answer, field for
    /// field — the wrapper is a projection, not a second implementation.
    #[test]
    fn the_generic_reconciler_and_the_retail_wrapper_agree_exactly() {
        use crate::extract::{Availability, RetailMethod, RetailObservation};

        let mk = |av, qty, at| RetailObservation {
            store: "shop.example".into(),
            sku: "SKU1".into(),
            availability: av,
            quantity: qty,
            price_minor: None,
            currency: None,
            method: RetailMethod::StructuredData,
            observed_at: at,
        };
        let v = vec![
            (ik(1), mk(Availability::InStock, Some(10), 100)),
            (ik(2), mk(Availability::InStock, Some(11), 100)),
            (ik(3), mk(Availability::InStock, Some(9), 100)),
            (ik(4), mk(Availability::OutOfStock, Some(0), 100)),
        ];
        let generic = reconcile_observations(&v, 100, params(3)).unwrap();
        let wrapped = reconcile(&v, 100, params(3)).unwrap();

        assert_eq!(wrapped.availability, generic.claim);
        assert_eq!(wrapped.quantity, generic.quantity);
        assert_eq!(wrapped.support, generic.support);
        assert_eq!(wrapped.dissent, generic.dissent);
        assert_eq!(wrapped.availability, Availability::InStock);
        assert_eq!(wrapped.quantity, Some(10));
    }
}
