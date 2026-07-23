//! [`VunaIndex`] — the concrete [`vuna_core::index::Index`]: wires the tantivy keyword index
//! ([`crate::keyword::KeywordIndex`]), the per-space HNSW vector stores ([`crate::vector::VectorSpaceIndex`]),
//! and the link graph ([`crate::graph::Graph`]) together behind the one frozen trait.

use std::collections::HashMap;

use vuna_core::extract::Chunk;
use vuna_core::index::{Embedder, Hit, Index, IndexedDoc, Query};
use vuna_core::space::{EmbeddingSpace, Quant, SpaceId, Vector};
use vuna_core::ContentId;

use crate::graph::Graph;
use crate::keyword::KeywordIndex;
use crate::vector::VectorSpaceIndex;

/// Display metadata kept per document so a vector-only hit (a neighbor with no keyword match) can
/// still be rendered as a [`Hit`] without a second round-trip through tantivy's doc store.
#[derive(Clone, Debug)]
struct DocMeta {
    url: String,
    title: Option<String>,
    snippet: String,
}

/// The node-local index: one keyword index, one vector index per served space, one link graph.
/// `Index::search` has no embedder parameter (it's a frozen, model-agnostic trait), so `VunaIndex`
/// owns the embedders it needs to turn a query's *text* into a vector for whichever space the
/// caller asks for — see [`Self::register_embedder`].
pub struct VunaIndex {
    keyword: KeywordIndex,
    vectors: HashMap<SpaceId, VectorSpaceIndex>,
    embedders: HashMap<SpaceId, Box<dyn Embedder>>,
    graph: Graph,
    docs: HashMap<ContentId, DocMeta>,
}

impl VunaIndex {
    pub fn new() -> vuna_core::Result<Self> {
        Ok(Self {
            keyword: KeywordIndex::new()?,
            vectors: HashMap::new(),
            embedders: HashMap::new(),
            graph: Graph::new(),
            docs: HashMap::new(),
        })
    }

    /// Declares a served embedding space up front, mainly so its [`Quant`] is known before any
    /// vectors land (`upsert_vectors` otherwise lazily creates an f32-quant space on first write,
    /// inferring `dim` from the vectors themselves — a reasonable default, just not quantized).
    pub fn register_space(&mut self, space: EmbeddingSpace) {
        self.vectors.entry(space.id.clone()).or_insert_with(|| VectorSpaceIndex::new(space));
    }

    /// Registers the [`Embedder`] used to embed a *query's* text into `space` at search time, and
    /// declares the space (see [`Self::register_space`]) if not already known.
    pub fn register_embedder(&mut self, space: EmbeddingSpace, embedder: Box<dyn Embedder>) {
        self.register_space(space);
        self.embedders.insert(embedder.space().clone(), embedder);
    }

    /// Direct graph access beyond the `Index` trait's surface — `vuna-query`'s Min-PPR ranking
    /// (and tests) need to walk edges directly rather than through a search call.
    pub fn graph(&self) -> &Graph {
        &self.graph
    }

    pub fn add_edge(&mut self, from: &str, to: &str) {
        self.graph.add_edge(from, to);
    }

    pub fn neighbors(&self, from: &str) -> &[String] {
        self.graph.neighbors(from)
    }
}

impl Index for VunaIndex {
    fn upsert(&mut self, doc: &IndexedDoc, links_to: &[String]) -> vuna_core::Result<()> {
        self.keyword.upsert(doc)?;
        self.graph.set_edges(&doc.url, links_to);
        self.docs.insert(
            doc.url_id,
            DocMeta { url: doc.url.clone(), title: doc.title.clone(), snippet: doc.snippet.clone() },
        );
        Ok(())
    }

    fn upsert_vectors(&mut self, url_id: ContentId, space: &SpaceId, vectors: &[Vector]) -> vuna_core::Result<()> {
        let store = self.vectors.entry(space.clone()).or_insert_with(|| {
            let dim = vectors.first().map(|v| v.values.len()).unwrap_or(0);
            VectorSpaceIndex::new(EmbeddingSpace {
                id: space.clone(),
                model_id: String::new(),
                dim,
                quant: Quant::F32,
                default: false,
            })
        });
        store.upsert(url_id, vectors)
    }

    fn search(&self, q: &Query) -> vuna_core::Result<Vec<Hit>> {
        if q.limit == 0 {
            return Ok(Vec::new());
        }

        let mut hits = self.keyword.search(&q.text, q.limit)?;

        if let Some(space) = &q.space {
            if let (Some(store), Some(embedder)) = (self.vectors.get(space), self.embedders.get(space)) {
                if !store.is_empty() {
                    let query_chunk = Chunk { ordinal: 0, text: q.text.clone() };
                    if let Ok(qvecs) = embedder.embed(std::slice::from_ref(&query_chunk)) {
                        if let Some(qv) = qvecs.into_iter().next() {
                            let nn = store.search(&qv, q.limit);
                            hits = blend(hits, nn, &self.docs);
                        }
                    }
                }
            }
        }

        hits.truncate(q.limit);
        Ok(hits)
    }

    fn served_spaces(&self) -> Vec<SpaceId> {
        self.vectors.iter().filter(|(_, v)| !v.is_empty()).map(|(id, _)| id.clone()).collect()
    }
}

/// Keyword-primary blend (per `vuna_core::index`'s doc: "keyword-primary is fine for v0"):
/// existing keyword hits get a small similarity boost when a vector neighbor agrees with them;
/// vector-only neighbors (no keyword match at all) are appended, ranked by similarity alone. This
/// is deliberately simple — real fusion (RRF, learned blending) is `vuna-query`'s job once results
/// are fanned out across shards; this is just the single-shard v0 contribution.
fn blend(mut kw_hits: Vec<Hit>, nn: Vec<(ContentId, f32)>, docs: &HashMap<ContentId, DocMeta>) -> Vec<Hit> {
    const VECTOR_BOOST: f32 = 0.25;

    let index_of: HashMap<String, usize> = kw_hits.iter().enumerate().map(|(i, h)| (h.url.clone(), i)).collect();
    let mut extra: Vec<Hit> = Vec::new();
    let mut seen_extra: HashMap<String, usize> = HashMap::new();

    for (cid, vscore) in nn {
        let Some(meta) = docs.get(&cid) else { continue };
        if let Some(&i) = index_of.get(&meta.url) {
            kw_hits[i].score += VECTOR_BOOST * vscore;
        } else if let Some(&i) = seen_extra.get(&meta.url) {
            extra[i].score = extra[i].score.max(VECTOR_BOOST * vscore);
        } else {
            seen_extra.insert(meta.url.clone(), extra.len());
            extra.push(Hit {
                url: meta.url.clone(),
                title: meta.title.clone(),
                snippet: meta.snippet.clone(),
                score: VECTOR_BOOST * vscore,
            });
        }
    }

    kw_hits.sort_by(|a, b| b.score.partial_cmp(&a.score).unwrap_or(std::cmp::Ordering::Equal));
    extra.sort_by(|a, b| b.score.partial_cmp(&a.score).unwrap_or(std::cmp::Ordering::Equal));
    kw_hits.extend(extra);
    kw_hits
}
