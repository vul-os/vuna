//! What a node *is*, on the wire: which lists it subscribes to, which spaces and extractor kinds it
//! serves. This is the opt-in plurality made concrete — a node serving the default space + one list
//! is already a complete, useful participant.

use crate::{extract::ExtractorKind, space::SpaceId, ContentId, IdentityKey, UnixSecs};
use serde::{Deserialize, Serialize};

/// Signed node advertisement (a KOTVA coordinator/indexer descriptor in substrate terms).
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct NodeDescriptor {
    pub node: IdentityKey,
    /// URL lists this node crawls from.
    pub subscribed_lists: Vec<ContentId>,
    /// Embedding spaces this node builds + serves vectors for.
    pub served_spaces: Vec<SpaceId>,
    /// Extractor verticals this node runs (e.g. `web`, `retail`).
    pub extractors: Vec<ExtractorKind>,
    /// Query-channel visibility, disclosed (SRCH-9): can the operator read queries in the clear?
    pub query_visibility: QueryVisibility,
    pub updated_at: UnixSecs,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum QueryVisibility {
    /// Operator can read queries in the clear.
    Terminating,
    /// Queries are shielded from the operator (PIR/DP — a later layer).
    Blind,
}

impl NodeDescriptor {
    /// The out-of-the-box node: default space, `web` extractor, one list, honest visibility.
    pub fn default_participant(node: IdentityKey, list: ContentId, default_space: SpaceId, now: UnixSecs) -> Self {
        Self {
            node,
            subscribed_lists: vec![list],
            served_spaces: vec![default_space],
            extractors: vec!["web".to_string()],
            query_visibility: QueryVisibility::Terminating,
            updated_at: now,
        }
    }
}
