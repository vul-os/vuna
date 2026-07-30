import { useEffect, useState } from "react";

type Theme = "light" | "dark";

/** Suite-wide key — magnetite's site writes light|dark|system to the same one. Anything that
 *  is not an explicit light/dark choice means "follow the OS", which the token layer already
 *  handles via `prefers-color-scheme`. `vuna-theme` is read once as a migration. */
const KEY = "vulos-theme";
const LEGACY_KEY = "vuna-theme";

function readStored(): Theme | null {
  try {
    const v = localStorage.getItem(KEY) ?? localStorage.getItem(LEGACY_KEY);
    return v === "light" || v === "dark" ? v : null;
  } catch {
    return null;
  }
}

function systemTheme(): Theme {
  return window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light";
}

export function ThemeToggle() {
  const [override, setOverride] = useState<Theme | null>(readStored);

  useEffect(() => {
    const root = document.documentElement;
    if (override) root.setAttribute("data-theme", override);
    else root.removeAttribute("data-theme");
  }, [override]);

  const effective = override ?? systemTheme();
  const next: Theme = effective === "dark" ? "light" : "dark";

  function toggle() {
    setOverride(next);
    try {
      localStorage.setItem(KEY, next);
    } catch {
      /* private mode — the choice just does not persist */
    }
  }

  return (
    <button
      type="button"
      className="theme-toggle"
      onClick={toggle}
      aria-label={next === "dark" ? "Switch to the dark theme" : "Switch to the light theme"}
      title={next === "dark" ? "Switch to lamplight" : "Switch to daylight"}
    >
      {effective === "dark" ? (
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
          <circle cx="12" cy="12" r="4.4" fill="currentColor" />
          <g stroke="currentColor" strokeWidth="1.8" strokeLinecap="round">
            <path d="M12 2.6v2.5M12 18.9v2.5M21.4 12h-2.5M5.1 12H2.6M18.3 5.7l-1.8 1.8M7.5 16.5l-1.8 1.8M18.3 18.3l-1.8-1.8M7.5 7.5 5.7 5.7" />
          </g>
        </svg>
      ) : (
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
          <path d="M20.5 14.6A8.6 8.6 0 1 1 9.4 3.5a7 7 0 0 0 11.1 11.1Z" fill="currentColor" />
        </svg>
      )}
    </button>
  );
}
