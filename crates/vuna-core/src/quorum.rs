//! Quorum reconciliation for the retail vertical. The observed store is a **non-participant** — it
//! signs nothing — so unlike the web vertical there is no authoritative feed to defer to. Consensus
//! of K independent, anchored observers is the only ground truth. This module is pure, deterministic
//! logic (no deps) and is fully unit-tested.

use crate::{extract::{Availability, RetailObservation}, IdentityKey, UnixSecs};

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

/// Reconcile observations for a single (store, sku) at time `now`.
///
/// - Only observations within `window_secs` count.
/// - Each **distinct observer** votes once (its latest in-window observation) — this is the Sybil
///   floor: `k` must be distinct `IdentityKey`s, so ballot-stuffing from one key can't reach quorum.
///   (Anchoring those keys — proof-of-personhood / stake — is the caller's job via KOTVA ATTEST.)
/// - Availability is decided by plurality; a value is only *accepted* if its support ≥ `k`, and a
///   **tie for the plurality is not a decision** — see below.
/// - Quantity is the median of agreeing observers' counts, accepted only if ≥ `k` of them fall
///   within `qty_tolerance` of that median. With an even number of counts the **upper** median is
///   taken, so the value returned is always one an observer actually reported rather than an
///   average of two that nobody saw.
///
/// Returns `None` when the network does not (yet) know, which the caller MUST surface rather than
/// fabricate a value. That is the answer in three cases:
///
/// 1. fewer than `k` distinct observers reported inside the window;
/// 2. no single availability value reached `k` distinct backers;
/// 3. two or more values tied for the most support. A tie is disagreement, not a majority, and
///    breaking it — by enum order, insertion order, or a "conservative" preference for
///    out-of-stock — would be inventing a winner the observations do not contain.
///
/// A `None` availability does **not** mean the network says "unavailable"; [`Availability::Unknown`]
/// is a value observers can themselves report and reach quorum on. The two are deliberately
/// distinct: "nobody agrees" and "everybody agrees they can't tell" are different facts.
pub fn reconcile(
    observers: &[(IdentityKey, RetailObservation)],
    now: UnixSecs,
    params: QuorumParams,
) -> Option<Agreed> {
    // 1. window filter + one latest vote per distinct observer.
    let mut latest: std::collections::BTreeMap<IdentityKey, &RetailObservation> = Default::default();
    for (ik, obs) in observers {
        if now.saturating_sub(obs.observed_at) > params.window_secs {
            continue;
        }
        latest
            .entry(*ik)
            .and_modify(|cur| {
                if obs.observed_at > cur.observed_at {
                    *cur = obs;
                }
            })
            .or_insert(obs);
    }
    if latest.len() < params.k {
        return None;
    }

    // 2. availability plurality across distinct observers.
    let mut tally: std::collections::BTreeMap<u8, usize> = Default::default();
    for obs in latest.values() {
        *tally.entry(avail_code(obs.availability)).or_default() += 1;
    }
    let (win_code, support) = tally.iter().map(|(c, n)| (*c, *n)).max_by_key(|(_, n)| *n)?;
    if support < params.k {
        return None;
    }
    // A tie is not a plurality. Whichever value an ordering happened to favour would be an
    // artifact of the implementation, not something the observers said.
    if tally.values().filter(|n| **n == support).count() > 1 {
        return None;
    }
    let availability = code_avail(win_code);
    let dissent = latest.len() - support;

    // 3. quantity: median over observers who backed the winning availability AND reported a count.
    let mut qtys: Vec<u64> = latest
        .values()
        .filter(|o| avail_code(o.availability) == win_code)
        .filter_map(|o| o.quantity)
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

    Some(Agreed { availability, quantity, support, dissent })
}

fn avail_code(a: Availability) -> u8 {
    match a {
        Availability::InStock => 0,
        Availability::OutOfStock => 1,
        Availability::LowStock => 2,
        Availability::Unknown => 3,
    }
}
fn code_avail(c: u8) -> Availability {
    match c {
        0 => Availability::InStock,
        1 => Availability::OutOfStock,
        2 => Availability::LowStock,
        _ => Availability::Unknown,
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
