import { useEffect, useRef, useState } from "react";
import { nodeStatus, search, stats as fetchStats } from "./lib/api";
import type { NodeStatus, RankedHit, Stats } from "./types";
import { Wordmark } from "./components/Logo";
import { ThemeToggle } from "./components/ThemeToggle";
import { MockBanner } from "./components/MockBanner";
import { SearchBar } from "./components/SearchBar";
import { ResultsList } from "./components/ResultsList";
import { NodeDashboard } from "./components/NodeDashboard";

const DEBOUNCE_MS = 260;

export default function App() {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<RankedHit[]>([]);
  const [loading, setLoading] = useState(true);
  const [searched, setSearched] = useState(false);

  const [status, setStatus] = useState<NodeStatus | null>(null);
  const [nodeStats, setNodeStats] = useState<Stats | null>(null);

  const debounceRef = useRef<number | undefined>(undefined);
  const requestSeq = useRef(0);
  const searchRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    nodeStatus().then(setStatus).catch(() => void 0);
    fetchStats().then(setNodeStats).catch(() => void 0);
  }, []);

  function runSearch(q: string) {
    const seq = ++requestSeq.current;
    setLoading(true);
    search(q)
      .then((hits) => {
        if (seq !== requestSeq.current) return; // a newer query superseded this one
        setResults(hits);
        setSearched(true);
      })
      .finally(() => {
        if (seq === requestSeq.current) setLoading(false);
      });
  }

  // initial load: show the local shard's trending/default set immediately
  useEffect(() => {
    runSearch("");
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // "/" focuses the query box from anywhere, the way every search tool does.
  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      const el = document.activeElement;
      const typing = el instanceof HTMLInputElement || el instanceof HTMLTextAreaElement;
      if (e.key === "/" && !typing && !e.metaKey && !e.ctrlKey && !e.altKey) {
        e.preventDefault();
        searchRef.current?.focus();
        searchRef.current?.select();
      }
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  function onChange(v: string) {
    setQuery(v);
    window.clearTimeout(debounceRef.current);
    debounceRef.current = window.setTimeout(() => runSearch(v), DEBOUNCE_MS);
  }

  function onSubmit() {
    window.clearTimeout(debounceRef.current);
    runSearch(query);
  }

  const localCount = results.filter((r) => r.source === "local").length;
  const peerCount = results.filter((r) => r.source === "peer").length;
  const indexerCount = results.filter((r) => r.source === "indexer").length;

  return (
    <div className="app-shell">
      <header className="app-header">
        <Wordmark />
        <div className="header-right">
          <span className="header-status" title="Peers this node is currently connected to">
            <span
              className={`online-dot-small${status?.online ? " is-online" : ""}`}
              aria-hidden="true"
            />
            {status ? `${status.peers_connected} peers` : "connecting…"}
          </span>
          <ThemeToggle />
        </div>
      </header>

      <MockBanner />

      <main className="layout">
        <section className="main-col">
          <div className="hero">
            <h1 className="hero-title">What are you harvesting today?</h1>
            <SearchBar
              inputRef={searchRef}
              value={query}
              onChange={onChange}
              onSubmit={onSubmit}
              loading={loading}
            />

            <div className="reach">
              <span className="reach-line">
                Searching your local shard
                {status && status.peers_connected > 0 ? ` + ${status.peers_connected} peers` : ""} —
                index only, pages stay where they live.
              </span>
              {searched && !loading && results.length > 0 && (
                <span className="reach-counts">
                  <span className="rc rc--local">{localCount} local</span>
                  <span className="rc rc--peer">{peerCount} peer</span>
                  <span className="rc rc--indexer">{indexerCount} indexer</span>
                </span>
              )}
            </div>
          </div>

          <p className="sr" role="status" aria-live="polite">
            {loading
              ? "Searching."
              : searched
                ? `${results.length} results: ${localCount} local, ${peerCount} peer, ${indexerCount} indexer. All mock data.`
                : ""}
          </p>

          <ResultsList results={results} query={query} loading={loading} searched={searched} />
        </section>

        <NodeDashboard status={status} stats={nodeStats} />
      </main>

      <footer className="app-footer">
        <p>
          The index is <strong>derived and rebuildable</strong>, never authoritative — a stale
          shard is a staleness problem, not a correctness one. No pages are archived, no token,
          no new crypto.
        </p>
      </footer>
    </div>
  );
}
