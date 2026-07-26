# Architecture

Companion to [Getting started](/products/vuna/docs/getting-started) (the *what*) and
[Viability](/products/vuna/docs/viability) (the *should we*). This is the *how*: the crate map,
the data flow, the object lifecycle, the KOTVA seam, and the storage math the "index, not pages"
claim rests on.

**Status as of this writing:** `vuna-core` (the contract) compiles and is unit-tested —
`quorum.rs`'s reconciliation logic is real and covered; everything else is typed but not fully
wired. `vuna-frontier`, `vuna-crawl`, `vuna-extract`, and `vuna-index` are being built out.
`vuna-query` and `vuna-node` are stubs. `app/` exists and runs, but against a mock corpus, not a
live `vuna-node`. Read everything below as the designed shape, not a description of a running
network.

## The crate map

```
                       ┌─────────────┐
                       │  vuna-core  │  frozen contract: types + seam traits, no heavy deps
                       └──────┬──────┘
          ┌──────────┬────────┼────────┬──────────┐
          ▼          ▼        ▼        ▼          ▼
   vuna-frontier vuna-crawl vuna-extract vuna-index vuna-query
          │                                              │
          └──────────────────────┬───────────────────────┘
                                  ▼
                             vuna-node  (daemon: wiring + kotva-core binding)
                                  │
                                  ▼
                                 app/  (Tauri shell — builds, runs on mock data)
```

| Crate | Role | Status |
|---|---|---|
| **vuna-core** | Frozen contract: shared types + seam traits. No tantivy/libp2p/reqwest/crypto — always compiles offline. | Compiles, tested |
| **vuna-frontier** | Distributed URL lists: subscribe, dedup, DHT crawl-assignment, K× replication of each `UrlEntry`. | In progress |
| **vuna-crawl** | Polite fetch: reqwest + optional headless, robots.txt, per-host rate limits. Produces `FetchedPage`. | In progress |
| **vuna-extract** | Pluggable extractors: `web` (chunks + links + snippet) and `retail` (declarative JSON-LD/JSON-endpoint adapters). | In progress |
| **vuna-index** | Tantivy keyword index + per-space HNSW vectors + link/knowledge graph. | In progress |
| **vuna-query** | The KOTVA SEARCH read path: local-first search, optional peer/indexer fan-out, Min-PPR merge. | Stub |
| **vuna-node** | The daemon: roles, the crawl→extract→index→publish loop, and the **only** crate that touches `kotva-core` (pinned by tag, never `HEAD`). | Stub |
| **app/** | Tauri desktop app (Rust backend + React) — the downloadable node every user would run. | Builds, mock data only |

Every crate builds against the trait objects `vuna-core` declares, never against each other
directly — `vuna-index` doesn't know `vuna-crawl` exists, it only knows a `FetchedPage` went in
somewhere upstream and an `IndexedDoc` comes out. That's what lets the stage crates be built,
tested, and swapped independently, and what lets `vuna-core` itself compile in a bare CI runner
with no network and no substrate.

## Data flow — five stages, one direction

```
distributed URL lists ──▶ crawl ──▶ extract ──▶ index (keyword + vectors/space + graph) ──▶ query
   (vuna-frontier)      (vuna-crawl) (vuna-extract)          (vuna-index)                 (vuna-query)
        │                                                                                     │
        └──────────────────────── KOTVA substrate: identity · PUB · DHT · SEARCH ─────────────┘
                                             (vuna-node binds kotva-core)
```

1. **Frontier → Crawl.** `Frontier::due(now, limit)` hands the crawler the slice of `UrlEntry` this
   node is DHT-assigned to fetch next, staleness-first.
2. **Crawl → Extract.** The fetcher turns each URL into a `FetchedPage { url, status, content_type,
   body, fetched_at }` — bytes only; the extractor decides charset and parse strategy.
3. **Extract → Index.** Whichever registered extractor reports `applies(&page)` runs, producing
   `Extraction::Web(WebDoc)` or `Extraction::Retail(Vec<RetailObservation>)`.
4. **Index.** `IndexedDoc::from_web` builds the durable per-URL record; the keyword+graph
   contribution is space-agnostic; the embedder produces vectors per served space.
5. **Query.** The query engine reads the local shard first (works offline, zero coordinator),
   optionally fans out to peers or an opt-in indexer, and merges with Min-PPR-style ranking.

The retail vertical does **not** produce an `IndexedDoc` — a price/stock observation isn't a
document, it's a claim about a non-participant (the store never signs it). Those go through
quorum reconciliation instead (below), and only the *agreed* result — never a single observer's
word — is what a query ever surfaces.

## Two verticals, two trust shapes

A web page is signed-by-nobody-but-still-authoritative — the author's content governs. A store's
stock level is a claim about an unwilling subject, so it needs consensus, not a single fetch.

**Web/RAG path:** `UrlList → UrlEntry → FetchedPage → WebDoc → IndexedDoc` (chunks retained as
*text*, not just vectors — the whole mechanism behind "no forced re-embed") `→ Index::upsert +
Embedder::embed → QueryEngine::search`.

**Retail/radar path:** each observing node's identity produces one `RetailObservation` per
`(store, sku)`. `quorum::reconcile(observers, now, QuorumParams { k, window_secs, qty_tolerance })`:

1. one latest in-window vote per distinct identity key (a ballot-stuffing floor, not a
   personhood check),
2. availability accepted only if support ≥ `k`,
3. quantity = median of agreeing observers, accepted only if ≥ `k` agree within tolerance.

`reconcile` returns `None` — not a fabricated value — when fewer than `k` distinct identities
agree; the caller must surface "the network doesn't know." Distinct-identity-counting stops one
key from voting itself to quorum, but says nothing about whether those `k` keys are `k` different
*people* — anchoring that is explicitly left to the caller "via KOTVA ATTEST." See
[Viability](/products/vuna/docs/viability) for why this is the single biggest open risk in the
retail vertical, not a solved detail.

## The KOTVA seam

`vuna-node` is the **only** crate in the workspace that depends on `kotva-core`, and it's pinned
by tag (e.g. `core-v0.2.0`), never `HEAD` — the comment in the workspace `Cargo.toml` calls this
"the isango lesson," after a prior integration that churned badly against a moving substrate
branch. Three traits, nothing else:

| Trait | Binds |
|---|---|
| `NodeIdentity` | KOTVA identity — an Ed25519 keypair *is* the identity |
| `Publisher` | KOTVA PUB — signed public objects on an append-only author feed (`UrlList`, `EmbeddingSpace`, `NodeDescriptor` all publish through this) |
| `ContentAddresser` | BLAKE3-256 of canonical bytes — what makes a `ContentId` the same on every node that sees the same object |

Every other crate sees only these trait objects and has no idea a Rust dependency called
`kotva-core` exists — the whole workspace except `vuna-node` compiles and tests **offline**.

## Storage math

The "index, not pages" claim only matters if the arithmetic backs it up.

| Component | Budget | Notes |
|---|---|---|
| Keyword postings | ~2.0 KB | Tantivy/BM25-style postings for a typical page's term set |
| Embeddings (int8, one served space) | ~2.5 KB | Each additional served space adds roughly this line again |
| Link/knowledge graph | ~0.3 KB | Outbound-edge records feeding Min-PPR ranking |
| Metadata + snippet + short chunk text | ~0.7 KB | Title, snippet, frontier bookkeeping, a handful of short retained chunks |
| **Total** | **~6 KB/page** | Assumes short, RAG-sized chunk retention — not full-page HTML |

`per_node_bytes ≈ pages × 6 KB × K / N` (K = replication factor, N = node count). At K = 3:

| Pages ＼ Nodes | 1,000 | 10,000 | 100,000 | 1,000,000 |
|---|---|---|---|---|
| 100M | 1.8 GB | 180 MB | 18 MB | 1.8 MB |
| **1B** | 18 GB | **1.8 GB** | 180 MB | 18 MB |
| 10B | 180 GB | 18 GB | 1.8 GB | 180 MB |

**A billion-page index, replicated 3×, spread across 10,000 nodes, is ~1.8 GB/node.** Contrast:
the open-web brief this design replaced put a "reliable" storage tier at 300–800 PB committed for
real open-web scale — datacenter-operator territory the moment you try to run it.

**The recurring cost this table doesn't show is compute, not storage.** Every page needs an
embedding forward pass per served space, at ingest and again on every re-crawl that changes its
content hash — and that cost does *not* shrink as the network grows the way per-node storage does.
See [Viability](/products/vuna/docs/viability) §1 for why this is the sharpest open risk in the
whole design.
