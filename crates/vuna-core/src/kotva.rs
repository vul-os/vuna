//! The KOTVA substrate seam. Vuna rides KOTVA identity / PUB / DHT — it invents no crypto. Core
//! declares only these traits; `vuna-node` implements them against `kotva-core` (pinned by tag),
//! so the rest of the workspace stays substrate-agnostic and compiles offline.

use crate::{ContentId, IdentityKey};

/// This node's identity — an Ed25519 keypair, per KOTVA §1. The private key never leaves the impl.
pub trait NodeIdentity: Send + Sync {
    fn public(&self) -> IdentityKey;
    /// Detached signature over `msg` (DS-tagged per the KOTVA binding).
    fn sign(&self, msg: &[u8]) -> Vec<u8>;
}

/// Publish a signed public object (a KOTVA `PubAnnounce` on this node's author feed) and content
/// address it. Vuna objects — URL lists, space announcements, node descriptors — publish through
/// this. Returns the content address the object is fetched by.
pub trait Publisher: Send + Sync {
    fn publish(&self, canonical_bytes: &[u8]) -> crate::Result<ContentId>;
    /// Verify a signed object's author + integrity. Returns the author IK on success.
    fn verify(&self, canonical_bytes: &[u8], sig: &[u8], author: IdentityKey) -> crate::Result<()>;
}

/// Content-address canonical bytes (BLAKE3-256), matching KOTVA §2.2. Behind the seam so core pulls
/// in no crypto; `vuna-node` supplies the real hasher from `kotva-core`.
pub trait ContentAddresser: Send + Sync {
    fn address(&self, bytes: &[u8]) -> ContentId;
}
