import { MOCK_ORIGIN } from "../lib/api";

/**
 * Permanent, non-dismissible disclosure that nothing on screen is real.
 *
 * This is not decoration and it is not a placeholder. `vuna-query` and `vuna-node` are
 * Wave-2 stubs, so every figure the app renders — result scores, peer counts, documents
 * indexed, storage used — is a fixture. It stays on screen until a real node is answering,
 * and it is deliberately not dismissible: a screenshot of this UI without this banner
 * would be a false claim.
 */
export function MockBanner() {
  return (
    <div className="mock-banner" role="status">
      <span className="mock-banner-tag">mock data</span>
      <p>
        <strong>No live node is answering.</strong> <code>vuna-query</code> and{" "}
        <code>vuna-node</code> are Wave-2 stubs, so every result, count and byte figure below
        is a fixture from <code>{MOCK_ORIGIN}</code>. The engine crates it will eventually
        talk to are real and tested; this end-to-end path is not built yet.
      </p>
    </div>
  );
}
