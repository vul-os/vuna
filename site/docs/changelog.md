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
- Stub crates scaffolded for `vuna-frontier`, `vuna-crawl`, `vuna-extract`, `vuna-index`,
  `vuna-query`, `vuna-node`, each with a module doc comment describing its designed role.
- `app/` scaffolded: a Tauri + React desktop shell (search bar + node dashboard), builds and runs
  against an in-process mock corpus. No live daemon behind it yet.
- `adapters/` established for declarative, TOML-based retail-vertical site adapters.
- Design (`01-design.md`), architecture (`02-architecture.md`), and an unusually candid viability
  assessment (`00-viability.md`) written and kept in sync with what actually compiles.
- 63 tests passing across the workspace, concentrated in `vuna-core` and the crates furthest along.

Nothing has shipped beyond this scaffold. There is no tagged release, no built installer, and no
live network — see [Getting started](/products/vuna/docs/getting-started) for the honest,
current-as-of-today status.
