# Roadmap

Vuna is a v0 scaffold. This page is ordered by what has to be true before the next thing is worth
building — not a marketing timeline.

## Now: wire the stage crates to each other

`vuna-core` is frozen and tested, and the four stage crates behind it are each implemented and
unit-tested on their own:

- `vuna-frontier`: a redb-backed persistent frontier, URL canonicalization, and DHT crawl-assignment
  behind a trait.
- `vuna-crawl`: a polite fetcher — reqwest, robots.txt, per-host rate limiting, hard body cap.
- `vuna-extract`: the `web` extractor (chunking + link extraction), the `retail` extractor, and the
  interpreter for declarative `adapters/*.toml` manifests — so growing site coverage is a reviewed
  data change, not a new Rust impl.
- `vuna-index`: tantivy keyword index + per-space HNSW vectors + the link/knowledge graph.

What does **not** exist is the wiring. Nothing yet runs crawl → extract → index end to end, because
that loop lives in `vuna-node`, which is still a stub. The specific pieces still missing behind the
seam traits are the real libp2p/Kademlia assignment binding and K× replication of `UrlEntry`,
optional headless rendering for pages that need it, and a real transformer embedder in place of
`vuna-index`'s hashing stand-in.

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
