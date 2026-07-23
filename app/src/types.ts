// Mirrors the JSON shape of `app/src-tauri/src/commands.rs`, which in turn mirrors the frozen
// contract in `crates/vuna-core/src` (index::Hit, query::{Source,RankedHit}, node::NodeDescriptor).
// See commands.rs's doc comment for the exact future-wiring mapping.

export type Source = "local" | "peer" | "indexer";

export interface Hit {
  url: string;
  title: string | null;
  snippet: string;
  score: number;
}

export interface RankedHit {
  hit: Hit;
  source: Source;
  rank: number;
}

export interface ExtractorInfo {
  kind: string;
  label: string;
  enabled: boolean;
}

export interface SpaceInfo {
  id: string;
  label: string;
  dim: number;
  default: boolean;
}

export interface ListInfo {
  id: string;
  label: string;
  url_count: number;
}

export type QueryVisibility = "terminating" | "blind";

export interface NodeStatus {
  node_id: string;
  online: boolean;
  uptime_secs: number;
  query_visibility: QueryVisibility;
  extractors: ExtractorInfo[];
  served_spaces: SpaceInfo[];
  subscribed_lists: ListInfo[];
  peers_connected: number;
}

export interface Stats {
  docs_indexed: number;
  storage_mb: number;
  spaces_served: number;
  lists_subscribed: number;
  peers: number;
  keyword_postings: number;
  graph_edges: number;
}
