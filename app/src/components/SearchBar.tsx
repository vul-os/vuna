import type { Ref } from "react";

interface Props {
  value: string;
  onChange: (v: string) => void;
  onSubmit: () => void;
  loading: boolean;
  inputRef?: Ref<HTMLInputElement>;
}

export function SearchBar({ value, onChange, onSubmit, loading, inputRef }: Props) {
  return (
    <form
      className="search-form"
      onSubmit={(e) => {
        e.preventDefault();
        onSubmit();
      }}
      role="search"
    >
      <svg
        className="search-glyph"
        width="19"
        height="19"
        viewBox="0 0 24 24"
        fill="none"
        aria-hidden="true"
      >
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
        autoFocus
        placeholder="Search the harvested index…"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        aria-label="Search query"
      />
      {value ? (
        <button
          type="button"
          className="search-clear"
          onClick={() => onChange("")}
          aria-label="Clear query"
        >
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path
              d="M6 6l12 12M18 6 6 18"
              stroke="currentColor"
              strokeWidth="2.2"
              strokeLinecap="round"
            />
          </svg>
        </button>
      ) : (
        <kbd className="search-kbd" aria-hidden="true">
          /
        </kbd>
      )}
      <button type="submit" className="search-submit" disabled={loading}>
        {loading ? <span className="search-spinner" aria-hidden="true" /> : "Reap"}
        <span className="sr">{loading ? "Searching" : "Run search"}</span>
      </button>
    </form>
  );
}
