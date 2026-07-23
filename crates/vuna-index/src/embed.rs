//! [`Embedder`] implementations. The default, [`HashingEmbedder`], is a deterministic bag-of-words
//! hashing embedder: no model, no download, no randomness — just a fixed hash function — so the
//! whole workspace tests offline. A real transformer embedder (candle/ONNX) lives behind the
//! `candle` feature as a documented stub; nothing pulls in a model runtime by default.

use vuna_core::extract::Chunk;
use vuna_core::space::{SpaceId, Vector};
use vuna_core::Result;

/// The frozen contract trait this module implements.
use vuna_core::index::Embedder;

/// Deterministic hashing ("feature hashing") bag-of-words embedder. Tokenizes on non-alphanumeric
/// boundaries, hashes each token with FNV-1a, and accumulates a signed +/-1 into `hash % dim` —
/// the standard trick for a fixed-dimension embedding with no vocabulary to train or ship. Same
/// text always yields the same vector, and the vector is always L2-normalized.
pub struct HashingEmbedder {
    space: SpaceId,
    dim: usize,
}

impl HashingEmbedder {
    pub fn new(space: impl Into<SpaceId>, dim: usize) -> Self {
        Self { space: space.into(), dim: dim.max(1) }
    }
}

impl Embedder for HashingEmbedder {
    fn space(&self) -> &SpaceId {
        &self.space
    }

    fn embed(&self, chunks: &[Chunk]) -> Result<Vec<Vector>> {
        Ok(chunks
            .iter()
            .map(|c| Vector { space: self.space.clone(), values: hash_embed(&c.text, self.dim) })
            .collect())
    }
}

fn hash_embed(text: &str, dim: usize) -> Vec<f32> {
    let mut v = vec![0f32; dim];
    for tok in tokenize(text) {
        let h = fnv1a(tok.as_bytes());
        let idx = (h % dim as u64) as usize;
        // Use a second, independent-ish bit of the same hash for the sign — the standard
        // feature-hashing trick to keep collisions unbiased in expectation.
        let sign = if (h >> 63) & 1 == 1 { 1.0 } else { -1.0 };
        v[idx] += sign;
    }
    l2_normalize(&mut v);
    v
}

fn tokenize(text: &str) -> impl Iterator<Item = String> + '_ {
    text.split(|c: char| !c.is_alphanumeric()).filter(|s| !s.is_empty()).map(|s| s.to_lowercase())
}

/// FNV-1a — tiny, dependency-free, and (unlike `std`'s `RandomState`-backed hashers reached via a
/// `HashMap`) deterministic across runs, which is the whole point of this embedder.
fn fnv1a(bytes: &[u8]) -> u64 {
    let mut hash: u64 = 0xcbf29ce484222325;
    for &b in bytes {
        hash ^= b as u64;
        hash = hash.wrapping_mul(0x0000_0100_0000_01b3);
    }
    hash
}

fn l2_normalize(v: &mut [f32]) {
    let norm = v.iter().map(|x| x * x).sum::<f32>().sqrt();
    if norm > f32::EPSILON {
        for x in v.iter_mut() {
            *x /= norm;
        }
    }
}

/// Real transformer embedder — deliberately unimplemented. Wiring in an actual model (candle +
/// ONNX/safetensors weights) is a separate, deliberate step so the default build never needs a
/// model download to compile or test. Behind `--features candle`.
#[cfg(feature = "candle")]
pub mod candle {
    use super::*;

    /// TODO(agent): load a real sentence-embedding model (e.g. bge-small) via candle and produce
    /// real vectors in `embed`. Until then this space is declared but not servable.
    pub struct CandleEmbedder {
        space: SpaceId,
        dim: usize,
    }

    impl CandleEmbedder {
        pub fn new(space: impl Into<SpaceId>, dim: usize) -> Self {
            Self { space: space.into(), dim }
        }
    }

    impl Embedder for CandleEmbedder {
        fn space(&self) -> &SpaceId {
            &self.space
        }

        fn embed(&self, _chunks: &[Chunk]) -> Result<Vec<Vector>> {
            Err(vuna_core::Error::Other(format!(
                "candle embedder for space {} (dim {}) is a stub — TODO: load model + real inference",
                self.space, self.dim
            )))
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn stable_and_correctly_dimensioned() {
        let e = HashingEmbedder::new("test@8/f32".to_string(), 8);
        let chunk = Chunk { ordinal: 0, text: "the quick brown fox".into() };
        let a = e.embed(std::slice::from_ref(&chunk)).unwrap();
        let b = e.embed(std::slice::from_ref(&chunk)).unwrap();
        assert_eq!(a, b, "same text must hash to the same vector");
        assert_eq!(a[0].values.len(), 8);
    }

    #[test]
    fn different_text_differs() {
        let e = HashingEmbedder::new("test@16/f32".to_string(), 16);
        let a = e.embed(&[Chunk { ordinal: 0, text: "rust programming language".into() }]).unwrap();
        let b = e.embed(&[Chunk { ordinal: 0, text: "banana bread recipe".into() }]).unwrap();
        assert_ne!(a[0].values, b[0].values);
    }
}
