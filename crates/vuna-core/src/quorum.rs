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
/// - Availability is decided by plurality; a value is only *accepted* if its support ≥ `k`.
/// - Quantity is the median of agreeing observers' counts, accepted only if ≥ `k` of them fall
///   within `qty_tolerance` of that median.
///
/// Returns `None` when no availability reaches `k` distinct backers — i.e. the network does not
/// (yet) know, which the caller MUST surface rather than fabricate a value.
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
    let (win_code, &support) = tally.iter().max_by_key(|(_, n)| **n)?;
    if support < params.k {
        return None;
    }
    let availability = code_avail(*win_code);
    let dissent = latest.len() - support;

    // 3. quantity: median over observers who backed the winning availability AND reported a count.
    let mut qtys: Vec<u64> = latest
        .values()
        .filter(|o| avail_code(o.availability) == *win_code)
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
}
