#!/usr/bin/env node
//
// Light/dark screenshots of the vuna landing page and the Tauri desktop shell.
//
//   npm install               # repo root, installs playwright (devDependency)
//   npx playwright install chromium
//   npm run screenshots
//
// Writes to docs/shots/ (used by README.md) and mirrors into site/shots/ so the site
// directory deploys standalone. Exits non-zero if any capture fails.
//
// ── On mock data ────────────────────────────────────────────────────────────────────────
// The desktop shell is captured in a plain browser against app/dist, where lib/api.ts falls
// back to its in-process fixtures. That is not a screenshot-only shortcut: `vuna-query` and
// `vuna-node` are Wave-2 stubs, so the Tauri build is equally mock (see the doc comment at
// the top of app/src-tauri/src/commands.rs, which opens "**mock data only**"). The app
// renders a permanent, non-dismissible mock-data banner, and this script asserts that the
// banner is present in every app capture before writing the file. A screenshot of this UI
// without that banner would be a false claim, so it is a hard failure, not a warning.

import { chromium } from "playwright";
import { mkdir, copyFile, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { startStaticServer } from "./lib/static-server.mjs";

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const SITE_DIR = path.join(ROOT, "site");
const APP_DIST = path.join(ROOT, "app", "dist");
const OUT = path.join(ROOT, "docs", "shots");
const SITE_OUT = path.join(ROOT, "site", "shots");

const THEMES = ["light", "dark"];
const VIEWPORT = { width: 1440, height: 900 };

/** Landing/docs pages, captured full-page. */
const SITE_ROUTES = [
  { name: "site", url: "/index.html", waitFor: "h1", fullPage: true },
  { name: "site-docs", url: "/docs.html", waitFor: ".md h1", fullPage: true },
];

/** Desktop-shell surfaces. `clip` crops to an element; `type` fills the query box first. */
const APP_ROUTES = [
  { name: "search", url: "/index.html", waitFor: ".result-item", type: "embedding model" },
  { name: "dashboard", url: "/index.html", waitFor: ".dashboard .kv", clip: ".dashboard" },
];

const problems = [];

/** Console errors, uncaught exceptions and failed requests are all build failures here:
    a clean-looking screenshot of a page that threw is a false positive. */
function watch(page, where) {
  page.on("console", (m) => {
    if (m.type() === "error") problems.push(`${where}: console error — ${m.text()}`);
  });
  page.on("pageerror", (e) => problems.push(`${where}: uncaught — ${e.message}`));
  page.on("requestfailed", (r) => {
    const f = r.failure();
    problems.push(`${where}: request failed — ${r.url()} (${f ? f.errorText : "unknown"})`);
  });
  page.on("response", (r) => {
    if (r.status() >= 400) problems.push(`${where}: HTTP ${r.status()} — ${r.url()}`);
  });
}

async function context(browser, theme) {
  const ctx = await browser.newContext({
    viewport: VIEWPORT,
    deviceScaleFactor: 2,
    colorScheme: theme,
    // Reveal animations must have settled, or two runs of this script differ byte for byte.
    reducedMotion: "reduce",
  });
  // Both the site and the app read the suite-wide key; seed it so the capture never depends
  // on which of the two theme paths happens to win.
  await ctx.addInitScript((t) => {
    try {
      localStorage.setItem("vulos-theme", t);
    } catch (e) {
      /* ignore */
    }
  }, theme);
  return ctx;
}

async function shoot(ctx, base, route, theme, label) {
  const page = await ctx.newPage();
  watch(page, `${label}/${route.name} (${theme})`);
  const file = path.join(OUT, `${route.name}-${theme}.png`);

  await page.goto(base + route.url, { waitUntil: "networkidle" });
  await page.evaluate(() => document.fonts.ready);
  if (route.waitFor) await page.waitForSelector(route.waitFor, { timeout: 15000 });

  // The vendored display face must have actually applied. A broken @font-face path renders a
  // plausible-looking page in a fallback serif, which is exactly the failure a screenshot hides.
  const displayFont = await page.evaluate(() => {
    const h = document.querySelector("h1, .wordmark-title");
    return h ? getComputedStyle(h).fontFamily : "";
  });
  if (!/Fraunces/i.test(displayFont)) {
    problems.push(`${label}/${route.name} (${theme}): display font did not apply — got "${displayFont}"`);
  }

  // Walk the page so nothing below the fold is still unfetched when the shutter fires — a
  // full-page capture of an unloaded image is a blank box, which is how one shipped once.
  if (route.fullPage) {
    await page.evaluate(async () => {
      const step = window.innerHeight;
      for (let y = 0; y < document.body.scrollHeight; y += step) {
        window.scrollTo(0, y);
        await new Promise((r) => setTimeout(r, 90));
      }
      window.scrollTo(0, 0);
    });
    await page.evaluate(() =>
      Promise.all(
        [...document.images].filter((i) => !i.complete).map((i) => i.decode().catch(() => {}))
      )
    );
  }

  await page.waitForTimeout(450);

  if (route.clip) {
    const el = await page.waitForSelector(route.clip, { timeout: 10000 });
    await el.screenshot({ path: file });
  } else {
    await page.screenshot({ path: file, fullPage: Boolean(route.fullPage) });
  }

  await page.close();
  return file;
}

async function main() {
  await mkdir(OUT, { recursive: true });
  await mkdir(SITE_OUT, { recursive: true });

  const site = await startStaticServer(SITE_DIR);
  const app = await startStaticServer(APP_DIST);
  const browser = await chromium.launch({ headless: true });
  const written = [];

  try {
    for (const theme of THEMES) {
      const ctx = await context(browser, theme);
      try {
        for (const route of SITE_ROUTES) {
          written.push(await shoot(ctx, site.base, route, theme, "site"));
        }

        for (const route of APP_ROUTES) {
          const page = await ctx.newPage();
          watch(page, `app/${route.name} (${theme})`);
          const file = path.join(OUT, `${route.name}-${theme}.png`);

          await page.goto(app.base + route.url, { waitUntil: "networkidle" });
          await page.evaluate(() => document.fonts.ready);
          await page.waitForSelector(route.waitFor, { timeout: 15000 });

          if (route.type) {
            await page.fill(".search-input", route.type);
            await page.waitForTimeout(700); // debounce + fixture latency
            await page.waitForSelector(".result-item", { timeout: 10000 });
          }

          // Hard gate: whatever region this capture covers, the mock-data disclosure has to be
          // inside it. The banner covers the full-window shots; the panel chip covers the crop.
          const scope = route.clip ?? ".mock-banner";
          const disclosure = await page.textContent(scope);
          // No \b anchors: textContent concatenates without separators ("Your node" + "mock").
          if (!disclosure || !/mock/i.test(disclosure)) {
            throw new Error(
              `app/${route.name} (${theme}): no mock-data disclosure inside the captured region ` +
                `(${scope}). Every screenshot of this UI must carry one — refusing to write ` +
                `${path.basename(file)}.`
            );
          }

          const displayFont = await page.evaluate(
            () => getComputedStyle(document.querySelector(".wordmark-title")).fontFamily
          );
          if (!/Fraunces/i.test(displayFont)) {
            problems.push(`app/${route.name} (${theme}): display font did not apply — "${displayFont}"`);
          }

          await page.waitForTimeout(450);
          if (route.clip) {
            const el = await page.waitForSelector(route.clip, { timeout: 10000 });
            await el.screenshot({ path: file });
          } else {
            await page.screenshot({ path: file });
          }
          await page.close();
          written.push(file);
        }
      } finally {
        await ctx.close();
      }
    }
  } finally {
    await browser.close();
    await site.close();
    await app.close();
  }

  // site/ has to deploy on its own, so it gets its own copy of every shot it references.
  for (const file of written) {
    await copyFile(file, path.join(SITE_OUT, path.basename(file)));
  }

  await writeFile(
    path.join(OUT, "README.md"),
    `# Screenshots

Generated by \`npm run screenshots\` (\`scripts/screenshots.mjs\`). Do not hand-edit; rerun
the script. Every file is captured at ${VIEWPORT.width}×${VIEWPORT.height} at 2× device scale,
in both themes, with animations disabled so reruns are comparable.

**Every \`search-*\` and \`dashboard-*\` shot shows mock data.** \`vuna-query\` and
\`vuna-node\` are Wave-2 stubs, so the desktop shell renders fixtures from
\`app/src/lib/api.ts\` (browser) or \`app/src-tauri/src/commands.rs\` (Tauri, whose doc
comment opens "**mock data only**"). The app carries a permanent, non-dismissible mock-data
banner and this script refuses to write an app screenshot that does not contain it.

| File | Surface | Data |
|---|---|---|
${written
  .map((f) => path.basename(f))
  .sort()
  .map((f) => {
    const app = /^(search|dashboard)-/.test(f);
    return `| \`${f}\` | ${app ? "Tauri desktop shell" : "Landing page"} | ${app ? "**mock**" : "n/a"} |`;
  })
  .join("\n")}
`,
    "utf8"
  );

  console.log(`\nwrote ${written.length} screenshots to docs/shots/ and site/shots/:`);
  for (const f of written) console.log(`  ${path.relative(ROOT, f)}`);

  if (problems.length) {
    console.error(`\n${problems.length} problem(s) during capture:`);
    for (const p of problems.slice(0, 20)) console.error(`  ✗ ${p}`);
    process.exit(1);
  }
  console.log("\nno console errors, no failed requests, display font applied everywhere.");
}

main().catch((err) => {
  console.error(err.message ?? err);
  process.exit(1);
});
