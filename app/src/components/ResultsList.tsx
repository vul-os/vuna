import type { RankedHit } from "../types";
import { ResultItem } from "./ResultItem";
import { LogoMark } from "./Logo";

interface Props {
  results: RankedHit[];
  query: string;
  loading: boolean;
  searched: boolean;
}

export function ResultsList({ results, query, loading, searched }: Props) {
  if (loading && results.length === 0) {
    return (
      <ul className="results-list" aria-busy="true">
        {[0, 1, 2].map((i) => (
          <li key={i} className="result-skeleton" style={{ animationDelay: `${i * 90}ms` }}>
            <div className="skel skel-tag" />
            <div className="skel skel-title" />
            <div className="skel skel-line" />
            <div className="skel skel-line short" />
          </li>
        ))}
      </ul>
    );
  }

  if (searched && results.length === 0) {
    return (
      <div className="state-empty">
        <LogoMark size={40} />
        <p>No harvest for that query yet. Try broadening it, or subscribe to another list.</p>
      </div>
    );
  }

  if (!searched) {
    return (
      <div className="state-empty">
        <LogoMark size={40} />
        <p>Sow a query above to begin the harvest — showing a sample of the local shard.</p>
      </div>
    );
  }

  return (
    <ul className="results-list">
      {results.map((r, i) => (
        <ResultItem key={r.hit.url} hit={r} query={query} index={i} />
      ))}
    </ul>
  );
}
