//! Query fan-out + merge across the federation — the KOTVA SEARCH read path. A client queries the
//! local shard first (works with zero coordinator), and MAY fan out to peers / an opt-in `indexer`
//! for reach. No index is presented as complete against the whole network (SRCH-7).

use crate::index::{Hit, Query};
use serde::{Deserialize, Serialize};

/// Where results came from — surfaced to the user, never hidden (SRCH-7/SRCH-9).
#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum Source {
    /// The querying node's own shard — always available, offline-safe.
    Local,
    /// A peer reached over the mesh/DHT.
    Peer,
    /// An opt-in global `indexer` coordinator (adds reach, never authority).
    Indexer,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct RankedHit {
    pub hit: Hit,
    pub source: Source,
    /// Min-PPR / web-of-trust rank contribution, disclosed as policy (SRCH-4).
    pub rank: f32,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Answer {
    pub hits: Vec<RankedHit>,
    /// True only if every source that should have answered did — else results are partial and the
    /// UI MUST say so rather than imply completeness.
    pub complete_over_reachable: bool,
}

/// The read engine. `vuna-query` implements local-first search + optional fan-out + Min-PPR merge.
pub trait QueryEngine {
    fn search(&self, q: &Query) -> crate::Result<Answer>;
}
