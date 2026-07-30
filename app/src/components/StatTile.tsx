interface Props {
  label: string;
  value: string;
  /** Which token tints the tile's left rule. `grain` is reserved for "yours". */
  accent?: "grain" | "peer" | "running";
  /** Optional second line — a unit, a qualifier, or what the number is derived from. */
  note?: string;
}

export function StatTile({ label, value, accent, note }: Props) {
  return (
    <div className={`stat-tile${accent ? ` stat-tile--${accent}` : ""}`}>
      <span className="stat-value">{value}</span>
      <span className="stat-label">{label}</span>
      {note && <span className="stat-note">{note}</span>}
    </div>
  );
}
