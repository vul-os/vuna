# Vuna — Architecture

> Companion to [`01-design.md`](01-design.md) (the *why*) and [`00-viability.md`](00-viability.md)
> (the *should we*). This document is the *how*: the crate map, the data flow, the object
> lifecycle, the KOTVA seam, and the storage math the "index, not pages" claim rests on.

**Status as of this writing:** `vuna-core` (the contract) compiles and is unit-tested —
`quorum.rs`'s reconciliation logic is real and covered; everything else is typed but not wired.
Every other crate (`vuna-frontier`, `vuna-crawl`, `vuna-extract`, `vuna-index`, `vuna-query`,
`vuna-node`) is a stub: a module doc comment, `#![allow(dead_code)]`, and a `TODO(agent)`. `app/`
does not exist yet — no Tauri scaffold has been generated. Read everything below as the designed
shape, not a description of running code.

---

## 1. The crate map

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
                                 app/  (Tauri shell — not yet scaffolded)
```

| Crate | Depends on | Implements (from `vuna-core`) | Role | Status |
|---|---|---|---|---|
| **vuna-core** | — | — | Frozen contract: shared types + seam traits. No tantivy/libp2p/reqwest/crypto — always compiles offline. | Compiles, tested |
| **vuna-frontier** | `vuna-core` | `frontier::Frontier` | Distributed URL lists: subscribe, dedup, DHT crawl-assignment, K× replication of `UrlEntry`. | Stub |
| **vuna-crawl** | `vuna-core` | `crawl::Fetcher` (planned) | Polite fetch: reqwest + optional headless, robots.txt, per-host rate limits. Produces `FetchedPage`. | Stub |
| **vuna-extract** | `vuna-core` | `extract::Extractor` | Pluggable extractors. First two: `web` (chunks+links+snippet) and `retail` (declarative JSON-LD/JSON-endpoint adapters — see [`adapters/`](../adapters/)). | Stub |
| **vuna-index** | `vuna-core` | `index::Index`, `index::Embedder` | Tantivy keyword index + per-space HNSW vectors + link/knowledge graph. | Stub |
| **vuna-query** | `vuna-core` | `query::QueryEngine` | KOTVA SEARCH read path: local-first search, optional peer/indexer fan-out, Min-PPR merge. | Stub |
| **vuna-node** | `vuna-core`, `kotva-core` (tag-pinned) | `kotva::NodeIdentity`, `kotva::Publisher`, `kotva::ContentAddresser` | The daemon: roles, the crawl→extract→index→publish loop, and the **only** crate that touches `kotva-core`. | Stub |
| **app/** | `vuna-node` | — | Tauri desktop app (Rust backend + React) — the downloadable node every user runs. | Not scaffolded |

**Why `vuna-core` has no heavy deps:** every other crate builds *against* the trait objects it
declares, never against each other directly. `vuna-index` doesn't know `vuna-crawl` exists; it
only knows `FetchedPage` went in somewhere upstream and an `IndexedDoc` comes out. This is what
lets the stage crates be built, tested, and swapped independently — and what lets `vuna-core`
itself compile in a bare CI runner with no network and no substrate.

---

## 2. Data flow

```
distributed URL lists ──▶ crawl ──▶ extract ──▶ index (keyword + vectors/space + graph) ──▶ query
   (vuna-frontier)      (vuna-crawl) (vuna-extract)          (vuna-index)                 (vuna-query)
        │                                                                                     │
        └──────────────────────── KOTVA substrate: identity · PUB · DHT · SEARCH ─────────────┘
                                             (vuna-node binds kotva-core)
```

Five stages, one direction, one seam trait per boundary:

1. **Frontier → Crawl.** `Frontier::due(now, limit)` hands `vuna-crawl` the slice of `UrlEntry`
   this node is DHT-assigned to fetch next, staleness-first.
2. **Crawl → Extract.** `Fetcher` turns each URL into a `FetchedPage {url, status, content_type,
   body, fetched_at}`. Bytes only — the extractor decides charset and parse strategy.
3. **Extract → Index.** The extractor registry picks whichever registered `Extractor` reports
   `applies(&page)`, and calls `extract()`, producing an `Extraction::Web(WebDoc)` or
   `Extraction::Retail(Vec<RetailObservation>)`.
4. **Index.** `IndexedDoc::from_web` builds the durable per-URL record; `Index::upsert` adds the
   keyword+graph contribution (space-agnostic); `Embedder::embed(chunks)` produces vectors per
   served space, added via `Index::upsert_vectors`.
5. **Query.** `QueryEngine::search` reads the local shard first (`Source::Local`, works offline,
   zero coordinator), optionally fans out to peers or an opt-in `indexer` (`Source::Peer` /
   `Source::Indexer`), and merges with Min-PPR-style ranking into an `Answer`.

The retail vertical **does not** produce an `IndexedDoc` — a `RetailObservation` isn't a document,
it's a claim about a non-participant (the store never signs). Those go through
`vuna_core::quorum::reconcile` instead (§3.2), and only the *agreed* result — never a single
observer's word — is what a query surfaces.

---

## 3. Object lifecycle

Two paths share the frontier and crawl stages, then diverge at extraction because the two
verticals have fundamentally different trust shapes: a web page is a **signed-by-nobody-but-still
authoritative** document (the author's content governs, per SRCH-2); a store's stock level is a
**claim about an unwilling subject**, so it needs consensus, not a single fetch.

### 3.1 The web/RAG path

```
UrlList (signed, versioned, subscribable)
   │  Frontier::subscribe(list)
   ▼
UrlEntry { url, id, last_crawled, content_hash, last_embedded[space], applied[] }
   │  Frontier::due(now, limit) — DHT-assigned, staleness-first
   ▼
FetchedPage { url, status, content_type, body, fetched_at }
   │  Extractor::applies() → Extractor::extract()
   ▼
Extraction::Web(WebDoc { url, title, snippet, chunks, links, discovered })
   │  IndexedDoc::from_web(doc, url_id, now)
   ▼
IndexedDoc { url, url_id, title, snippet, chunks, indexed_at }   ◄── no raw page, ever
   │  Index::upsert(doc, links_to)         — keyword postings + graph edges
   │  Embedder::embed(doc.chunks)          — per served space
   │  Index::upsert_vectors(url_id, space, vectors)
   ▼
Frontier::record(entry)   — last_crawled/content_hash/last_embedded[space]/applied updated
   │
   ▼
QueryEngine::search(Query { text, space, limit })
   │  Index::search local-first → optional peer/indexer fan-out → Min-PPR merge
   ▼
Answer { hits: Vec<RankedHit>, complete_over_reachable }
```

Two details worth naming because they're the whole point of the design:

- **`IndexedDoc.chunks` retains chunk *text*, not just vectors.** This is the entire mechanism
  behind "no forced re-embed": when a new `EmbeddingSpace` is announced and a node opts in, it
  calls `Embedder::embed` again over text it already has. No re-crawl.
- **`content_hash` on `UrlEntry`** lets a re-crawl skip re-extraction entirely when the fetched
  bytes hash to the same value as last time — the recurring cost is crawl bandwidth + a hash
  comparison, not full re-extraction, on unchanged pages.

### 3.2 The retail/radar path

```
UrlEntry (same frontier, extractors=["retail"])
   │  FetchedPage
   ▼
Extraction::Retail(Vec<RetailObservation> { store, sku, availability, quantity,
                                             price_minor, currency, method, observed_at })
   │  one observation per (IdentityKey, store, sku) — the observing NODE's identity, not the store's
   ▼
quorum::reconcile(observers: &[(IdentityKey, RetailObservation)], now, QuorumParams { k, window_secs, qty_tolerance })
   │  1. one latest in-window vote per distinct IdentityKey (ballot-stuffing floor)
   │  2. availability by plurality, accepted only if support ≥ k
   │  3. quantity = median of agreeing observers, accepted only if ≥ k agree within tolerance
   ▼
Option<Agreed { availability, quantity, support, dissent }>
```

`reconcile` returns `None`, not a fabricated value, when fewer than `k` distinct identities agree
— the caller **must** surface "the network doesn't know" rather than guess. This is the honest
floor: distinct-`IdentityKey`-counting stops one key from voting itself to quorum, but it says
nothing about whether those `k` keys are `k` different *people*. Anchoring that is explicitly
punted to the caller "via KOTVA ATTEST" per the module doc — see [`00-viability.md`](00-viability.md)
§2 for why that punt is the single biggest open risk in the retail vertical, not a solved detail.

---

## 4. The KOTVA seam

`vuna_core::kotva` declares three traits and nothing else:

| Trait | Method(s) | KOTVA primitive it binds |
|---|---|---|
| `NodeIdentity` | `public() -> IdentityKey`, `sign(&self, msg) -> Vec<u8>` | **Identity** (KOTVA §1) — an Ed25519 keypair *is* the identity. |
| `Publisher` | `publish(canonical_bytes) -> ContentId`, `verify(bytes, sig, author) -> Result<()>` | **PUB** (§22 public objects + §25 pubsub feeds) — signed public objects on an append-only author feed. `UrlList`, `EmbeddingSpace`, and `NodeDescriptor` all publish through this. |
| `ContentAddresser` | `address(bytes) -> ContentId` | BLAKE3-256 of canonical bytes (§2.2) — what makes a `ContentId`/`UrlId` the same on every node that sees the same object. |

**`vuna-node` is the only crate in the workspace that depends on `kotva-core`, and it's pinned by
tag** (e.g. `core-v0.2.0`), never `HEAD` — the `Cargo.toml` comment calls this "the isango lesson"
(a prior integration that churned badly against a moving substrate branch). `vuna-node` implements
the three seam traits against the real `kotva-core` primitives; every other crate sees only the
trait objects via `vuna-core` and has no idea a Rust dependency called `kotva-core` exists. Two
consequences fall out of that:

- **The whole workspace except `vuna-node` compiles and tests offline.** No substrate, no network,
  no keys needed to run `cargo test` on `vuna-frontier`/`vuna-extract`/`vuna-index`/`vuna-query`.
- **The substrate binding is swappable in principle** — nothing about `vuna-core`'s traits
  mentions KOTVA by name — but this is a hygiene/testability property, not a stated goal of
  running Vuna over a different substrate. KOTVA is the only substrate Vuna targets.

**DHT and SEARCH are not literally traits in `vuna-core` today** — they show up at the crate level,
not the contract level. `vuna-frontier`'s crawl-assignment (§3.2's "DHT-assigned, staleness-first")
is designed to ride the libp2p Kademlia binding KOTVA's PUB/PUBSUB feed-reach already adopts
(KOTVA `profiles/search.md`, "Feed reach — libp2p + DMTAP-PUBSUB"). `vuna-query`'s fan-out is
designed as a KOTVA **SEARCH** profile client: local-first over the following-graph (here, the
lists a node subscribes to) is the floor that works with zero coordinator, and an opt-in `indexer`
coordinator adds reach without becoming authoritative — the same **SRCH-2** rule `index.rs`'s doc
comment cites ("derived, rebuildable, never authoritative"). Both of these are `vuna-node`/
`vuna-frontier`/`vuna-query` implementation concerns, wired through the three seam traits above —
`vuna-core` doesn't need its own `Dht` or `SearchFanout` trait because nothing upstream of the
substrate binding needs to call one directly.

---

## 5. Multi-space / pluggable-extractor plurality

Two independent axes of "add capability without a flag-day," both expressed the same way in
`node::NodeDescriptor`:

```rust
pub struct NodeDescriptor {
    pub subscribed_lists: Vec<ContentId>,   // which frontiers this node crawls
    pub served_spaces: Vec<SpaceId>,        // which embedding models this node embeds into
    pub extractors: Vec<ExtractorKind>,     // which verticals this node runs
    // ...
}
```

- **Embedding spaces** (`space::EmbeddingSpace`, id = `model_id@dim/quant`) are independent,
  parallel vector indexes over the same corpus. The keyword index and the link graph are
  model-agnostic and shared — they never redo work when a space is added or dropped. A node
  announces a new space as a signed PUB object; other nodes opt in on their own schedule.
  Because `IndexedDoc` retains chunk text, opting in is a local `Embedder::embed` call over
  already-fetched text, not a re-crawl (§3.1).
- **Extractor kinds** (`extract::ExtractorKind`, a plain `String` so core never changes for a new
  vertical) work the same way: `web` and `retail` are the first two `Extractor` impls; a
  `NodeDescriptor.extractors` list is exactly analogous to `served_spaces` — opt in, run
  independently, no coordination with nodes that haven't adopted it. The retail vertical goes
  further and pushes *most* of its per-site variation out of Rust entirely into declarative
  adapter manifests (`adapters/*.toml` — see [`adapters/README.md`](../adapters/README.md)), so
  that growing site coverage doesn't even require a new `Extractor` impl, just a new file.

`NodeDescriptor::default_participant` is the floor: one list, the default space, the `web`
extractor, `Terminating` query visibility. A node running exactly that is already a complete,
useful participant — everything else is opt-in surface area layered on top, never a prerequisite.

---

## 6. Storage math

The claim "index, not pages" only matters if the arithmetic backs it up. This section is that
arithmetic, not a vibe.

### 6.1 Per-page budget

| Component | Budget | Notes |
|---|---|---|
| Keyword postings | ~2.0 KB | Tantivy/BM25-style postings for a typical page's term set. |
| Embeddings (int8) | ~2.5 KB | One space at `int8` quantization, e.g. a 384-dim model at 1 byte/dim (`Quant::bytes_per_dim`) ≈ 384 B raw vector, plus HNSW graph overhead per `space::EmbeddingSpace::vector_bytes` (excludes ANN-graph edges) and per-space multiplication — budget assumes **one served space**; each additional space a node opts into adds roughly this line again. |
| Link/knowledge graph | ~0.3 KB | Outbound-edge records feeding Min-PPR ranking. |
| Metadata + snippet + short chunk text | ~0.7 KB | Title, snippet, `UrlEntry` bookkeeping (`last_crawled`, `content_hash`, `last_embedded` watermarks), and a small number of short retained chunks. |
| **Total** | **~5.5–6 KB/page** | Rounded to **6 KB** below. |

**This budget assumes short, RAG-sized chunk retention** — a page reduced to a handful of chunks
(a few hundred words total), not full-page HTML or full body text. `IndexedDoc.chunks` is what
makes re-embedding free (§3.1), but that mechanism only stays cheap if extractors are disciplined
about chunk size. A vertical that needs to retain large chunks (long-form article bodies, full
product descriptions) should budget that as an *additional* line, not assume it's absorbed here.

### 6.2 K/N scaling

Total index bytes ≈ `pages × 6 KB × K` (K = replication factor, how many nodes hold each shard),
spread across `N` nodes by the DHT, so expected **per-node** share is:

```
per_node_bytes ≈ pages × 6 KB × K / N
```

At `K = 3` (three-way redundant, survives churn without a central archive):

| Pages ＼ Nodes (N) | 1,000 | 10,000 | 100,000 | 1,000,000 |
|---|---|---|---|---|
| 100M | 1.8 GB | 180 MB | 18 MB | 1.8 MB |
| 1B | 18 GB | **1.8 GB** | 180 MB | 18 MB |
| 10B | 180 GB | **18 GB** | 1.8 GB | 180 MB |
| 100B | 1.8 TB | 180 GB | 18 GB | 1.8 GB |

The two bolded cells are the headline numbers: **a billion-page index, replicated 3×, spread
across 10,000 nodes, is ~1.8 GB/node**; a 10-billion-page index at the same N and K is ~18 GB/node.
Every 10× shift in either axis moves per-node storage by exactly 10× — it's a flat product, no
superlinear term, because nothing about the index format grows with network size.

Contrast: the open-web brief this design replaced (`00-discussion-history.md` §1) put a
"reliable" storage tier at 300–800 PB committed for real open-web scale — which resolves to
datacenter operators, not volunteers, the moment you try to run it. The index-only design is what
keeps this on a laptop's spare disk instead.

**Note on the two `K`s in this codebase:** the replication factor `K` above (index-shard
durability, a `vuna-frontier`/`vuna-index` concern) and the `k` in `quorum::QuorumParams`
(distinct-observer Sybil floor for retail reconciliation, §3.2) are **different knobs that happen
to share a letter and a default of 3** in current examples. Don't conflate them — one is about not
losing data when nodes churn, the other is about not trusting one identity's word for a stock
level.

### 6.3 The recurring cost is compute, not storage

Storage above is bought once per page (amortized over however long a shard is worth keeping) and
shrinks per-node as the network grows. **Compute does not shrink the same way**, and it recurs on
a schedule the network doesn't control:

- **Every page needs an embedding forward pass per served space**, at ingest and again at every
  re-crawl that changes `content_hash`. Adding a second default space roughly doubles embedding
  compute for every node serving both, permanently — vector storage growth is linear and cheap;
  embedding-compute growth is linear and not cheap, because it's CPU/GPU-seconds, not disk-bytes.
- **Freshness is a compute tax with no ceiling.** A stale shard is a staleness problem, not a
  correctness one (§01-design.md), but "eventually re-crawl and re-embed everything" is a
  recurring job whose cost scales with corpus size the same way the one-time ingest did — it just
  never stops.
- Storage has an obvious, near-universal volunteer analog: idle disk. Compute's volunteer analog —
  spare GPU/CPU cycles donated for free, on a schedule, indefinitely — has no comparably strong
  precedent (see [`00-viability.md`](00-viability.md) §1 for why this is the sharpest open risk in
  the whole design, not a footnote).
