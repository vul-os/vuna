# FAQ

### Is Vuna ready to use?

No. It's a v0 scaffold. `vuna-core` (the contract) compiles and is unit-tested; the stage crates
(`vuna-frontier`, `vuna-crawl`, `vuna-extract`, `vuna-index`) are being built out; `vuna-query` and
`vuna-node` are stubs; the desktop app runs against a **mock corpus**, not a live daemon. There are
zero live nodes and zero users today. See [Getting started](/products/vuna/docs/getting-started)
for the precise state, and [Viability](/products/vuna/docs/viability) for the honest risk
assessment.

### What does "index, not pages" actually mean?

Vuna nodes never archive the pages they crawl. Each node keeps keyword postings, vector
embeddings, link/knowledge-graph edges, a short snippet, a small number of retained chunks of
text, and a pointer back to the live URL. The design's whole storage-math argument (see
[Architecture](/products/vuna/docs/architecture) §6) depends on this: it's what collapses a
billion-page index down to roughly 1.8 GB per node at 10,000 nodes, instead of the
hundreds-of-petabytes "reliable" storage tier a full open-web page archive needs.

### Does Vuna have a token?

No, and it's not planned. Vuna is designed to use KOTVA's existing identity/PUB/DHT/SEARCH
primitives instead of introducing new crypto or an incentive token — though that binding is a
Wave-2 stub and is not wired up yet. This is the honest, mission-aligned choice — but
it also means Vuna has no proven mechanism for motivating ongoing volunteer compute the way, say,
mining rewards do elsewhere. See Viability §1 for why that's a real open risk, not a footnote.

### What happens when a new, better embedding model comes out?

Nothing forces a fleet-wide migration. Each embedding model is its own independent, parallel
vector index — a "space" — that nodes opt into on their own schedule. Because the index retains
chunk *text* alongside vectors, adopting a new space is a local recompute over already-fetched
text, not a re-crawl. Old spaces keep working for as long as any node still serves them, because
the index is derived state with no authoritative status.

### What is the "retail radar" vertical?

An extractor that reads price/stock availability from a store's own structured data (schema.org
JSON-LD, OpenGraph) rather than a document to be ranked. Because a store's stock level is a claim
about a non-participant, not signed content, it's reconciled by quorum across multiple independent
observing nodes (median quantity, availability by plurality, accepted only above a distinct-
identity threshold `k`) instead of trusted from a single observer's fetch. See Architecture for
the exact reconciliation rule, and Viability §2 for why the Sybil-resistance question underneath
it is unresolved, not solved.

### How is Sybil resistance handled?

Only partially, and the docs say so directly. The reconciliation logic requires `k` distinct
identity keys to agree before accepting a claim — a floor against one key voting itself to quorum,
not a check that those keys are `k` different people. Anchoring identity to real humans is
delegated to KOTVA's proof-of-personhood primitive, which is real but imperfect (single-vendor
fragility, coverage gaps) and raises the cost of a Sybil identity without eliminating it. This is
the single largest open risk called out in [Viability](/products/vuna/docs/viability) §2.

### Will Vuna ever crawl the whole open web?

Not as a v1 promise, and the discipline against it is explicit: decentralized open-web search
inherits a permanent SEO adversary and an unsolved permissionless-trust problem at the same time.
Vuna's plan is bounded-corpus-first — federation content or one curated, vetted vertical — with
open-web crawl kept as a research track. Nothing in the code enforces this discipline; it has to
be re-chosen deliberately every time growth looks tempting.

### What license is it under?

MIT OR Apache-2.0, same as the rest of the vul-os ecosystem.
