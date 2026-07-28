# Changelog

## v0.0.1 — v0 scaffold

The initial workspace. This is the current state — there has been no release beyond this.

- `vuna-core` established as the frozen contract: shared types (`UrlEntry`, `IndexedDoc`,
  `EmbeddingSpace`, `NodeDescriptor`, `RetailObservation`, …) and the seam traits every other crate
  builds against (`Frontier`, `Fetcher`, `Extractor`, `Index`, `Embedder`, and the three KOTVA seam
  traits: `NodeIdentity`, `Publisher`, `ContentAddresser`). Compiles offline, unit-tested.
- `quorum::reconcile` implemented and tested — the retail vertical's distinct-identity,
  plurality/median reconciliation rule described in
  [Architecture](/products/vuna/docs/architecture).
- `vuna-frontier`, `vuna-crawl`, `vuna-extract`, and `vuna-index` implemented and unit-tested
  behind their seam traits — individually, not yet wired to each other. `vuna-query` and
  `vuna-node` remain stubs, so nothing runs crawl → extract → index → query end to end.
- `app/` scaffolded: a Tauri + React desktop shell (search bar + node dashboard), builds and runs
  against an in-process mock corpus. No live daemon behind it yet.
- `adapters/` established for declarative, TOML-based retail-vertical site adapters, together with
  the interpreter in `vuna-extract` that runs them as `Extractor`s. Three worked manifests
  (Shopify, generic schema.org JSON-LD, WooCommerce Store API), each exercised against realistic
  payloads by a corpus test that fails if a manifest is added without a fixture.
- Design (`01-design.md`), architecture (`02-architecture.md`), and an unusually candid viability
  assessment (`00-viability.md`) written and kept in sync with what actually compiles.
- 132 tests passing across the workspace, concentrated in `vuna-core` and the crates furthest
  along, run in CI on every push and pull request.

Nothing has shipped beyond this scaffold. There is no tagged release, no built installer, and no
live network — see [Getting started](/products/vuna/docs/getting-started) for the honest,
current-as-of-today status.
