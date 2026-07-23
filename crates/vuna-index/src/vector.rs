//! Per-space vector index: one of these per served [`SpaceId`], holding one point per document
//! (a document's chunk vectors are averaged into a single doc-level point — good enough for v0
//! nearest-neighbor recall; per-chunk retrieval can split this later without touching the trait).
//!
//! Storage honors the space's [`Quant`]: values are quantized to int8/binary on the way in and
//! dequantized back to f32 only transiently, to feed the ANN distance function. `instant-distance`
//! is a pure-Rust, dependency-light HNSW — no BLAS/C toolchain, so the workspace still builds and
//! tests offline.

use std::collections::HashMap;

use instant_distance::{Builder, Point as AnnPoint, Search};
use vuna_core::space::{EmbeddingSpace, Quant, Vector};
use vuna_core::ContentId;

/// A point as `instant-distance` sees it: a plain f32 vector compared by cosine distance (the
/// hashing embedder — and most sentence embedders — produce roughly unit-norm vectors, so cosine
/// is the natural default).
#[derive(Clone, Debug)]
struct HPoint(Vec<f32>);

impl AnnPoint for HPoint {
    fn distance(&self, other: &Self) -> f32 {
        cosine_distance(&self.0, &other.0)
    }
}

fn cosine_distance(a: &[f32], b: &[f32]) -> f32 {
    let n = a.len().min(b.len());
    let (mut dot, mut na, mut nb) = (0f32, 0f32, 0f32);
    for i in 0..n {
        dot += a[i] * b[i];
        na += a[i] * a[i];
        nb += b[i] * b[i];
    }
    if na <= f32::EPSILON || nb <= f32::EPSILON {
        return 1.0; // a zero vector shares no direction with anything — treat as maximally distant
    }
    (1.0 - dot / (na.sqrt() * nb.sqrt())).clamp(0.0, 2.0)
}

/// A vector as actually held in memory, honoring the space's [`Quant`]. `to_f32` dequantizes for
/// the ANN pass; the point of storing quantized is the memory saving `EmbeddingSpace::vector_bytes`
/// promises, not search-time precision.
#[derive(Clone, Debug)]
enum StoredVector {
    F32(Vec<f32>),
    Int8 { values: Vec<i8>, scale: f32 },
    /// Sign-bit only, packed 8-per-byte. Cheapest, lossiest — distance after dequant degrades to
    /// a Hamming-like signal since only direction-per-dimension survives.
    Binary { bits: Vec<u8>, dim: usize },
}

impl StoredVector {
    fn quantize(values: &[f32], quant: Quant) -> Self {
        match quant {
            Quant::F32 => StoredVector::F32(values.to_vec()),
            Quant::Int8 => {
                let max_abs = values.iter().fold(0f32, |m, &v| m.max(v.abs()));
                let scale = if max_abs > f32::EPSILON { max_abs / 127.0 } else { 1.0 };
                let packed = values.iter().map(|&v| (v / scale).round().clamp(-127.0, 127.0) as i8).collect();
                StoredVector::Int8 { values: packed, scale }
            }
            Quant::Binary => {
                let dim = values.len();
                let mut bits = vec![0u8; dim.div_ceil(8)];
                for (i, &v) in values.iter().enumerate() {
                    if v >= 0.0 {
                        bits[i / 8] |= 1 << (i % 8);
                    }
                }
                StoredVector::Binary { bits, dim }
            }
        }
    }

    fn to_f32(&self) -> Vec<f32> {
        match self {
            StoredVector::F32(v) => v.clone(),
            StoredVector::Int8 { values, scale } => values.iter().map(|&v| v as f32 * scale).collect(),
            StoredVector::Binary { bits, dim } => {
                (0..*dim).map(|i| if bits[i / 8] & (1 << (i % 8)) != 0 { 1.0 } else { -1.0 }).collect()
            }
        }
    }
}

/// One space's worth of vectors: a flat store (upsert-by-[`ContentId`]) plus an ANN index rebuilt
/// on demand at search time. Rebuild-on-search is the deliberately simple v0 choice — fine at
/// shard scale, and it sidesteps `instant-distance`'s batch-build-only API needing incremental
/// insert support it doesn't have.
pub struct VectorSpaceIndex {
    space: EmbeddingSpace,
    entries: Vec<(ContentId, StoredVector)>,
    by_id: HashMap<ContentId, usize>,
}

impl VectorSpaceIndex {
    pub fn new(space: EmbeddingSpace) -> Self {
        Self { space, entries: Vec::new(), by_id: HashMap::new() }
    }

    pub fn space(&self) -> &EmbeddingSpace {
        &self.space
    }

    pub fn is_empty(&self) -> bool {
        self.entries.is_empty()
    }

    /// Averages `vectors` into this doc's single point and upserts by `url_id`. A document's
    /// chunks are retained separately in `IndexedDoc` for re-embedding; here we only need the
    /// doc-level centroid for nearest-neighbor recall.
    pub fn upsert(&mut self, url_id: ContentId, vectors: &[Vector]) -> vuna_core::Result<()> {
        if vectors.is_empty() {
            return Ok(());
        }
        let dim = if self.space.dim > 0 { self.space.dim } else { vectors[0].values.len() };
        let mut acc = vec![0f32; dim];
        for v in vectors {
            if v.values.len() != dim {
                return Err(vuna_core::Error::Index(format!(
                    "vector dim {} does not match space {} dim {dim}",
                    v.values.len(),
                    self.space.id
                )));
            }
            for (a, b) in acc.iter_mut().zip(&v.values) {
                *a += b;
            }
        }
        let n = vectors.len() as f32;
        for a in acc.iter_mut() {
            *a /= n;
        }
        let stored = StoredVector::quantize(&acc, self.space.quant);

        match self.by_id.get(&url_id) {
            Some(&i) => self.entries[i] = (url_id, stored),
            None => {
                self.by_id.insert(url_id, self.entries.len());
                self.entries.push((url_id, stored));
            }
        }
        Ok(())
    }

    /// Nearest neighbors of `query` (already-embedded, same space), nearest first. Score is a
    /// `1/(1+distance)` similarity in `(0, 1]` so it composes simply with BM25 in the blend step.
    pub fn search(&self, query: &Vector, k: usize) -> Vec<(ContentId, f32)> {
        if self.entries.is_empty() || k == 0 {
            return Vec::new();
        }
        let points: Vec<HPoint> = self.entries.iter().map(|(_, sv)| HPoint(sv.to_f32())).collect();
        let ids: Vec<ContentId> = self.entries.iter().map(|(id, _)| *id).collect();
        let map = Builder::default().build(points, ids);

        let mut search = Search::default();
        let qpoint = HPoint(query.values.clone());
        map.search(&qpoint, &mut search)
            .take(k)
            .map(|item| (*item.value, 1.0 / (1.0 + item.distance)))
            .collect()
    }
}
