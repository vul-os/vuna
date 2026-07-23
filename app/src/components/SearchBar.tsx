import { useEffect, useRef } from "react";

interface Props {
  value: string;
  onChange: (v: string) => void;
  onSubmit: () => void;
  loading: boolean;
}

export function SearchBar({ value, onChange, onSubmit, loading }: Props) {
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  return (
    <form
      className="search-form"
      onSubmit={(e) => {
        e.preventDefault();
        onSubmit();
      }}
      role="search"
    >
      <svg className="search-glyph" width="19" height="19" viewBox="0 0 24 24" fill="none" aria-hidden="true">
        <circle cx="11" cy="11" r="7" stroke="currentColor" strokeWidth="2" />
        <path d="M20.5 20.5 16 16" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
      </svg>
      <input
        ref={inputRef}
        className="search-input"
        type="text"
        inputMode="search"
        autoComplete="off"
        spellCheck={false}
        placeholder="Search the harvested index…"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        aria-label="Search query"
      />
      {value && (
        <button
          type="button"
          className="search-clear"
          onClick={() => onChange("")}
          aria-label="Clear query"
        >
          ×
        </button>
      )}
      <button type="submit" className="search-submit" disabled={loading}>
        {loading ? <span className="search-spinner" aria-hidden="true" /> : "Reap"}
      </button>
    </form>
  );
}
