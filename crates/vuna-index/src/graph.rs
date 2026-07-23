//! A minimal link/knowledge graph: `url -> out-neighbor urls`. Keyed by URL (not [`ContentId`])
//! because a link's *target* may not be crawled/indexed yet — the edge should still be recordable.
//! In-memory, with JSON save/load; enough surface for Min-PPR ranking in `vuna-query` later.

use std::collections::HashMap;
use std::path::Path;

use serde::{Deserialize, Serialize};

#[derive(Clone, Debug, Default, Serialize, Deserialize)]
pub struct Graph {
    edges: HashMap<String, Vec<String>>,
}

impl Graph {
    pub fn new() -> Self {
        Self::default()
    }

    /// Replaces `from`'s entire out-edge set. This is what [`Index::upsert`](vuna_core::index::Index::upsert)
    /// calls with a page's freshly-extracted `links_to` — upsert semantics, matching the rest of
    /// the index (a re-crawl's new link set fully replaces the old one).
    pub fn set_edges(&mut self, from: &str, to: &[String]) {
        self.edges.insert(from.to_string(), to.to_vec());
    }

    /// Adds a single edge without disturbing `from`'s other edges — handy for tests and for
    /// callers building the graph incrementally outside the `Index::upsert` path.
    pub fn add_edge(&mut self, from: &str, to: &str) {
        let out = self.edges.entry(from.to_string()).or_default();
        if !out.iter().any(|u| u == to) {
            out.push(to.to_string());
        }
    }

    pub fn neighbors(&self, from: &str) -> &[String] {
        self.edges.get(from).map(Vec::as_slice).unwrap_or(&[])
    }

    pub fn node_count(&self) -> usize {
        self.edges.len()
    }

    pub fn save(&self, path: &Path) -> vuna_core::Result<()> {
        let bytes = serde_json::to_vec_pretty(self)?;
        std::fs::write(path, bytes)
            .map_err(|e| vuna_core::Error::Index(format!("graph save to {}: {e}", path.display())))
    }

    pub fn load(path: &Path) -> vuna_core::Result<Self> {
        let bytes = std::fs::read(path)
            .map_err(|e| vuna_core::Error::Index(format!("graph load from {}: {e}", path.display())))?;
        Ok(serde_json::from_slice(&bytes)?)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn add_edge_and_neighbors() {
        let mut g = Graph::new();
        g.add_edge("https://a.example", "https://b.example");
        g.add_edge("https://a.example", "https://c.example");
        g.add_edge("https://a.example", "https://b.example"); // duplicate, should not double up

        let mut n = g.neighbors("https://a.example").to_vec();
        n.sort();
        assert_eq!(n, vec!["https://b.example".to_string(), "https://c.example".to_string()]);
        assert!(g.neighbors("https://nowhere.example").is_empty());
    }

    #[test]
    fn set_edges_replaces() {
        let mut g = Graph::new();
        g.set_edges("https://a.example", &["https://b.example".to_string()]);
        g.set_edges("https://a.example", &["https://c.example".to_string()]);
        assert_eq!(g.neighbors("https://a.example"), &["https://c.example".to_string()]);
    }

    #[test]
    fn save_and_load_round_trips() {
        let mut g = Graph::new();
        g.add_edge("https://a.example", "https://b.example");
        let path = std::env::temp_dir().join(format!("vuna-index-graph-test-{}.json", std::process::id()));
        g.save(&path).unwrap();
        let loaded = Graph::load(&path).unwrap();
        assert_eq!(loaded.neighbors("https://a.example"), g.neighbors("https://a.example"));
        let _ = std::fs::remove_file(&path);
    }
}
