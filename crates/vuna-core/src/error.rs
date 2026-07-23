//! One error type for the contract. Downstream crates may wrap their own on top.

use thiserror::Error;

pub type Result<T> = std::result::Result<T, Error>;

#[derive(Error, Debug)]
pub enum Error {
    #[error("fetch failed for {url}: {reason}")]
    Fetch { url: String, reason: String },

    #[error("extractor {kind} could not parse {url}: {reason}")]
    Extract { kind: String, url: String, reason: String },

    #[error("no extractor registered for kind {0}")]
    UnknownExtractor(String),

    #[error("embedding space {0} is not served by this node")]
    UnservedSpace(String),

    #[error("index error: {0}")]
    Index(String),

    #[error("frontier error: {0}")]
    Frontier(String),

    #[error("substrate/kotva error: {0}")]
    Kotva(String),

    #[error("serialization: {0}")]
    Serde(#[from] serde_json::Error),

    #[error("{0}")]
    Other(String),
}
