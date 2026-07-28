# Vuna

**Vuna** (Swahili: *to harvest / reap*) is a decentralized **crawl → extract → index/graph** engine
built on the [KOTVA](/kotva) substrate. It stores the **index, not the pages** — keyword postings,
vector embeddings, link/knowledge-graph edges, a snippet, and a pointer back to the live URL —
replicated across a network of volunteer nodes. One engine, pluggable extractors: **web search/RAG**
and **retail radar** (price/stock observation) are the first two verticals.

> **Status: v0 scaffold.** This is early software, not a shipped product. Read this page as a
> preview of the design and what's actually running today, not a claim that Vuna is usable yet.
> See the [Architecture](/products/vuna/docs/architecture) chapter for a precise crate-by-crate
> status table, and [Viability](/products/vuna/docs/viability) for the honest "should we build
> this" assessment — it is unusually candid about the two open risks (compute-volunteer supply and
> Sybil resistance) that no version of this design has solved yet.

## The idea in one paragraph

Nodes crawl a **distributed, subscribable list of URLs** (the ad-filter-list model — publish a
list, others opt in, no central frontier authority), run **pluggable extractors** over each page,
and contribute the result into a shared index replicated K× across the network. Pages themselves
are never archived, which is what collapses storage from *petabytes* (the datacenter-operator
scale a "reliable" open-web archive needs) to **single-digit gigabytes per node** even at
billion-page scale. No token, no page archive, no new crypto — it rides KOTVA's identity, PUB,
DHT, and SEARCH primitives.

## Why it's built this way

- **Multiple embedding models, opt-in.** Each embedding model gets its own independent, parallel
  vector index. The keyword index and the link graph are shared and model-agnostic. A new model
  is a new *space* nodes opt into on their own schedule — never a fleet-wide re-embed. Shards keep
  the chunk *text*, not just vectors, so adopting a new space is a local recompute, not a re-crawl.
- **Pluggable verticals, one engine.** `web` emits chunks + links for search/RAG. `retail` emits
  price/stock observations from a store's own structured data (JSON-LD/OpenGraph), reconciled by
  quorum across independent observing nodes rather than trusted from a single fetch. A new
  vertical is a new extractor over the same crawl/frontier/index/query stack.
- **Bounded corpus first.** Open-web crawling inherits a permanent SEO adversary and an unsolved
  permissionless-trust problem. Vuna's discipline is federation content or one curated vertical
  first; the open web is a research track, never a v1 promise.

## What's real today

| Piece | State |
|---|---|
| `vuna-core` (frozen contract: types + seam traits) | Compiles offline, unit-tested |
| `vuna-frontier`, `vuna-crawl`, `vuna-extract`, `vuna-index` | Implemented and unit-tested individually — not yet wired end to end |
| `vuna-query`, `vuna-node` (the daemon that binds the KOTVA substrate) | Stubs |
| `app/` — the Tauri desktop shell | Builds and runs, **against a mock corpus** (no live daemon behind it yet) |
| Test suite | 132 tests, concentrated in the frontier/crawl/extract/index/core layer, run in CI |

**132 tests passing does not mean 132 features shipped.** It means the parts that are wired compile
and behave as designed against their own fixtures — there are zero live nodes and zero real users
today. See [Architecture](/products/vuna/docs/architecture) for the crate-level detail.

## Try the workspace

```bash
git clone https://github.com/vul-os/vuna
cd vuna

# run the whole test suite (offline — no substrate, no network, no keys needed
# for every crate except vuna-node)
cargo test --workspace

# run just the frozen contract's tests
cargo test -p vuna-core

# build and run the desktop shell (mock data — see the note on the app/lib/api.ts
# fallback in Architecture)
cd app && npm install && npm run tauri dev
```

`app/` also runs as a plain Vite app (`npm run dev` from `app/`) without Tauri at all — the API
layer falls back to an in-process mock corpus so the UI can be built and previewed on its own.

## Where to go next

- [**Architecture**](/products/vuna/docs/architecture) — the crate map, data flow, the KOTVA seam,
  and the storage-math arithmetic behind "index, not pages."
- [**Viability**](/products/vuna/docs/viability) — the honest risk assessment: what this design
  solves, what it doesn't, and what could kill it.
- [**Roadmap**](/products/vuna/docs/roadmap) — what's next, in order, and the discipline that has
  to be re-chosen every time growth looks tempting.
- [**FAQ**](/products/vuna/docs/faq).
