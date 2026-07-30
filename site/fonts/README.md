# Vendored fonts

`site/index.html` and `site/docs.html` declare these with `@font-face` and local
`./fonts/*.woff2` paths. Nothing here is fetched at runtime — no CDN, no Google
Fonts, no `@import url(https://…)`. The `site/` directory deploys standalone.

The same three families, same files, are also vendored at `app/src/assets/fonts/`
for the desktop app, so the app and the site set type identically.

| Family | Files | Copyright | Licence |
|---|---|---|---|
| **Fraunces** | `fraunces-300-normal`, `fraunces-400-normal`, `fraunces-400-italic`, `fraunces-600-normal` | The Fraunces Project Authors — <https://github.com/undercasetype/Fraunces> | SIL OFL 1.1, `OFL.txt` |
| **Archivo** | `archivo-400-normal`, `archivo-500-normal`, `archivo-600-normal`, `archivo-700-normal` | The Archivo Project Authors — <https://github.com/Omnibus-Type/Archivo> | SIL OFL 1.1, `OFL.txt` |
| **IBM Plex Mono** | `plexmono-400-normal`, `plexmono-500-normal` | IBM Corp. — <https://github.com/IBM/plex> | SIL OFL 1.1, `OFL.txt` |

All three are latin subsets packaged by [Fontsource](https://fontsource.org). `OFL.txt`
is the full text of SIL Open Font License 1.1, which governs all three; per §5 of that
licence the reserved font names are unchanged and the files are unmodified.
