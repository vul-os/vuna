# Vuna — Design: index-not-pages, multi-model, pluggable verticals

> **Name:** Vuna (Swahili *to harvest / reap*). Search-RAG and retail-radar were earlier considered
> as separate projects (Scout / Rada); they are folded here as **verticals of one engine**.
> **Status:** design. Supersedes the open-web thinking in `00-discussion-history.md`.

---

## The shape in one paragraph

A decentralized **index + RAG + graph** over a corpus of URLs, built on the **Kotva** substrate.
The network stores the **index, not the pages** — keyword postings, vector embeddings, a
link/knowledge graph, and a pointer back to the live URL. **Multiple embedding models coexist**
as parallel indexes; **nodes choose** which to serve; **defaults** ship so it works out of the
box; a **distributed, subscribable list of URLs** is the crawl frontier. No token, no page
archive, no new crypto — it rides Kotva identity / PUB / DHT / SEARCH.

---

## Four objects — that's the whole system

1. **URL list (the frontier)** — a signed, versioned, **subscribable** list of URLs to crawl
   (the ad-filter-list model, à la uBlock / Mwmbl). Distributed and deduped across nodes by
   URL-hash (DHT). Anyone can publish one; you choose which to trust — no central frontier authority.
2. **Embedding space** — a recognized model index: `(model_id, dim, quantization)`. Announced as
   a signed object. A small curated set are the **defaults**.
3. **Index shard** — derived per-node state, content-addressed and rebuildable: keyword postings +
   **vectors, one set per embedding space the node serves** + graph edges + metadata/snippet +
   **live-URL pointer** + **stored chunk text**.
4. **Node descriptor** — declares which URL **lists** it subscribes to and which embedding
   **spaces** it serves.

Reads use the Kotva **SEARCH** query fan-out/merge. The index is **derived, never authoritative**
(SRCH-2): on any disagreement, the author's signed content governs.

---

## Multiple embedding models — how lock-in dies

The one sharp future-proofing risk was: *new embedding model → forced fleet-wide re-embed.* This
design removes it.

- Each embedding space is an **independent, parallel index** over the same corpus. The keyword
  index and the graph are **shared and model-agnostic** — they never need re-doing.
- **Nodes choose** their spaces (in the descriptor). **Defaults:** the client ships subscribed to
  1–2 default spaces (a small multilingual model) so a fresh node is useful immediately.
- **New model → new space, announced → nodes opt in → its index builds incrementally.** There is
  **no forced re-embed**: adoption is voluntary and parallel; an old space lives as long as any
  node still serves it (it's derived state, so nothing breaks when one is dropped).
- Because each shard stores **chunk text**, adopting a new model is a **local recompute**, not a
  re-crawl. This is the whole reason multi-model is cheap.
- A query **names its space** (or uses the default); results merge only within a space.

*Net: the network accrues indexes for many models over time, side by side, and drifts toward newer
ones by opt-in rather than by a coordinated flag-day.*

---

## Distributed URL lists — the frontier

- The frontier is a **shared, distributed set of URLs**, partitioned by URL-hash across nodes
  (DHT), deduped, each entry carrying metadata: `last_crawled`, `last_embedded[space]`,
  `content_hash`.
- **Seed / default lists** are curated, signed, and **subscribable** like ad-filter lists. Nodes
  subscribe; **assignment** (which node crawls which slice) is DHT-derived — the same pattern the
  observation network uses for store-assignment.
- **Nodes keep it all, redundantly:** each entry replicates to K nodes near its key, so the
  frontier survives churn without any central list server.

---

## Roadmap — organic frontier growth (opt-in app)

- A browser app / extension that contributes **URLs the user actually visits** to the frontier —
  human-curated growth. This is **Mwmbl's proven Firefox-extension model**, the one live template
  the research endorsed.
- **Privacy-first and simple:** it contributes the **URL** (a public fact), never a
  browsing-history-tied-to-identity trail; local allow/deny filtering; **strictly opt-in, off by
  default**.
- **Roadmap, not v1.** v1 seeds from curated lists + volunteer crawlers; the visit-tracking app is
  a later reach that grows the frontier once the core works.

---

## Keep-it-simple guardrails

- **No page archive. No token. No new crypto.** Rides Kotva identity / PUB / DHT / SEARCH.
- **Four objects, one query path.** A node that serves the **default space** + subscribes to **one
  list** is already a complete, useful participant.
- **Bounded corpus first** (federation content or a curated vertical), where crawl is cheap and
  Sybil is dodged via Kotva's vetted operators. **Open web is a later research track**, never a v1
  promise.
- Everything derived is **rebuildable** — a stale or lost shard is a staleness problem, never a
  correctness one.
