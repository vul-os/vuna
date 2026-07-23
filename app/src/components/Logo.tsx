/** The wheat-sheaf mark, recreated as crisp inline SVG (the app icon is a PNG/ICNS render of the
 * same shape). Sized for header use; colors follow the theme via CSS variables. */
export function LogoMark({ size = 30 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 100 100" className="logo-mark" aria-hidden="true">
      <path d="M50 14 L58 46 L42 46 Z" className="logo-blade logo-blade--center" />
      <path d="M30 24 L48 50 L34 52 Z" className="logo-blade logo-blade--left" />
      <path d="M70 24 L52 50 L66 52 Z" className="logo-blade logo-blade--right" />
      <ellipse cx="50" cy="66" rx="22" ry="11" className="logo-band" />
      <path d="M31 62 A22 11 0 0 0 50 74" className="logo-band-hi" />
    </svg>
  );
}

export function Wordmark() {
  return (
    <div className="wordmark">
      <LogoMark />
      <div className="wordmark-text">
        <span className="wordmark-title">Vuna</span>
        <span className="wordmark-tagline">reap the open web</span>
      </div>
    </div>
  );
}
