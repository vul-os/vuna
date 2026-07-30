# vuna brand

The mark, the palette and the type scheme, plus the rules that keep them honest.
`tokens.css` in this directory is the **single authority for colour and type**; the
desktop app imports it and the landing page restates it inline (it has to stay one
self-contained file). Everything below is measured, not asserted — the contrast
tables are real WCAG 2.1 ratios and the method for re-deriving them is at the end.

## Inventory

| File | What it is |
|---|---|
| `mark.svg` | The mark on its own. Stem is `currentColor`; grains follow `--vuna-grain`. Inline this. |
| `logo.svg` | The app tile — dark rounded-rect ground + the mark. For product grids and the README hero. |
| `favicon.svg` | Thickened five-grain reduction for 16–32px. Flips with `prefers-color-scheme` internally. |
| `tokens.css` | Colour + type + shape tokens for both themes. Imported by `app/src/index.css`. |
| `logo-1024.png` | Raster export of the tile, for tools that cannot take an SVG. |

`site/vuna-icon.svg` and `app/public/vuna-icon.svg` are copies of `favicon.svg`, because
`site/` has to deploy as a standalone directory and Vite serves the app's favicon from
`public/`. Keep all three in sync.

**Still on the old mark:** the bundled desktop icons under `app/src-tauri/icons/`
(`icon.icns`, `icon.ico`, `128x128.png`, …) are build artefacts and were not regenerated
here, because doing it properly needs the packaging toolchain rather than a hand-written
file. Regenerate them when the app is next packaged:

```bash
cd app && npm run tauri icon ../brand/logo-1024.png
```

## The mark — "the ear"

An **ear of grain whose grains are drawn as nodes on a stem**: a 14-unit stem, five
mirrored pairs of stalks, eleven filled circles.

It carries two readings and both are true of vuna:

- **an ear of wheat** — *vuna* is Swahili for *to harvest / reap*, and the harvest is
  the product's whole thesis;
- **a stem with node-terminated edges** — which is what vuna actually keeps: a link
  graph and its postings, never the pages.

That second reading is why the grains are circles on visible stalks rather than the
usual pointed wheat glyph, and it is the family resemblance: evermesh's mark is also
a shape with edges terminating in node circles. Same grammar, different word.

### Rules

- **No `<text>`, no `font-family`, no gradients, no filters, no shadows.** A mark that
  needs a font to render is not a mark. (The previous `wordmark-dark.svg` set "Vuna"
  in `<text font-family="Georgia">` with hard-coded `#FAFAFA` ink — it rendered
  differently on every platform and was invisible on a light page. It is gone; the
  wordmark is now HTML text in vendored Fraunces, which is what kotva, patala and
  magnetite all do.)
- Two flat colours, ever: the stem, and the grains.
- Keep the spindle. Widening the pairs turns the ear into a tree — it was drawn that
  way once and read as a shrub with berries.
- Below ~24px use `favicon.svg`, not a scaled `mark.svg`. Eleven grains become a
  smudge; five thick ones do not.
- The tile in `logo.svg` follows the suite convention — dark ground, corner radius
  ~0.22× the side (kotva 8/40, magnetite 26/120, patala 28/128, evermesh 56/256,
  vuna 56/256).

## Why evolve rather than replace

The previous identity was a 🌾 emoji in the README plus a gradient-filled "wheat
sheaf" tile — three triangles over an ellipse, which read as arrows in a bowl. Three
things were kept and three were replaced.

**Kept.** *Harvest* as the concept: it is the name, it is unclaimed in the family, and
it survives translation. The **warm** surface temperature: warm paper and warm dark
are vuna's slot in the suite, and the desktop app already committed to it well. The
**type scheme** — Fraunces / Archivo / IBM Plex Mono was already vendored, already
distinctive, and already matches the family's display-serif + grotesk + mono pattern.

**Replaced.** The **geometry**, because a sheaf of triangles is not legible and the
two `linearGradient`s broke the suite's flat-colour rule. The **accent hue**: vuna sat
at rust `#b5540f`/`#e2762f`, which is within a few degrees of kotva's `#cb7245` and
patala's `#f06b3d` — three of five siblings in the same orange would read as one
product. Vuna moved to the yellow side of the harvest, **grain gold**, which is
recognisably not terracotta and not vermilion. And the **semantics**: the old palette
used one green for both "node online" and "peer result", and rust for both the brand
and the local-result chip, so colour carried no reliable meaning.

## Palette — "grain on paper, grain on loam"

Five roles, four colours, and one of them is the absence of colour.

| Role | Meaning | Light | Dark |
|---|---|---|---|
| `--grain` | Identity. Also the `local` result source — **your own shard**. Nothing else is gold. | `#9A6A0A` graphic · `#C98A0E` fill · `#7A5308` text | `#E8B84B` · `#E8B84B` · `#F0C86B` |
| `--peer` | Reach beyond your node — another volunteer's shard. | `#3A5F94` · `#2E4E7C` text | `#8AB2E8` · `#A0C2EE` text |
| *(none)* | `indexer` results. An opt-in indexer adds reach, **never authority**, so it gets no colour of its own — on purpose. | `--ink-3` on `--line` | same |
| `--running` | Built, tested, exercised. You can run it today. | `#3F7A2E` · `#2E5C22` text | `#86C46A` · `#9AD47E` text |
| `--stub` | Specified in the design docs, **not implemented**. | `#A8461B` · `#8E3A14` text | `#EE8A55` · `#F2A077` text |

Surfaces and ink:

| Token | Light | Dark |
|---|---|---|
| `--void` page ground | `#F1EADB` | `#12100A` |
| `--paper` cards | `#FDFAF3` | `#1A170F` |
| `--sunk` wells, code | `#E7DFCC` | `#0D0B07` |
| `--surface-2` band | `#F7F2E6` | `#221E14` |
| `--line` / `--line-2` decorative rules | `#DDD3BD` / `#C4B79C` | `#2E2819` / `#463E29` |
| `--edge` component boundary | `#847756` | `#7C6E50` |
| `--ink` / `--ink-2` / `--ink-3` | `#221C12` / `#554B37` / `#695E45` | `#F2EBDA` / `#C1B499` / `#95896F` |

Note the split that gold forces: **`#C98A0E` on paper is only 2.5:1**, under even the
3:1 non-text floor. So the bright gold is allowed *only* as a fill with `--on-grain`
ink on top of it (5.7:1), and light-theme gold graphics and borders use the deeper
`#9A6A0A` (3.95:1). Dark theme has no such problem and uses one gold for everything.
`--line` and `--line-2` are decorative rules and deliberately below 3:1; anything that
is the boundary of an actual control uses `--edge`.

## Measured contrast — WCAG 2.1

Text pairs, 4.5:1 floor. Every value is a pass.

| Pair | on `--void` | on `--paper` | on `--sunk` | on `--surface-2` |
|---|---|---|---|---|
| **light** `--ink` | 14.11:1 | 16.21:1 | 12.73:1 | 15.13:1 |
| **light** `--ink-2` | 7.16:1 | 8.23:1 | 6.46:1 | 7.68:1 |
| **light** `--ink-3` | 5.33:1 | 6.13:1 | 4.81:1 | 5.72:1 |
| **light** `--grain-text` | 5.71:1 | 6.57:1 | 5.16:1 | 6.13:1 |
| **light** `--peer-text` | 7.02:1 | 8.07:1 | 6.34:1 | — |
| **light** `--running-text` | 6.57:1 | 7.54:1 | 5.93:1 | — |
| **light** `--stub-text` | 6.34:1 | 7.28:1 | 5.72:1 | — |
| **dark** `--ink` | 16.00:1 | 15.06:1 | 16.55:1 | 13.98:1 |
| **dark** `--ink-2` | 9.29:1 | 8.74:1 | 9.60:1 | 8.12:1 |
| **dark** `--ink-3` | 5.51:1 | 5.19:1 | 5.70:1 | 4.82:1 |
| **dark** `--grain-text` | 11.94:1 | 11.23:1 | 12.34:1 | 10.43:1 |
| **dark** `--peer-text` | 10.37:1 | 9.75:1 | 10.72:1 | — |
| **dark** `--running-text` | 10.95:1 | 10.31:1 | 11.32:1 | — |
| **dark** `--stub-text` | 9.12:1 | 8.58:1 | 9.43:1 | — |

Filled buttons and chips: `--on-grain` on `--grain-fill` is **5.66:1** light and
**10.01:1** dark.

Non-text pairs, 3:1 floor — graphics, chip dots, control borders:

| Pair | Light | Dark |
|---|---|---|
| `--grain` graphic on `--void` / `--paper` / `--sunk` | 3.95 / 4.54 / 3.56:1 | 10.31 / 9.71 / — |
| `--peer` graphic on `--void` | 5.40:1 | 8.71:1 |
| `--running` graphic on `--void` | 4.35:1 | 9.16:1 |
| `--stub` graphic on `--void` | 4.93:1 | 7.61:1 |
| `--edge` on `--void` / `--paper` / `--sunk` | 3.69 / 4.24 / 3.33:1 | 3.81 / 3.58 / —:1 |

On the `logo.svg` tile ground `#12100A`: bone stem `#E7DFCC` **14.33:1**, grain gold
`#E8B84B` **10.31:1**.

`--line` (`#DDD3BD` light, 1.3:1) and `--line-2` (`#C4B79C`, 1.65:1) are decorative
hairlines that never carry information on their own, which is why they sit below the
floor. If a rule ever becomes the only cue for something, promote it to `--edge`.

## Type

| Role | Family | Weights | Why |
|---|---|---|---|
| Display | **Fraunces** | 300, 400, 400 italic, 600 | A high-contrast, slightly wonky old-style serif. Warm and agricultural without being rustic pastiche, and unclaimed across the suite (magnetite has Instrument Serif, evermesh Syne, patala Familjen Grotesk). The italic carries the asides. |
| UI / body | **Archivo** | 400, 500, 600, 700 | A grotesk with enough width and a low, even texture at small sizes — the app is dense. |
| Data | **IBM Plex Mono** | 400, 500 | Every URL, node id, score, byte count and test count. Unambiguous `0`/`O`, which matters when the thing on screen is an identifier. Shared with kotva/patala/magnetite — the family's mono. |

All three are SIL OFL 1.1 and **vendored as latin-subset woff2**: `app/src/assets/fonts/`
for the app, `site/fonts/` for the landing page. Licences ship alongside. Nothing is
fetched at runtime — no CDN, no Google Fonts, no `@import url(https://…)`.

The mono rule is the one worth restating: **if a user could copy it and paste it
somewhere that cares, it is mono.** URLs, node ids, space ids (`bge-small-en-v1.5@384/int8`),
crate names, counts.

## Usage rules

1. **Never recolour outside the token set.** Add a token with a measured ratio instead.
2. **Colour means something.** Never use `--stub` because ember looks good, and never
   use `--running` for a thing that is not running. The landing page and the app both
   print a reading key next to the first use.
3. **`indexer` stays uncoloured.** It is the only source that carries no authority; the
   palette says so by omission.
4. Gold is scarce. One gold thing per view, plus the local chips.
5. Pick the light/dark asset with `prefers-color-scheme` (via `<picture>` for raster,
   via the internal `@media` for `favicon.svg`), never by JS after paint.
6. Keep `site/vuna-icon.svg` in sync with `favicon.svg`, and the values in
   `site/index.html`'s inline `<style>` in sync with `tokens.css`.

## Re-deriving the contrast tables

Standard WCAG 2.1 relative luminance — sRGB channel to linear via
`c <= 0.03928 ? c/12.92 : ((c+0.055)/1.055)^2.4`, luminance
`0.2126R + 0.7152G + 0.0722B`, ratio `(Lmax+0.05)/(Lmin+0.05)`. Any implementation of
that formula reproduces the numbers above from the hex values in `tokens.css`. Text
floor 4.5:1, non-text floor 3:1; no value in the palette relies on the 3:1
large-text exemption.
