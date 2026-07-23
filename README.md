# Vuna

> **Swahili: *to harvest / reap.*** Vuna reaps signal from the public web across a network of
> volunteer nodes and weaves it into a shared, ownerless index + knowledge graph — **storing the
> index, not the pages.** One engine, pluggable extractors: **web search/RAG** and **retail radar**
> are the first two verticals.

**Status:** v0 scaffold. The frozen contract (`vuna-core`) compiles and is tested; the stage crates
are being built out. Not yet shippable.

---

## The idea in one paragraph

A decentralized **crawl → extract → index/graph** engine on the [KOTVA](../kotva) substrate. Nodes
crawl a **distributed, subscribable list of URLs**, run **pluggable extractors** over each page, and
contribute the result — keyword postings, vector embeddings, link/knowledge-graph edges, a snippet,
and a **pointer back to the live URL** — into a shared index replicated K× across the network. Pages
themselves are never archived, which is what collapses storage from *petabytes (datacenters)* to
*~2 GB/node (volunteers)* for a billion-page index. No token, no page archive, no new crypto.

## Why it's future-proof: everything is opt-in plurality

- **Embedding models** — each is an independent, parallel vector *space*. The keyword index and
  graph are model-agnostic and shared. New model → new space, adopted by **opt-in**, never a
  fleet-wide flag-day. Shards keep chunk *text*, so re-embedding is local recompute, not a re-crawl.
- **Verticals** — each is an **extractor**. `web` emits chunks+links; `retail` emits price/stock
  observations. New vertical → new extractor, same crawl/frontier/distribution/query underneath.

A node serving the **default space + one URL list** is already a complete, useful participant.

## Architecture

```
distributed URL lists ──▶ crawl ──▶ extract ──▶ index (keyword + vectors/space + graph) ──▶ query
   (vuna-frontier)      (vuna-crawl) (vuna-extract)          (vuna-index)                 (vuna-query)
        │                                                                                     │
        └──────────────────────── KOTVA substrate: identity · PUB · DHT · SEARCH ─────────────┘
                                             (vuna-node binds kotva-core)
```

| Crate | Role |
|-------|------|
| **vuna-core** | Frozen contract: shared types + seam traits. No heavy deps; compiles offline. |
| **vuna-frontier** | Distributed URL lists — subscribe, dedup, DHT crawl-assignment, K× replication. |
| **vuna-crawl** | Polite fetch — reqwest + optional headless, robots.txt, per-host rate limits. |
| **vuna-extract** | Pluggable extractors: `web` (chunks+links) and `retail` (JSON-LD/OpenGraph). |
| **vuna-index** | Tantivy keyword index + per-space HNSW vectors + link/knowledge graph. |
| **vuna-query** | SEARCH read path: local-first + optional fan-out + Min-PPR ranking. |
| **vuna-node** | The daemon; binds `kotva-core` (pinned by tag) behind the core Kotva seams. |
| **app/** | Tauri desktop app (Rust backend + React) — the downloadable node every user runs. |

## Language & shape

**Rust + Tauri.** The node must speak the KOTVA wire (`kotva-core` is Rust, tag-pinned), and Tauri
gives a downloadable cross-platform app so every user runs their own node — the decentralization is
in the distribution, not just the protocol. Per-store/site scraper detail is **declarative adapters**
(`adapters/*.toml`), so the long tail of sites is data, not Rust.

## Discipline (so it ships)

- **Bounded corpus first** (federation content or one vertical), then open web as a research track.
- Combine the *engine*; **launch one vertical live.** Pluggable-from-day-one ≠ everything-at-once.
- The index is **derived, rebuildable, never authoritative** — a stale shard is a staleness problem,
  not a correctness one.

See [`docs/`](docs/) for the design (`01-design.md`), the honest viability assessment
(`00-viability.md`), and the historical reasoning (`00-discussion-history.md`).

## License

MIT OR Apache-2.0.
