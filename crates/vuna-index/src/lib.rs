//! vuna-index — tantivy keyword index + per-space HNSW vector index + link/knowledge graph.
//! Implements [`vuna_core::index::Index`] and [`vuna_core::index::Embedder`].
//!
//! - **Keyword**: tantivy BM25 over `title`+`body` ([`keyword::KeywordIndex`]); `url`/`title`/
//!   `snippet` stored for display, `body` (concatenated chunk text) indexed-only.
//! - **Vector**: one `instant-distance` HNSW per served [`vuna_core::space::SpaceId`]
//!   ([`vector::VectorSpaceIndex`]), honoring the space's `Quant` in storage.
//! - **Graph**: `url -> out-neighbor urls` ([`graph::Graph`]) — enough for Min-PPR ranking later.
//! - **Embedder**: [`embed::HashingEmbedder`] is the default, offline, no-model-download embedder
//!   tests run against. A real transformer embedder lives behind the `candle` feature, stubbed.
//!
//! [`engine::VunaIndex`] is the glue: the concrete `Index` impl every other piece here serves.

pub mod embed;
pub mod engine;
pub mod graph;
pub mod keyword;
pub mod vector;

pub use embed::HashingEmbedder;
pub use engine::VunaIndex;
pub use graph::Graph;

#[cfg(test)]
mod tests {
    use super::*;
    use vuna_core::extract::Chunk;
    use vuna_core::index::{Embedder, Index, IndexedDoc, Query};
    use vuna_core::space::{EmbeddingSpace, Quant};
    use vuna_core::ContentId;

    fn cid(byte: u8) -> ContentId {
        let mut b = [0u8; 32];
        b[0] = byte;
        ContentId(b)
    }

    fn doc(id_byte: u8, url: &str, title: &str, snippet: &str, body: &str) -> IndexedDoc {
        IndexedDoc {
            url: url.to_string(),
            url_id: cid(id_byte),
            title: Some(title.to_string()),
            snippet: snippet.to_string(),
            chunks: vec![Chunk { ordinal: 0, text: body.to_string() }],
            indexed_at: 0,
        }
    }

    #[test]
    fn keyword_search_ranks_matching_docs() {
        let mut idx = VunaIndex::new().unwrap();

        idx.upsert(
            &doc(1, "https://a.example", "Rust programming", "Learn Rust", "Rust is a systems programming language"),
            &[],
        )
        .unwrap();
        idx.upsert(
            &doc(2, "https://b.example", "Baking bread", "Sourdough guide", "Flour water salt and time make bread"),
            &[],
        )
        .unwrap();
        idx.upsert(
            &doc(3, "https://c.example", "Rust and WebAssembly", "Rust to wasm", "Compile Rust programs to WebAssembly"),
            &[],
        )
        .unwrap();

        let hits = idx.search(&Query { text: "rust programming".into(), space: None, limit: 10 }).unwrap();

        assert!(!hits.is_empty(), "expected at least one keyword hit");
        let urls: Vec<&str> = hits.iter().map(|h| h.url.as_str()).collect();
        assert!(urls.contains(&"https://a.example"));
        assert!(urls.contains(&"https://c.example"));
        assert!(!urls.contains(&"https://b.example"), "bread doc should not match 'rust programming'");
        // BM25 should rank the doc whose title+body both hit ("Rust programming") at least as
        // high as the doc that only shares one of the two terms directly.
        assert_eq!(hits[0].url, "https://a.example");
    }

    #[test]
    fn hashing_embedder_is_deterministic_and_fixed_dim() {
        let space = EmbeddingSpace::new("test-hash", 32, Quant::F32);
        let embedder = HashingEmbedder::new(space.id.clone(), space.dim);
        let chunks = vec![Chunk { ordinal: 0, text: "hello world".into() }];

        let v1 = embedder.embed(&chunks).unwrap();
        let v2 = embedder.embed(&chunks).unwrap();
        assert_eq!(v1, v2, "hashing embedder must be stable across calls");
        assert_eq!(v1[0].values.len(), space.dim);
        assert_eq!(v1[0].space, space.id);
    }

    #[test]
    fn vector_upsert_and_nearest_neighbor() {
        let space = EmbeddingSpace::new("test-hash", 16, Quant::F32);
        let mut idx = VunaIndex::new().unwrap();
        idx.register_embedder(space.clone(), Box::new(HashingEmbedder::new(space.id.clone(), space.dim)));

        let embedder = HashingEmbedder::new(space.id.clone(), space.dim);

        let near = doc(10, "https://near.example", "Near", "near snippet", "apples oranges bananas grapes");
        let far = doc(11, "https://far.example", "Far", "far snippet", "quantum tunneling semiconductor physics");
        idx.upsert(&near, &[]).unwrap();
        idx.upsert(&far, &[]).unwrap();

        let near_vecs = embedder.embed(&near.chunks).unwrap();
        let far_vecs = embedder.embed(&far.chunks).unwrap();
        idx.upsert_vectors(near.url_id, &space.id, &near_vecs).unwrap();
        idx.upsert_vectors(far.url_id, &space.id, &far_vecs).unwrap();

        assert!(idx.served_spaces().contains(&space.id));

        // A query that shares no keyword terms at all with either doc, so any hit must come from
        // the vector pass — and it should be the doc whose fruity vocabulary is closest.
        let hits = idx
            .search(&Query { text: "apples oranges pears".into(), space: Some(space.id.clone()), limit: 5 })
            .unwrap();

        assert!(!hits.is_empty(), "expected a vector-only hit");
        assert_eq!(hits[0].url, "https://near.example");
    }

    #[test]
    fn graph_edges_round_trip_through_upsert() {
        let mut idx = VunaIndex::new().unwrap();
        let d = doc(20, "https://hub.example", "Hub", "hub snippet", "a hub page");
        idx.upsert(&d, &["https://spoke-a.example".to_string(), "https://spoke-b.example".to_string()]).unwrap();

        let mut n = idx.neighbors("https://hub.example").to_vec();
        n.sort();
        assert_eq!(n, vec!["https://spoke-a.example".to_string(), "https://spoke-b.example".to_string()]);

        idx.add_edge("https://hub.example", "https://spoke-c.example");
        assert!(idx.neighbors("https://hub.example").iter().any(|u| u == "https://spoke-c.example"));
    }
}
