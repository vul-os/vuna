import type { RankedHit } from "../types";
import { hostOf, highlightTerms } from "../lib/format";
import { SourceTag } from "./SourceTag";

export function ResultItem({ hit, query, index }: { hit: RankedHit; query: string; index: number }) {
  const pieces = highlightTerms(hit.hit.snippet, query);
  const gaugePct = Math.max(4, Math.min(100, hit.rank * 100));

  return (
    <li className="result-item" style={{ animationDelay: `${Math.min(index, 8) * 45}ms` }}>
      <div className="result-top">
        <SourceTag source={hit.source} />
        <span className="result-host">{hostOf(hit.hit.url)}</span>
      </div>
      <a className="result-title" href={hit.hit.url} target="_blank" rel="noreferrer">
        {hit.hit.title ?? hit.hit.url}
      </a>
      <div className="result-url">{hit.hit.url}</div>
      <p className="result-snippet">
        {pieces.map((p, i) =>
          p.match ? (
            <mark key={i}>{p.text}</mark>
          ) : (
            <span key={i}>{p.text}</span>
          )
        )}
      </p>
      <div className="result-foot">
        <div className="result-gauge" aria-hidden="true">
          <div className="result-gauge-fill" style={{ width: `${gaugePct}%` }} />
        </div>
        <span className="result-score">score {hit.hit.score.toFixed(1)}</span>
      </div>
    </li>
  );
}
