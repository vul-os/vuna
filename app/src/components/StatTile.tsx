export function StatTile({ label, value, accent }: { label: string; value: string; accent?: "rust" | "gold" | "olive" }) {
  return (
    <div className={`stat-tile${accent ? ` stat-tile--${accent}` : ""}`}>
      <span className="stat-value">{value}</span>
      <span className="stat-label">{label}</span>
    </div>
  );
}
