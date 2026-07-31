//! The index seam: keyword postings + per-space vectors + link graph. Implementations (`tantivy`,
//! an HNSW vector store, a graph store) live in `vuna-index`. The index is **derived, rebuildable,
//! never authoritative** (KOTVA SEARCH SRCH-2): on disagreement, the author's signed content wins.

use crate::{
    extract::{Chunk, WebDoc},
    space::{SpaceId, Vector},
    ContentId, UnixSecs,
};
use serde::{Deserialize, Serialize};

/// Turns chunk text into vectors for a space. The embedding-model binding (a local runtime such as
/// ONNX/candle) implements this in `vuna-index`; core stays model-free.
pub trait Embedder: Send + Sync {
    fn space(&self) -> &SpaceId;
    fn embed(&self, chunks: &[Chunk]) -> crate::Result<Vec<Vector>>;
}

/// A document as indexed — the durable per-URL record. Note: **no raw page**, only derived signal
/// plus a pointer back to the live URL.
///
/// This is the artifact a peer actually contributes to the shared index, so it carries
/// [`source_hash`](Self::source_hash) forward from the extraction. Dropping the anchor here would
/// reopen the same hole one stage later: a replicated `IndexedDoc` with nothing to check against.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct IndexedDoc {
    pub url: String,
    pub url_id: ContentId,
    pub title: Option<String>,
    pub snippet: String,
    /// Retained so re-embedding into a NEW space needs no re-crawl.
    pub chunks: Vec<Chunk>,
    pub indexed_at: UnixSecs,
    /// The [`WebDoc::source_hash`] this record was derived from — carried verbatim, never
    /// recomputed here (there are no bytes left to recompute it from). See [`crate::trust`].
    pub source_hash: ContentId,
}

impl IndexedDoc {
    pub fn from_web(doc: &WebDoc, url_id: ContentId, now: UnixSecs) -> Self {
        Self {
            url: doc.url.clone(),
            url_id,
            title: doc.title.clone(),
            snippet: doc.snippet.clone(),
            chunks: doc.chunks.clone(),
            indexed_at: now,
            source_hash: doc.source_hash,
        }
    }
}

/// A query into the index. `space` selects the vector index; None = keyword-only.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Query {
    pub text: String,
    pub space: Option<SpaceId>,
    pub limit: usize,
}

/// A ranked hit. `score` is index-local, disclosed policy — never a network-wide authority.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Hit {
    pub url: String,
    pub title: Option<String>,
    pub snippet: String,
    pub score: f32,
}

/// The node-local index. One keyword index + graph, plus one vector index per served space.
pub trait Index {
    /// Add/replace a document's keyword + graph contribution (space-agnostic).
    fn upsert(&mut self, doc: &IndexedDoc, links_to: &[String]) -> crate::Result<()>;

    /// Add a document's vectors into a specific space's vector index.
    fn upsert_vectors(&mut self, url_id: ContentId, space: &SpaceId, vectors: &[Vector]) -> crate::Result<()>;

    /// Keyword (BM25) + optional vector search over the local shard.
    fn search(&self, q: &Query) -> crate::Result<Vec<Hit>>;

    /// Which spaces this shard actually holds vectors for.
    fn served_spaces(&self) -> Vec<SpaceId>;
}
