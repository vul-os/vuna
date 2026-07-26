# Roadmap

Vuna is a v0 scaffold. This page is ordered by what has to be true before the next thing is worth
building — not a marketing timeline.

## Now: wire the stage crates

`vuna-core` is frozen and tested. The next work is making `vuna-frontier`, `vuna-crawl`,
`vuna-extract`, and `vuna-index` real against each other — the crate map already fixes the seam
traits (`Frontier`, `Fetcher`, `Extractor`, `Index`, `Embedder`), so this is filling in
implementations behind traits that already compile, not redesigning the shape.

- `vuna-frontier`: real DHT-assignment and K× replication of `UrlEntry`, not just the in-memory
  contract.
- `vuna-crawl`: a polite fetcher — reqwest + optional headless rendering, robots.txt, per-host rate
  limiting.
- `vuna-extract`: the `web` extractor (chunking + link extraction) and the first `retail` adapters
  (declarative `adapters/*.toml`, so growing site coverage is a data change, not a new Rust impl).
- `vuna-index`: tantivy keyword index + per-space HNSW vectors + the link/knowledge graph.

## Next: `vuna-query` and `vuna-node`

`vuna-query` implements the KOTVA SEARCH read path (local-first, optional peer/indexer fan-out,
Min-PPR merge) once there's a real index to query. `vuna-node` is the daemon that binds
`kotva-core` (pinned by tag) behind the three seam traits — it's deliberately the *last* piece to
depend on the real substrate, so every other crate keeps compiling and testing offline the whole
way there.

## Then: bounded-corpus launch

Per the [Viability](/products/vuna/docs/viability) assessment, Vuna's discipline is a **bounded
corpus first** — federation content or one curated vertical, with pre-vetted participants that
sidestep the worst of the small-network Sybil problem — before anything resembling open-web crawl.
Launching one vertical live, well, beats launching every vertical badly.

## Later, explicitly roadmap (not v1)

- **Organic frontier growth via an opt-in browser extension** — contributing URLs a user actually
  visits to the frontier, following Mwmbl's proven Firefox-extension model. Privacy-first: it
  contributes the URL itself (a public fact), never a browsing-history-tied-to-identity trail;
  local allow/deny filtering; strictly opt-in, off by default.
- **Additional embedding spaces** beyond the 1–2 defaults a fresh node ships subscribed to —
  adopted opt-in, node by node, with no fleet-wide re-embed, because chunk text is retained
  alongside vectors.
- **Additional verticals** beyond `web` and `retail` — the extractor seam is designed so a new
  vertical is a new `Extractor` impl over the same crawl/frontier/index/query stack.

## What is explicitly *not* planned

- **No token.** Ruled out by design, even though it's the mechanism that has reliably motivated
  volunteer compute elsewhere — see Viability §1 for why that's a real, unresolved cost of the
  choice, not a footnote.
- **No page archive, ever.** The entire storage-math argument depends on never storing full pages.
- **No open-web crawl as a v1 promise.** It remains a research track behind the bounded-corpus
  discipline, explicitly re-chosen rather than assumed, every time growth looks tempting.
