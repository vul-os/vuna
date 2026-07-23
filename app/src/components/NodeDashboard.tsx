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
        <h2>Your node</h2>
        <span className={`online-pill${status?.online ? " is-online" : ""}`}>
          <span className="online-dot" aria-hidden="true" />
          {loading ? "connecting…" : status.online ? `online · ${formatUptime(status.uptime_secs)}` : "offline"}
        </span>
      </div>

      <div className="stat-grid">
        <StatTile label="docs indexed" value={loading ? "—" : formatInt(stats!.docs_indexed)} accent="rust" />
        <StatTile label="storage used" value={loading ? "—" : formatStorage(stats!.storage_mb)} accent="gold" />
        <StatTile label="peers" value={loading ? "—" : formatInt(stats!.peers)} accent="olive" />
        <StatTile label="spaces served" value={loading ? "—" : String(stats!.spaces_served)} />
      </div>

      <div className="dashboard-section">
        <h3>Extractors</h3>
        <div className="chip-row">
          {loading
            ? null
            : status!.extractors.map((ex) => (
                <span key={ex.kind} className={`chip${ex.enabled ? " chip--on" : " chip--off"}`}>
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
                <li key={l.id}>
                  <span className="list-label">{l.label}</span>
                  <span className="list-count">{formatInt(l.url_count)}</span>
                </li>
              ))}
        </ul>
      </div>

      <p className="visibility-note">
        {loading ? (
          " "
        ) : status!.query_visibility === "terminating" ? (
          <>
            <strong>Query visibility: terminating.</strong> This node's operator can read queries in
            the clear — the honest v0 default. No hidden telemetry.
          </>
        ) : (
          <>
            <strong>Query visibility: blind.</strong> Queries are shielded from this node's operator.
          </>
        )}
      </p>
    </aside>
  );
}
