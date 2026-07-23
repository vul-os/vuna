//! Embedding **spaces**: the multi-model plurality. Each space is an independent, parallel vector
//! index over the same corpus. The keyword index and the graph are model-agnostic and shared, so
//! they never need re-doing when a model changes.

use serde::{Deserialize, Serialize};

/// Stable id of an embedding space — `model_id@dim/quant`, e.g. `bge-small-en-v1.5@384/int8`.
/// Deterministic string so two nodes name the same space identically.
pub type SpaceId = String;

/// Vector quantization — the storage/recall knob. int8 ≈ 4× smaller than f32; binary ≈ 32×.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum Quant {
    F32,
    Int8,
    Binary,
}

impl Quant {
    /// Bytes per dimension (binary is bit-packed → reported as a whole-vector cost elsewhere).
    pub fn bytes_per_dim(&self) -> f32 {
        match self {
            Quant::F32 => 4.0,
            Quant::Int8 => 1.0,
            Quant::Binary => 0.125,
        }
    }
}

/// A recognized embedding space. Announced as a signed PUB object; a small curated set are the
/// defaults a fresh node ships subscribed to.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct EmbeddingSpace {
    pub id: SpaceId,
    /// Model identifier (e.g. a Hugging Face repo id) — the binding that produces the vectors.
    pub model_id: String,
    pub dim: usize,
    pub quant: Quant,
    /// True for the small default set every client serves out of the box.
    #[serde(default)]
    pub default: bool,
}

impl EmbeddingSpace {
    pub fn new(model_id: impl Into<String>, dim: usize, quant: Quant) -> Self {
        let model_id = model_id.into();
        let id = format!(
            "{model_id}@{dim}/{}",
            serde_json::to_value(quant).ok().and_then(|v| v.as_str().map(str::to_owned)).unwrap_or_default()
        );
        Self { id, model_id, dim, quant, default: false }
    }

    /// Estimated stored bytes for one vector in this space (excludes ANN-graph overhead).
    pub fn vector_bytes(&self) -> usize {
        (self.dim as f32 * self.quant.bytes_per_dim()).ceil() as usize
    }
}

/// A dense vector in a given space. Length must equal the space's `dim` (for f32/int8); binary is
/// pre-packed. Stored quantized on disk; f32 here is the in-memory working form.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Vector {
    pub space: SpaceId,
    pub values: Vec<f32>,
}
