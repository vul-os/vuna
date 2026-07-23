// Thin wrapper around the three Tauri commands (`search`, `node_status`, `stats`). Falls back to
// an in-process mock when the page is opened outside the Tauri shell (plain `vite dev`/`preview`
// in a browser) so the UI can be designed and screenshotted without a full `tauri dev` build.
// Inside the real app this delegates straight to `invoke`, which calls the mock Rust commands in
// `src-tauri/src/commands.rs` today and real `vuna-node` calls once that wiring lands.

import type { NodeStatus, RankedHit, Stats } from "../types";

const inTauri = typeof window !== "undefined" && "__TAURI_INTERNALS__" in window;

async function invokeOrMock<T>(cmd: string, args: Record<string, unknown>, mock: () => T): Promise<T> {
  if (inTauri) {
    const { invoke } = await import("@tauri-apps/api/core");
    return invoke<T>(cmd, args);
  }
  // Browser-only fallback: small artificial delay so loading states are visible in preview.
  await new Promise((r) => setTimeout(r, 180));
  return mock();
}

export function search(query: string): Promise<RankedHit[]> {
  return invokeOrMock("search", { query }, () => mockSearch(query));
}

export function nodeStatus(): Promise<NodeStatus> {
  return invokeOrMock("node_status", {}, mockNodeStatus);
}

export function stats(): Promise<Stats> {
  return invokeOrMock("stats", {}, mockStats);
}

// ---- browser-preview mocks (kept intentionally small; the real corpus lives in Rust) ----

function mockNodeStatus(): NodeStatus {
  return {
    node_id: "vuna1qf8k…m2wz",
    online: true,
    uptime_secs: 62 * 60 * 9 + 41,
    query_visibility: "terminating",
    extractors: [
      { kind: "web", label: "Web (chunks + links)", enabled: true },
      { kind: "retail", label: "Retail radar (JSON-LD)", enabled: false },
    ],
    served_spaces: [
      { id: "bge-small-en-v1.5@384/int8", label: "bge-small-en (multilingual default)", dim: 384, default: true },
      { id: "gte-base@768/int8", label: "gte-base (opt-in)", dim: 768, default: false },
    ],
    subscribed_lists: [
      { id: "list:default-seed", label: "Default seed list", url_count: 84213 },
      { id: "list:oss-docs", label: "Open-source docs & wikis", url_count: 12940 },
      { id: "list:indie-web", label: "Indie web ring", url_count: 3802 },
    ],
    peers_connected: 27,
  };
}

function mockStats(): Stats {
  return {
    docs_indexed: 100955,
    storage_mb: 1847.3,
    spaces_served: 2,
    lists_subscribed: 3,
    peers: 27,
    keyword_postings: 4218006,
    graph_edges: 512774,
  };
}

interface MockDoc {
  url: string;
  title: string;
  snippet: string;
  baseScore: number;
  source: "local" | "peer" | "indexer";
}

const CORPUS: MockDoc[] = [
  { url: "https://docs.kotva.dev/substrate/search", title: "SEARCH — the KOTVA read path", snippet: "Local-first query fan-out with optional peer and indexer reach. No shard is presented as complete against the whole network.", baseScore: 9.4, source: "local" },
  { url: "https://en.wikipedia.org/wiki/Distributed_hash_table", title: "Distributed hash table — Wikipedia", snippet: "A DHT provides a lookup service similar to a hash table: (key, value) pairs are stored, and any participating node can retrieve the value.", baseScore: 8.1, source: "peer" },
  { url: "https://mwmbl.org/blog/how-mwmbl-works", title: "How Mwmbl works — a non-profit, ad-free search engine", snippet: "Volunteers run crawlers that contribute pages to a shared index. The Firefox extension lets you crawl the pages you already visit.", baseScore: 7.9, source: "peer" },
  { url: "https://tantivy-search.github.io/tantivy/", title: "tantivy — a full-text search engine library in Rust", snippet: "Tantivy is inspired by Apache Lucene and focuses on being a strong, fast, and simple full-text search library.", baseScore: 7.4, source: "local" },
  { url: "https://blog.indieweb.org/2024/embeddings-that-outlive-their-model", title: "Embeddings that outlive their model", snippet: "Store the chunk text, not just the vector, and a new model becomes a local recompute instead of a re-crawl of the whole corpus.", baseScore: 7.2, source: "indexer" },
  { url: "https://arxiv.org/abs/2112.09332", title: "Personalized PageRank on web-scale graphs (Min-PPR)", snippet: "We present an approximate method for computing personalized PageRank that scales to graphs with billions of edges.", baseScore: 6.8, source: "indexer" },
  { url: "https://github.com/libp2p/specs/blob/master/relay/circuit-v2.md", title: "Circuit Relay v2 — libp2p specs", snippet: "A relay allows two peers that cannot establish a direct connection to communicate through a relay node using hop and stop protocols.", baseScore: 6.5, source: "peer" },
  { url: "https://ubuntu.com/blog/tag/uBlock-Origin", title: "Subscribable filter lists: the ad-block model, applied to crawling", snippet: "Anyone can publish a list; you choose which to trust. Deduplication and assignment happen by content hash, never a central authority.", baseScore: 6.1, source: "local" },
  { url: "https://retailradar.example/notes/json-ld-availability", title: "Reading stock availability from schema.org JSON-LD", snippet: "Structured-data parsing beats cart-probing: availability: InStock is already on the page, no scraping heuristics required.", baseScore: 5.8, source: "indexer" },
  { url: "https://news.ycombinator.com/item?id=39482017", title: "Show HN: a decentralized index that stores the index, not the pages", snippet: "\"Finally someone doing the ad-filter-list model for a crawl frontier — this scales differently than a full common-crawl mirror.\"", baseScore: 5.4, source: "peer" },
  { url: "https://www.gutenberg.org/ebooks/subjects", title: "Project Gutenberg — Browse by subject", snippet: "70,000+ free eBooks, indexed by subject, author, and language, all in the public domain.", baseScore: 4.6, source: "local" },
  { url: "https://rust-lang.github.io/async-book/", title: "Asynchronous Programming in Rust", snippet: "async/await is a way to write functions that can 'pause', return control to the runtime, and resume where they left off.", baseScore: 4.2, source: "peer" },
];

function mockSearch(query: string): RankedHit[] {
  const q = query.trim().toLowerCase();
  const terms = q.split(/\s+/).filter((t) => t.length >= 2);

  let scored = CORPUS.map((doc) => {
    const hay = `${doc.title} ${doc.snippet} ${doc.url}`.toLowerCase();
    const bonus = terms.reduce((acc, t) => (hay.includes(t) ? acc + 3 : acc), 0);
    return { doc, score: doc.baseScore + bonus };
  });

  if (q.length > 0) {
    const matched = scored.filter((s) => s.score > s.doc.baseScore);
    if (matched.length > 0) scored = matched;
  }

  scored.sort((a, b) => b.score - a.score);

  return scored.slice(0, 8).map(({ doc, score }, i) => ({
    hit: { url: doc.url, title: doc.title, snippet: doc.snippet, score },
    source: doc.source,
    rank: (1 / (i + 1)) * 0.9,
  }));
}
