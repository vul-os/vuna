/**
 * The vuna mark — an ear of grain whose grains are index nodes. Same geometry as
 * `brand/mark.svg`; keep the two in step. The stem takes `currentColor` and the grains
 * follow `--grain`, so the mark themes itself.
 */
export function LogoMark({ size = 30 }: { size?: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 256 256"
      className="logo-mark"
      fill="none"
      aria-hidden="true"
    >
      <g stroke="currentColor" strokeLinecap="round">
        <path d="M128 226V58" strokeWidth="14" />
        <g strokeWidth="9">
          <path d="M128 98 94 74" />
          <path d="M128 98l34-24" />
          <path d="M128 130 86 106" />
          <path d="M128 130l42-24" />
          <path d="M128 162 84 138" />
          <path d="M128 162l44-24" />
          <path d="M128 194 86 170" />
          <path d="M128 194l42-24" />
          <path d="M128 222 94 202" />
          <path d="M128 222l34-20" />
        </g>
      </g>
      <g className="logo-grain">
        <circle cx="128" cy="44" r="16" />
        <circle cx="94" cy="74" r="15" />
        <circle cx="162" cy="74" r="15" />
        <circle cx="86" cy="106" r="15" />
        <circle cx="170" cy="106" r="15" />
        <circle cx="84" cy="138" r="15" />
        <circle cx="172" cy="138" r="15" />
        <circle cx="86" cy="170" r="15" />
        <circle cx="170" cy="170" r="15" />
        <circle cx="94" cy="202" r="15" />
        <circle cx="162" cy="202" r="15" />
      </g>
    </svg>
  );
}

export function Wordmark() {
  return (
    <div className="wordmark">
      <LogoMark size={28} />
      <div className="wordmark-text">
        <span className="wordmark-title">vuna</span>
        <span className="wordmark-tagline">reap the open web</span>
      </div>
    </div>
  );
}
