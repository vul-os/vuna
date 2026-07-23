export function formatInt(n: number): string {
  return new Intl.NumberFormat("en-US").format(Math.round(n));
}

export function formatStorage(mb: number): string {
  if (mb >= 1024) return `${(mb / 1024).toFixed(1)} GB`;
  return `${mb.toFixed(0)} MB`;
}

export function formatUptime(secs: number): string {
  const h = Math.floor(secs / 3600);
  const m = Math.floor((secs % 3600) / 60);
  if (h > 0) return `${h}h ${m}m`;
  return `${m}m`;
}

export function hostOf(url: string): string {
  try {
    return new URL(url).hostname.replace(/^www\./, "");
  } catch {
    return url;
  }
}

/** Splits `text` into plain/matched pieces (case-insensitive) so the UI can bold query terms. */
export function highlightTerms(text: string, query: string): { text: string; match: boolean }[] {
  const rawTerms = query
    .trim()
    .split(/\s+/)
    .filter((t) => t.length >= 2);
  if (rawTerms.length === 0) return [{ text, match: false }];

  const escaped = rawTerms.map((t) => t.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"));
  const lowerTerms = new Set(rawTerms.map((t) => t.toLowerCase()));

  const parts = text.split(new RegExp(`(${escaped.join("|")})`, "gi"));
  return parts.filter((p) => p.length > 0).map((p) => ({ text: p, match: lowerTerms.has(p.toLowerCase()) }));
}
