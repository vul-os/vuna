//! Declarative retail adapters — the long tail of storefronts as **data**, not Rust.
//!
//! `adapters/*.toml` describes how one site or store platform publishes price/stock signal; this
//! module is the interpreter that reads such a manifest and runs it as a
//! [`vuna_core::extract::Extractor`]. Adding the hundredth storefront platform is a reviewed data
//! change, not a compiled release — the same shape as the frontier's `UrlList`, which is also a
//! signed object rather than a code change.
//!
//! ```no_run
//! use std::path::Path;
//! use vuna_extract::{AdapterExtractor, ExtractorRegistry};
//!
//! let mut registry = ExtractorRegistry::with_defaults();
//! for adapter in AdapterExtractor::load_dir(Path::new("adapters"))? {
//!     registry.register(Box::new(adapter));
//! }
//! # Ok::<(), vuna_core::Error>(())
//! ```
//!
//! The directory is **opt-in**: [`ExtractorRegistry::with_defaults`](crate::ExtractorRegistry::with_defaults)
//! never touches the filesystem, and every crate here builds, tests, and runs with `adapters/`
//! absent. A manifest can only ever add coverage; it can never be a build dependency.
//!
//! What a manifest can and cannot express today is in `adapters/README.md`; the boundaries the
//! interpreter enforces (no derived fetches, no secondary lookups, no multi-offer aggregation) are
//! load-time errors, not silent no-ops — see [`ManifestError`].

mod interp;
mod manifest;
mod pattern;
mod price;

pub use interp::{AdapterExtractor, MAX_OBSERVATIONS};
pub use manifest::{
    AdapterManifest, AdapterMeta, BodyFormat, Expr, FieldMap, FetchSpec, IterateSpec, ManifestError, MatchRules,
    FORMAT_VERSION, IMPLICIT_CAPTURES, SUPPORTED_KIND,
};
pub use price::{PriceRepresentation, DEFAULT_EXPONENT};
