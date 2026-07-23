//! vuna-frontier — distributed URL lists: subscribe, dedup, DHT crawl-assignment.
//!
//! Implements [`vuna_core::frontier::Frontier`] as [`FrontierStore`]: a `redb`-backed local
//! store of [`vuna_core::frontier::UrlEntry`], keyed by [`vuna_core::frontier::UrlId`]
//! (a BLAKE3-256 hash of the URL's canonical form — see [`canon`]).
//!
//! Real DHT (Kademlia) crawl-assignment is not implemented here — this crate never depends on
//! libp2p. Instead, ownership sits behind the [`assignment::Assignment`] trait ([`assignment`]);
//! `vuna-node` drops in the real Kademlia binding as an `Assignment` impl once it exists, and
//! nothing in this crate changes. Two impls ship today: [`assignment::OwnAll`] (dev/single-node)
//! and [`assignment::ModuloAssignment`] (a deterministic stand-in for a fixed-size cluster).
//!
//! K× replication (storing each entry on more than one owning node) is future scope, layered on
//! top of `Assignment` the same way the real DHT binding will be — not needed for a correct v0.

pub mod assignment;
pub mod canon;
pub mod store;

pub use assignment::{Assignment, ModuloAssignment, OwnAll};
pub use canon::{canonical_id, canonicalize, url_id};
pub use store::FrontierStore;
