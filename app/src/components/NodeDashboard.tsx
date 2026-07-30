import type { NodeStatus, Stats } from "../types";
import { formatInt, formatStorage, formatUptime } from "../lib/format";
import { StatTile } from "./StatTile";

interface Props {
  status: NodeStatus | null;
  stats: Stats | null;
}

export function NodeDashboard({ status, stats }: Props) {
  const loading = !status || !stats;

  return (
    <aside className="panel dashboard" aria-label="Node dashboard">
      <div className="panel-header">
        <h2>
          Your node
          {/* Repeated here, not just in the top banner: this panel is the most
              authoritative-looking surface in the app and it is the one most likely to be
              cropped out of context into a screenshot. */}
          <span className="panel-mock" title="Every figure in this panel is a fixture — vuna-node is a Wave-2 stub">
            mock
          </span>
        </h2>
        <span className={`online-pill${status?.online ? " is-online" : ""}`}>
          <span className="online-dot" aria-hidden="true" />
          {loading
            ? "connecting…"
            : status.online
              ? `online · ${formatUptime(status.uptime_secs)}`
              : "offline"}
        </span>
      </div>

      <div className="node-id" title="This node's identity on the substrate">
        <span className="node-id-label">node</span>
        <code>{loading ? "—" : status.node_id}</code>
      </div>

      <div className="stat-grid">
        <StatTile
          label="docs indexed"
          value={loading ? "—" : formatInt(stats!.docs_indexed)}
          accent="grain"
        />
        <StatTile
          label="storage used"
          value={loading ? "—" : formatStorage(stats!.storage_mb)}
          accent="grain"
          note={loading ? undefined : "index only, no pages"}
        />
        <StatTile label="peers" value={loading ? "—" : formatInt(stats!.peers)} accent="peer" />
        <StatTile
          label="spaces served"
          value={loading ? "—" : String(stats!.spaces_served)}
          accent="running"
        />
      </div>

      {/* Derived-state totals. These come straight off `Stats` and were previously
          collected but never shown. */}
      <ul className="kv">
        <li>
          <span className="kv-k">keyword postings</span>
          <span className="kv-v">{loading ? "—" : formatInt(stats!.keyword_postings)}</span>
        </li>
        <li>
          <span className="kv-k">graph edges</span>
          <span className="kv-v">{loading ? "—" : formatInt(stats!.graph_edges)}</span>
        </li>
        <li>
          <span className="kv-k">lists subscribed</span>
          <span className="kv-v">{loading ? "—" : formatInt(stats!.lists_subscribed)}</span>
        </li>
      </ul>

      <div className="dashboard-section">
        <h3>Extractors</h3>
        <div className="chip-row">
          {loading
            ? null
            : status!.extractors.map((ex) => (
                <span
                  key={ex.kind}
                  className={`chip${ex.enabled ? " chip--on" : " chip--off"}`}
                  title={ex.enabled ? `${ex.label} — running` : `${ex.label} — not enabled`}
                >
                  <span className="chip-dot" aria-hidden="true" />
                  {ex.label}
                </span>
              ))}
        </div>
      </div>

      <div className="dashboard-section">
        <h3>Embedding spaces served</h3>
        <div className="chip-row">
          {loading
            ? null
            : status!.served_spaces.map((sp) => (
                <span key={sp.id} className="chip chip--space" title={sp.id}>
                  {sp.label}
                  <span className="chip-dim">{sp.dim}d</span>
                  {sp.default && <span className="chip-badge">default</span>}
                </span>
              ))}
        </div>
      </div>

      <div className="dashboard-section">
        <h3>URL lists subscribed</h3>
        <ul className="list-rows">
          {loading
            ? null
            : status!.subscribed_lists.map((l) => (
                <li key={l.id} title={l.id}>
                  <span className="list-label">{l.label}</span>
                  <span className="list-count">{formatInt(l.url_count)}</span>
                </li>
              ))}
        </ul>
      </div>

      <p className="visibility-note">
        {loading ? (
          " "
        ) : status!.query_visibility === "terminating" ? (
          <>
            <strong>Query visibility: terminating.</strong> This node's operator can read
            queries in the clear — the honest v0 default. No hidden telemetry.
          </>
        ) : (
          <>
            <strong>Query visibility: blind.</strong> Queries are shielded from this node's
            operator.
          </>
        )}
      </p>
    </aside>
  );
}
