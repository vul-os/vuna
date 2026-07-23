import { useEffect, useRef, useState } from "react";
import { nodeStatus, search, stats as fetchStats } from "./lib/api";
import type { NodeStatus, RankedHit, Stats } from "./types";
import { Wordmark } from "./components/Logo";
import { ThemeToggle } from "./components/ThemeToggle";
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

  function onChange(v: string) {
    setQuery(v);
    window.clearTimeout(debounceRef.current);
    debounceRef.current = window.setTimeout(() => runSearch(v), DEBOUNCE_MS);
  }

  function onSubmit() {
    window.clearTimeout(debounceRef.current);
    runSearch(query);
  }

  const reachLine = status
    ? `Searching your local shard${status.peers_connected > 0 ? ` + ${status.peers_connected} peers` : ""}`
    : "Searching your local shard";

  return (
    <div className="app-shell">
      <header className="app-header">
        <Wordmark />
        <div className="header-right">
          <span className="header-status">
            <span className={`online-dot-small${status?.online ? " is-online" : ""}`} aria-hidden="true" />
            {status ? `${status.peers_connected} peers` : "connecting…"}
          </span>
          <ThemeToggle />
        </div>
      </header>

      <main className="layout">
        <section className="main-col">
          <div className="hero">
            <h1 className="hero-title">What are you harvesting today?</h1>
            <SearchBar value={query} onChange={onChange} onSubmit={onSubmit} loading={loading} />
            <p className="hero-meta">{reachLine} — index only, pages stay where they live.</p>
          </div>

          <ResultsList results={results} query={query} loading={loading} searched={searched} />
        </section>

        <NodeDashboard status={status} stats={nodeStats} />
      </main>

      <footer className="app-footer">
        <p>
          The index is <strong>derived and rebuildable</strong>, never authoritative — a stale shard
          is a staleness problem, not a correctness one. No pages are archived, no token, no new
          crypto.
        </p>
      </footer>
    </div>
  );
}
