//! # vuna-core — the frozen contract
//!
//! Vuna is a decentralized **crawl → extract → index/graph** engine on the KOTVA substrate. The
//! network stores the **index, not the pages**: keyword postings, vector embeddings (one set per
//! embedding *space*), a link/knowledge graph, and a pointer back to the live URL.
//!
//! This crate is the **contract** every other crate builds against: the shared data types and the
//! traits at the seams between stages. It deliberately has **no heavy dependencies** (no tantivy,
//! libp2p, reqwest, or crypto) so it always compiles offline and never churns. Implementations
//! live downstream:
//!
//! | Stage | Crate | Key trait(s) here |
//! |-------|-------|-------------------|
//! | distributed URL lists | `vuna-frontier` | [`frontier::Frontier`] |
//! | polite fetch | `vuna-crawl` | `Fetcher` (defined there — core has no fetch trait) |
//! | pluggable extractors | `vuna-extract` | [`extract::Extractor`] |
//! | keyword + vector + graph index | `vuna-index` | [`index::Index`], [`index::Embedder`] |
//! | query fan-out + ranking | `vuna-query` | [`query::QueryEngine`] |
//! | substrate binding | `vuna-node` | [`kotva::NodeIdentity`], [`kotva::Publisher`] |
//!
//! ## The one idea that makes it future-proof
//! **Embedding spaces and extractor kinds are the same opt-in plurality.** A node declares which
//! [`space::EmbeddingSpace`]s and which [`extract::ExtractorKind`]s it serves
//! ([`node::NodeDescriptor`]). New model or new vertical → new space/kind, announced, adopted by
//! opt-in — never a fleet-wide flag-day. Because shards keep chunk *text*, re-embedding into a new
//! space is local recompute, not a re-crawl.

pub mod error;
pub mod frontier;
pub mod space;
pub mod extract;
pub mod index;
pub mod query;
pub mod node;
pub mod kotva;
pub mod quorum;

pub use error::{Error, Result};

/// Unix seconds. Vuna never *generates* time in core logic — callers pass it in, so logic stays
/// deterministic and testable (and replay-safe on the substrate).
pub type UnixSecs = u64;

/// A content address: `BLAKE3-256` of canonical bytes, as computed by the KOTVA binding. Core
/// treats it as an opaque 32-byte identifier — it never hashes, so it pulls in no crypto dep.
#[derive(Clone, Copy, PartialEq, Eq, Hash, PartialOrd, Ord, serde::Serialize, serde::Deserialize)]
pub struct ContentId(pub [u8; 32]);

impl ContentId {
    pub const ZERO: ContentId = ContentId([0u8; 32]);
    pub fn to_hex(&self) -> String {
        let mut s = String::with_capacity(64);
        for b in self.0 {
            s.push(char::from_digit((b >> 4) as u32, 16).unwrap());
            s.push(char::from_digit((b & 0xf) as u32, 16).unwrap());
        }
        s
    }
}

impl std::fmt::Debug for ContentId {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "ContentId({}…)", &self.to_hex()[..12])
    }
}

/// A public identity key (Ed25519 public key bytes). The author of any signed Vuna object. The
/// keypair itself lives behind the [`kotva::NodeIdentity`] seam, never in core.
#[derive(Clone, Copy, PartialEq, Eq, Hash, PartialOrd, Ord, serde::Serialize, serde::Deserialize)]
pub struct IdentityKey(pub [u8; 32]);

impl std::fmt::Debug for IdentityKey {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let mut s = String::with_capacity(16);
        for b in &self.0[..6] {
            s.push(char::from_digit((b >> 4) as u32, 16).unwrap());
            s.push(char::from_digit((b & 0xf) as u32, 16).unwrap());
        }
        write!(f, "IK({s}…)")
    }
}
