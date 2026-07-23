import { useEffect, useState } from "react";

type Theme = "light" | "dark";

function systemTheme(): Theme {
  return window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

export function ThemeToggle() {
  const [override, setOverride] = useState<Theme | null>(() => {
    const stored = localStorage.getItem("vuna-theme");
    return stored === "light" || stored === "dark" ? stored : null;
  });

  useEffect(() => {
    const root = document.documentElement;
    if (override) root.setAttribute("data-theme", override);
    else root.removeAttribute("data-theme");
  }, [override]);

  const effective = override ?? systemTheme();

  function toggle() {
    setOverride(effective === "dark" ? "light" : "dark");
    localStorage.setItem("vuna-theme", effective === "dark" ? "light" : "dark");
  }

  return (
    <button
      type="button"
      className="theme-toggle"
      onClick={toggle}
      aria-label={effective === "dark" ? "Switch to light theme" : "Switch to dark theme"}
      title={effective === "dark" ? "Switch to daylight" : "Switch to lamplight"}
    >
      {effective === "dark" ? (
        <svg width="17" height="17" viewBox="0 0 24 24" fill="none" aria-hidden="true">
          <circle cx="12" cy="12" r="4.6" fill="currentColor" />
          <g stroke="currentColor" strokeWidth="1.8" strokeLinecap="round">
            <path d="M12 2.5v2.6M12 18.9v2.6M21.5 12h-2.6M5.1 12H2.5M18.4 5.6l-1.8 1.8M7.4 16.6l-1.8 1.8M18.4 18.4l-1.8-1.8M7.4 7.4 5.6 5.6" />
          </g>
        </svg>
      ) : (
        <svg width="17" height="17" viewBox="0 0 24 24" fill="none" aria-hidden="true">
          <path
            d="M20.5 14.6A8.6 8.6 0 1 1 9.4 3.5a7 7 0 0 0 11.1 11.1Z"
            fill="currentColor"
          />
        </svg>
      )}
    </button>
  );
}
