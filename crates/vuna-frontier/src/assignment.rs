//! DHT crawl-assignment, behind a trait — so the real Kademlia binding (in `vuna-node`, once it
//! wires libp2p) drops in later without this crate ever depending on libp2p. [`FrontierStore::due`]
//! filters candidates down to entries this node *owns* via whatever [`Assignment`] it was built
//! with.

use vuna_core::frontier::UrlId;

/// Decides whether this node owns (is responsible for crawling) a given [`UrlId`]. A stable,
/// pure function of the id — no lookups, no I/O — so any node can answer "do I own this?" for
/// any id without asking around, the same property Kademlia rendezvous gives.
pub trait Assignment: Send + Sync {
    fn owns(&self, id: &UrlId) -> bool;
}

/// Dev/single-node mode: this node owns everything. The default for a lone node, a test, or a
/// small self-hosted deployment that never shards.
#[derive(Clone, Copy, Debug, Default)]
pub struct OwnAll;

impl Assignment for OwnAll {
    fn owns(&self, _id: &UrlId) -> bool {
        true
    }
}

/// A stand-in for real DHT assignment: a fixed cluster of `node_count` nodes, numbered
/// `0..node_count`. Node `node_index` owns exactly the ids where `hash(id) % node_count ==
/// node_index`. BLAKE3 output is uniformly distributed, so this partitions any id set evenly and
/// deterministically — every id is owned by exactly one node, computable locally. This is the
/// seam the real Kademlia binding replaces: same trait, no rendezvous round-trip needed here.
#[derive(Clone, Copy, Debug)]
pub struct ModuloAssignment {
    pub node_index: u64,
    pub node_count: u64,
}

impl ModuloAssignment {
    pub fn new(node_index: u64, node_count: u64) -> Self {
        assert!(node_count > 0, "node_count must be at least 1");
        assert!(node_index < node_count, "node_index must be < node_count");
        Self { node_index, node_count }
    }
}

impl Assignment for ModuloAssignment {
    fn owns(&self, id: &UrlId) -> bool {
        // Reduce the first 8 bytes of the (already uniform) BLAKE3 hash mod node_count.
        let mut buf = [0u8; 8];
        buf.copy_from_slice(&id.0[..8]);
        u64::from_le_bytes(buf) % self.node_count == self.node_index
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::canon::url_id;

    #[test]
    fn own_all_owns_everything() {
        let id = url_id("https://example.com/x");
        assert!(OwnAll.owns(&id));
    }

    #[test]
    fn modulo_assignment_partitions_with_no_overlap_and_full_coverage() {
        let node_count = 5u64;
        let ids: Vec<UrlId> = (0..500)
            .map(|i| url_id(&format!("https://example.com/page/{i}")))
            .collect();
        let nodes: Vec<ModuloAssignment> =
            (0..node_count).map(|i| ModuloAssignment::new(i, node_count)).collect();

        for id in &ids {
            let owners = nodes.iter().filter(|n| n.owns(id)).count();
            // Full coverage (at least one owner) and no overlap (at most one owner) together mean
            // exactly one.
            assert_eq!(owners, 1, "id {id:?} must have exactly one owner across the cluster");
        }
    }

    #[test]
    #[should_panic]
    fn modulo_assignment_rejects_out_of_range_index() {
        ModuloAssignment::new(3, 3);
    }
}
