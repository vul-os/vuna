//! Tauri commands — **mock data only**. This crate does not depend on the `vuna-*` workspace
//! crates yet; the types below are local structs that deliberately *mirror* the frozen contract in
//! `crates/vuna-core/src` so the swap-in later is a rename, not a redesign:
//!
//! | This file                    | Future source of truth                          |
//! |-------------------------------|--------------------------------------------------|
//! | `Hit`                         | `vuna_core::index::Hit` (url/title/snippet/score) |
//! | `Source`                       | `vuna_core::query::Source` (Local/Peer/Indexer)   |
//! | `RankedHit`                    | `vuna_core::query::RankedHit` (hit/source/rank)   |
//! | `search()`                     | `vuna_core::query::QueryEngine::search()` via `vuna-node` |
//! | `ExtractorInfo.kind`            | `vuna_core::extract::ExtractorKind` ("web"/"retail") |
//! | `SpaceInfo`                    | `vuna_core::space::EmbeddingSpace`                |
//! | `ListInfo`                    | a subscribed frontier list (`vuna-frontier`)      |
//! | `NodeStatus`                   | `vuna_core::node::NodeDescriptor` + live counters  |
//!
//! When real wiring lands, `vuna-node` runs behind a `tauri::State<Mutex<NodeHandle>>` and these
//! functions become thin delegations instead of mock generators — the command signatures and JSON
//! shape returned to the frontend should not need to change.

use serde::Serialize;

/// A ranked hit's origin — surfaced to the user, never hidden (mirrors KOTVA SEARCH's SRCH-9
/// disclosure rule: query-channel visibility and result provenance are always shown).
#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum Source {
    /// Answered from this node's own shard — always available, offline-safe.
    Local,
    /// Reached over the mesh/DHT from another volunteer node.
    Peer,
    /// An opt-in global indexer coordinator — adds reach, never authority.
    Indexer,
}

/// Mirrors `vuna_core::index::Hit` exactly: url / title / snippet / score.
#[derive(Clone, Debug, Serialize)]
pub struct Hit {
    pub url: String,
    pub title: Option<String>,
    pub snippet: String,
    pub score: f32,
}

/// Mirrors `vuna_core::query::RankedHit`: a `Hit` plus where it came from and its rank
/// contribution. `rank` is index-local, disclosed policy (Min-PPR / web-of-trust), never a
/// network-wide authority.
#[derive(Clone, Debug, Serialize)]
pub struct RankedHit {
    pub hit: Hit,
    pub source: Source,
    pub rank: f32,
}

/// One enabled extractor vertical (`vuna_core::extract::ExtractorKind` is a string so new
/// verticals need no core change).
#[derive(Clone, Debug, Serialize)]
pub struct ExtractorInfo {
    pub kind: String,
    pub label: String,
    pub enabled: bool,
}

/// One embedding space this node serves vectors for (`vuna_core::space::EmbeddingSpace`).
#[derive(Clone, Debug, Serialize)]
pub struct SpaceInfo {
    pub id: String,
    pub label: String,
    pub dim: usize,
    pub default: bool,
}

/// One subscribed frontier URL list (`vuna-frontier`).
#[derive(Clone, Debug, Serialize)]
pub struct ListInfo {
    pub id: String,
    pub label: String,
    pub url_count: u64,
}

#[derive(Clone, Copy, Debug, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum QueryVisibility {
    /// This node's operator can read queries in the clear (the honest v0 default).
    Terminating,
    /// Queries are shielded from the operator (PIR/DP — a later layer). Not yet produced by the
    /// mock (no node runs this today), kept here so the variant/serde shape already matches
    /// `vuna_core::node::QueryVisibility` when real wiring lands.
    #[allow(dead_code)]
    Blind,
}

/// Mirrors `vuna_core::node::NodeDescriptor` plus a few live counters the desktop shell wants.
#[derive(Clone, Debug, Serialize)]
pub struct NodeStatus {
    pub node_id: String,
    pub online: bool,
    pub uptime_secs: u64,
    pub query_visibility: QueryVisibility,
    pub extractors: Vec<ExtractorInfo>,
    pub served_spaces: Vec<SpaceInfo>,
    pub subscribed_lists: Vec<ListInfo>,
    pub peers_connected: u32,
}

/// The dashboard's headline numbers — a rollup of the same derived, rebuildable state
/// `NodeStatus` describes (KOTVA SEARCH SRCH-2: derived, never authoritative; a stale number is a
/// staleness problem, not a correctness one).
#[derive(Clone, Debug, Serialize)]
pub struct Stats {
    pub docs_indexed: u64,
    pub storage_mb: f64,
    pub spaces_served: u32,
    pub lists_subscribed: u32,
    pub peers: u32,
    pub keyword_postings: u64,
    pub graph_edges: u64,
}

fn node_status_mock() -> NodeStatus {
    NodeStatus {
        node_id: "vuna1qf8k…m2wz".to_string(),
        online: true,
        uptime_secs: 62 * 60 * 9 + 41,
        query_visibility: QueryVisibility::Terminating,
        extractors: vec![
            ExtractorInfo { kind: "web".into(), label: "Web (chunks + links)".into(), enabled: true },
            ExtractorInfo { kind: "retail".into(), label: "Retail radar (JSON-LD)".into(), enabled: false },
        ],
        served_spaces: vec![
            SpaceInfo { id: "bge-small-en-v1.5@384/int8".into(), label: "bge-small-en (multilingual default)".into(), dim: 384, default: true },
            SpaceInfo { id: "gte-base@768/int8".into(), label: "gte-base (opt-in)".into(), dim: 768, default: false },
        ],
        subscribed_lists: vec![
            ListInfo { id: "list:default-seed".into(), label: "Default seed list".into(), url_count: 84_213 },
            ListInfo { id: "list:oss-docs".into(), label: "Open-source docs & wikis".into(), url_count: 12_940 },
            ListInfo { id: "list:indie-web".into(), label: "Indie web ring".into(), url_count: 3_802 },
        ],
        peers_connected: 27,
    }
}

fn stats_mock() -> Stats {
    Stats {
        docs_indexed: 100_955,
        storage_mb: 1_847.3,
        spaces_served: 2,
        lists_subscribed: 3,
        peers: 27,
        keyword_postings: 4_218_006,
        graph_edges: 512_774,
    }
}

/// One mock corpus entry: enough fields to synthesize a `RankedHit` on demand.
struct MockDoc {
    url: &'static str,
    title: &'static str,
    snippet: &'static str,
    base_score: f32,
    source: Source,
}

const CORPUS: &[MockDoc] = &[
    MockDoc {
        url: "https://docs.kotva.dev/substrate/search",
        title: "SEARCH — the KOTVA read path",
        snippet: "Local-first query fan-out with optional peer and indexer reach. No shard is presented as complete against the whole network (SRCH-7).",
        base_score: 9.4,
        source: Source::Local,
    },
    MockDoc {
        url: "https://en.wikipedia.org/wiki/Distributed_hash_table",
        title: "Distributed hash table — Wikipedia",
        snippet: "A DHT provides a lookup service similar to a hash table: (key, value) pairs are stored, and any participating node can retrieve the value for a given key.",
        base_score: 8.1,
        source: Source::Peer,
    },
    MockDoc {
        url: "https://mwmbl.org/blog/how-mwmbl-works",
        title: "How Mwmbl works — a non-profit, ad-free search engine",
        snippet: "Volunteers run crawlers that contribute pages to a shared index. The Firefox extension lets you crawl the pages you already visit.",
        base_score: 7.9,
        source: Source::Peer,
    },
    MockDoc {
        url: "https://tantivy-search.github.io/tantivy/",
        title: "tantivy — a full-text search engine library in Rust",
        snippet: "Tantivy is inspired by Apache Lucene and focuses on being a strong, fast, and simple full-text search library.",
        base_score: 7.4,
        source: Source::Local,
    },
    MockDoc {
        url: "https://blog.indieweb.org/2024/embeddings-that-outlive-their-model",
        title: "Embeddings that outlive their model",
        snippet: "Store the chunk text, not just the vector, and a new model becomes a local recompute instead of a re-crawl of the whole corpus.",
        base_score: 7.2,
        source: Source::Indexer,
    },
    MockDoc {
        url: "https://arxiv.org/abs/2112.09332",
        title: "Personalized PageRank on web-scale graphs (Min-PPR)",
        snippet: "We present an approximate method for computing personalized PageRank that scales to graphs with billions of edges without global recomputation.",
        base_score: 6.8,
        source: Source::Indexer,
    },
    MockDoc {
        url: "https://github.com/libp2p/specs/blob/master/relay/circuit-v2.md",
        title: "Circuit Relay v2 — libp2p specs",
        snippet: "A relay allows two peers that cannot establish a direct connection to communicate through a relay node using hop and stop protocols.",
        base_score: 6.5,
        source: Source::Peer,
    },
    MockDoc {
        url: "https://ubuntu.com/blog/tag/uBlock-Origin",
        title: "Subscribable filter lists: the ad-block model, applied to crawling",
        snippet: "Anyone can publish a list; you choose which to trust. Deduplication and assignment happen by content hash, never by a central authority.",
        base_score: 6.1,
        source: Source::Local,
    },
    MockDoc {
        url: "https://retailradar.example/notes/json-ld-availability",
        title: "Reading stock availability from schema.org JSON-LD",
        snippet: "Structured-data parsing beats cart-probing: `availability: InStock` is already on the page, no scraping heuristics required.",
        base_score: 5.8,
        source: Source::Indexer,
    },
    MockDoc {
        url: "https://news.ycombinator.com/item?id=39482017",
        title: "Show HN: a decentralized index that stores the index, not the pages",
        snippet: "Comments: \"finally someone doing the ad-filter-list model for a crawl frontier, this scales differently than a full common-crawl mirror.\"",
        base_score: 5.4,
        source: Source::Peer,
    },
    MockDoc {
        url: "https://www.gutenberg.org/ebooks/subjects",
        title: "Project Gutenberg — Browse by subject",
        snippet: "70,000+ free eBooks, indexed by subject, author, and language, all in the public domain.",
        base_score: 4.6,
        source: Source::Local,
    },
    MockDoc {
        url: "https://rust-lang.github.io/async-book/",
        title: "Asynchronous Programming in Rust",
        snippet: "async/await is a way to write functions that can 'pause', return control to the runtime, and resume where they left off.",
        base_score: 4.2,
        source: Source::Peer,
    },
];

/// A relevance-ish score bump for a naive substring match — enough to make results reorder
/// sensibly by query, nothing more. Real ranking is `vuna-query`'s BM25 + Min-PPR merge.
fn overlap_bonus(query_lower: &str, doc: &MockDoc) -> f32 {
    if query_lower.is_empty() {
        return 0.0;
    }
    let hay = format!("{} {} {}", doc.title, doc.snippet, doc.url).to_lowercase();
    let mut bonus = 0.0f32;
    for term in query_lower.split_whitespace() {
        if term.len() < 2 {
            continue;
        }
        if hay.contains(term) {
            bonus += 3.0;
        }
    }
    bonus
}

#[tauri::command]
pub fn search(query: String) -> Vec<RankedHit> {
    let q_lower = query.trim().to_lowercase();

    let mut ranked: Vec<(f32, &MockDoc)> = CORPUS
        .iter()
        .map(|doc| (doc.base_score + overlap_bonus(&q_lower, doc), doc))
        .collect();

    // Empty query -> trending/default set (still useful, per the design doc: "a node serving the
    // default space + one list is already a complete, useful participant").
    if q_lower.is_empty() {
        ranked.sort_by(|a, b| b.0.partial_cmp(&a.0).unwrap());
    } else {
        // Only keep docs that actually matched something, else fall back to the full corpus
        // sorted by base relevance so the UI never shows a jarring empty state for a scaffold.
        let matched: Vec<_> = ranked.iter().filter(|(score, doc)| *score > doc.base_score).cloned().collect();
        ranked = if matched.is_empty() { ranked } else { matched };
        ranked.sort_by(|a, b| b.0.partial_cmp(&a.0).unwrap());
    }

    ranked
        .into_iter()
        .take(8)
        .enumerate()
        .map(|(i, (score, doc))| RankedHit {
            hit: Hit {
                url: doc.url.to_string(),
                title: Some(doc.title.to_string()),
                snippet: doc.snippet.to_string(),
                score,
            },
            source: doc.source,
            rank: (1.0 / (i as f32 + 1.0)) * 0.9,
        })
        .collect()
}

#[tauri::command]
pub fn node_status() -> NodeStatus {
    node_status_mock()
}

#[tauri::command]
pub fn stats() -> Stats {
    stats_mock()
}
