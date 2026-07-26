<div align="center">

# 🌾 Vuna

### Reap the open web.

A decentralized **crawl → extract → index/graph** engine that stores the
**index, not the pages** — search/RAG and retail-radar as verticals of one engine,
on the [KOTVA](https://vulos.org) substrate. *Vuna* is Swahili for **to harvest / reap**.

[![License: MIT OR Apache-2.0](https://img.shields.io/badge/License-MIT%20OR%20Apache--2.0-D0471F.svg)](LICENSE-MIT)
[![Rust](https://img.shields.io/badge/Rust-1.75+-CE8B4E?logo=rust&logoColor=white)](https://www.rust-lang.org)
[![Tauri](https://img.shields.io/badge/Tauri-2-E9A23B?logo=tauri&logoColor=white)](https://tauri.app)
[![Substrate](https://img.shields.io/badge/substrate-KOTVA-14B8A6)](https://vulos.org)
[![Status](https://img.shields.io/badge/status-v0%20preview-E07A5F)](#status)
[![Tests](https://img.shields.io/badge/tests-63%20passing-6A994E)](#build--run)

[**Design**](docs/01-design.md) · [**Architecture**](docs/02-architecture.md) · [**Viability**](docs/00-viability.md) · [**Product page**](https://vulos.org/products/vuna)

<br/>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/shots/search-dark.png">
  <img src="docs/shots/search-light.png" alt="Vuna desktop app — local + peer search results with source chips and a node dashboard" width="900">
</picture>

<sub><em>The desktop app: local + peer results with source chips, and your node's dashboard. All data shown is <strong>mock</strong> — the query/node wiring is Wave 2 (see <a href="#status">status</a>).</em></sub>

</div>

---

## Status

**v0 preview — honest about what's real.** The substrate contract and four of the six
stage crates are implemented and tested (**63 tests green**); the desktop app builds and
runs on **mock data**. It does not yet crawl-to-query end to end.

| Component | State |
|---|---|
| `vuna-core` — frozen contract (types, seam traits, retail **quorum** reconciler) | ✅ done, tested |
| `vuna-crawl` — polite fetch (robots.txt, per-host rate-limit, body cap) | ✅ done (24 tests) |
| `vuna-extract` — `web` (chunks+links) + `retail` (JSON-LD/OG) extractors | ✅ done (13 tests) |
| `vuna-index` — tantivy BM25 + per-space HNSW vectors + link graph | ✅ done (9 tests) |
| `vuna-frontier` — distributed URL lists, dedup, DHT assignment | ✅ done (12 tests) |
| `vuna-query` — SEARCH read path (local-first + fan-out + Min-PPR) | 🚧 Wave 2 stub |
| `vuna-node` — daemon + `kotva-core` binding | 🚧 Wave 2 stub |
| `app/` — Tauri v2 desktop node (React) | ✅ builds, mock data |

---

## What it is

A network of volunteer nodes crawls a **distributed, subscribable list of URLs**, runs
**pluggable extractors** over each page, and contributes the result — keyword postings,
vector embeddings, link/knowledge-graph edges, a snippet, and a **pointer back to the live
URL** — into a shared index replicated K× across the network. **Pages are never archived**,
which collapses storage from *petabytes (datacenters)* to *~2 GB/node (volunteers)* for a
billion-page index. No token, no page archive, no new crypto — it rides KOTVA identity /
PUB / DHT / SEARCH.

## Why it's different — everything is opt-in plurality

- **Embedding models** — each is an independent, parallel vector *space*. The keyword index
  and graph are model-agnostic and shared. New model → new space, adopted by **opt-in**,
  never a fleet-wide flag-day. Shards keep chunk *text*, so re-embedding is local recompute,
  not a re-crawl. *(This is what makes it future-proof instead of a lock-in trap.)*
- **Verticals** — each is an **extractor**. `web` emits chunks+links (search/RAG); `retail`
  emits price/stock observations (radar). Same crawl/frontier/distribution/query underneath.
- A node serving the **default space + one URL list** is already a complete participant.

## Screenshots

<div align="center">
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/shots/dashboard-dark.png">
  <img src="docs/shots/dashboard-light.png" alt="Vuna node dashboard — docs indexed, storage, peers, spaces served, extractors, and query-visibility disclosure" width="380">
</picture>
<br/>
<sub><em>Your node, honestly surfaced: spaces served, lists subscribed, docs indexed, storage used, and a query-visibility disclosure (no hidden telemetry).</em></sub>
</div>

## How it works

```
distributed URL lists ─▶ crawl ─▶ extract ─▶ index (keyword + vectors/space + graph) ─▶ query
   (vuna-frontier)     (vuna-crawl)(vuna-extract)          (vuna-index)               (vuna-query)
        │                                                                                  │
        └───────────────────── KOTVA: identity · PUB · DHT · SEARCH ───────────────────────┘
                                      (vuna-node binds kotva-core)
```

The index is **derived, rebuildable, never authoritative** (KOTVA SEARCH SRCH-2): on any
disagreement, the author's signed content wins. For the retail vertical — where the store is
a *non-participant* that signs nothing — consensus of **K distinct, anchored observers** is
the only ground truth (`vuna-core::quorum`, one observer can't stuff the ballot).

## Architecture

The workspace compiles offline: `vuna-core` has minimal deps and defines the **frozen
contract** (types + seam traits); heavy deps (tantivy, HNSW, reqwest, embedding runtimes)
live only in the crate that needs them. `vuna-node` is the sole `kotva-core` dependent,
pinned by tag. Full write-up in [`docs/02-architecture.md`](docs/02-architecture.md).

## Storage — why volunteers suffice

Storing the index (not pages), ~6 KB/page: keyword ~2 KB · embeddings (int8) ~2.5 KB ·
graph ~0.3 KB · metadata ~0.7 KB.

| Corpus | 1 copy | ×3 redundant / 10k nodes |
|---|---|---|
| 1 B pages | ~6 TB | **~1.8 GB / node** |
| 10 B pages (Google-ish) | ~60 TB | ~18 GB / node |

Compute (embedding), not storage, is the recurring cost — and it's embarrassingly parallel.
The honest trade-offs (compute vs. disk volunteers, Sybil at small scale, freshness) are in
[`docs/00-viability.md`](docs/00-viability.md) — read it before believing the pitch.

## Build & run

```bash
# engine (workspace)
cargo test --workspace        # 63 tests green
cargo build --workspace

# desktop app (mock data today)
cd app && npm install && npm run build
cargo tauri dev               # or: npm run tauri dev
```

## Roadmap

- **Wave 2** — `vuna-query` (local-first fan-out + Min-PPR ranking) and `vuna-node`
  (crawl→extract→index→publish loop, `kotva-core` binding) to close the end-to-end path.
- **Declarative adapters** — the long tail of sites as `adapters/*.toml`, not Rust.
- **Roadmap app** — an opt-in browser extension contributing URLs-you-visit to the frontier
  (Mwmbl's proven model), off by default.

## License

[MIT](LICENSE-MIT) OR [Apache-2.0](LICENSE-APACHE) — © VulOS. Vuna is a VulOS
project; source and issues at [github.com/vul-os/vuna](https://github.com/vul-os/vuna).

---

<p align="center">
  <sub><a href="https://vulos.org"><b>vulos</b></a> — open by design</sub>
</p>
