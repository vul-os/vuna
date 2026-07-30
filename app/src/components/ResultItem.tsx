import type { RankedHit } from "../types";
import { hostOf, highlightTerms } from "../lib/format";
import { SourceTag } from "./SourceTag";

export function ResultItem({ hit, query, index }: { hit: RankedHit; query: string; index: number }) {
  const pieces = highlightTerms(hit.hit.snippet, query);
  const gaugePct = Math.max(4, Math.min(100, hit.rank * 100));

  return (
    <li className="result-item" style={{ animationDelay: `${Math.min(index, 8) * 45}ms` }}>
      <span className="result-rank" aria-hidden="true">
        {index + 1}
      </span>
      <div className="result-body">
        <div className="result-top">
          <SourceTag source={hit.source} />
          <span className="result-host">{hostOf(hit.hit.url)}</span>
        </div>
        <a className="result-title" href={hit.hit.url} target="_blank" rel="noreferrer">
          {hit.hit.title ?? hit.hit.url}
        </a>
        <div className="result-url">{hit.hit.url}</div>
        <p className="result-snippet">
          {pieces.map((p, i) => (p.match ? <mark key={i}>{p.text}</mark> : <span key={i}>{p.text}</span>))}
        </p>
        <div className="result-foot">
          {/* Two different numbers, deliberately labelled: `rank` is this node's merged
              rank contribution, `score` is the index's relevance score for the hit. */}
          <span className="result-metric" title="This node's merged rank contribution for the hit">
            <span className="result-metric-label">rank</span>
            <span className="result-gauge" aria-hidden="true">
              <span className="result-gauge-fill" style={{ width: `${gaugePct}%` }} />
            </span>
          </span>
          <span className="result-metric" title="Relevance score reported by the index">
            <span className="result-metric-label">score</span>
            <span className="result-score">{hit.hit.score.toFixed(1)}</span>
          </span>
        </div>
      </div>
    </li>
  );
}
