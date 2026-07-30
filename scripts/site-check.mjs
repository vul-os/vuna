#!/usr/bin/env node
//
// Verifies site/ in a real browser, in both themes:
//
//   · no console errors, no uncaught exceptions
//   · no failed requests and no 4xx/5xx — a missing vendored font or a missing doc fails here
//   · every same-origin link resolves (fragment to a real id, file to a real file)
//   · the vendored display face actually applied, which a broken @font-face path would hide
//   · every docs route renders real content rather than an error card
//
//   npm run site-check
//
// Exits non-zero on the first problem set. Nothing here touches the Cargo workspace.

import { chromium } from "playwright";
import { access } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { startStaticServer } from "./lib/static-server.mjs";

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const SITE_DIR = path.join(ROOT, "site");

const problems = [];
const note = (m) => problems.push(m);

function watch(page, where) {
  page.on("console", (m) => {
    if (m.type() === "error") note(`${where}: console error — ${m.text()}`);
  });
  page.on("pageerror", (e) => note(`${where}: uncaught — ${e.message}`));
  page.on("requestfailed", (r) => {
    const f = r.failure();
    note(`${where}: request failed — ${r.url()} (${f ? f.errorText : "unknown"})`);
  });
  page.on("response", (r) => {
    if (r.status() >= 400) note(`${where}: HTTP ${r.status()} — ${r.url()}`);
  });
}

/** Every same-origin href on the page has to go somewhere real. */
async function checkLinks(page, pageFile, docSlugs) {
  const hrefs = await page.$$eval("a[href]", (as) => as.map((a) => a.getAttribute("href")));
  for (const href of hrefs) {
    if (!href || /^(https?:|mailto:|#\s*$)/i.test(href)) continue;

    if (href.startsWith("#")) {
      const id = href.slice(1);
      // On docs.html the fragment is a viewer route, not an element id.
      if (pageFile === "docs.html" && docSlugs.includes(id)) continue;
      const ok = await page.evaluate((i) => Boolean(document.getElementById(i)), id);
      if (!ok) note(`${pageFile}: dead fragment #${id}`);
      continue;
    }

    const [file, hash] = href.replace(/^\.\//, "").split("#");
    try {
      await access(path.join(SITE_DIR, file));
    } catch {
      note(`${pageFile}: link to a file that is not in site/ — ${href}`);
      continue;
    }
    // A fragment on docs.html has to be a slug the viewer actually serves.
    if (file === "docs.html" && hash && !docSlugs.includes(hash)) {
      note(`${pageFile}: docs.html#${hash} is not a route the viewer serves`);
    }
  }
}

async function main() {
  const site = await startStaticServer(SITE_DIR);
  const browser = await chromium.launch({ headless: true });
  let routeCount = 0;

  try {
    // Learn the docs route table first — the landing page links into it.
    const probe = await browser.newPage();
    await probe.goto(`${site.base}/docs.html`, { waitUntil: "networkidle" });
    await probe.waitForSelector(".nav-side a");
    const docSlugs = await probe.$$eval(".nav-side a", (as) => as.map((a) => a.dataset.slug));
    await probe.close();
    if (docSlugs.length < 5) note(`docs.html: only ${docSlugs.length} routes in the sidebar`);

    for (const scheme of ["light", "dark"]) {
      const ctx = await browser.newContext({
        viewport: { width: 1440, height: 960 },
        colorScheme: scheme,
      });

      // ── Landing page ──
      const page = await ctx.newPage();
      watch(page, `index.html (${scheme})`);
      await page.goto(`${site.base}/index.html`, { waitUntil: "networkidle" });
      await page.evaluate(() => document.fonts.ready);

      const h1 = (await page.textContent("h1")) ?? "";
      if (!/tally/i.test(h1)) note(`index.html (${scheme}): hero headline missing — got "${h1}"`);

      const font = await page.evaluate(
        () => getComputedStyle(document.querySelector("h1")).fontFamily
      );
      if (!/Fraunces/i.test(font)) {
        note(`index.html (${scheme}): the vendored display face did not apply — got "${font}"`);
      }

      // The page must still say the honest thing, in both themes.
      const body = (await page.textContent("body")) ?? "";
      for (const claim of [
        "does not yet crawl-to-query end to end",
        "132 tests green",
        "mock data",
        "Wave-2 stub",
      ]) {
        if (!body.includes(claim)) {
          note(`index.html (${scheme}): the status caveat "${claim}" is gone from the page`);
        }
      }

      await checkLinks(page, "index.html", docSlugs);
      await page.close();

      // ── Docs viewer, every route ──
      const docs = await ctx.newPage();
      watch(docs, `docs.html (${scheme})`);
      await docs.goto(`${site.base}/docs.html`, { waitUntil: "networkidle" });
      await docs.evaluate(() => document.fonts.ready);

      for (const slug of docSlugs) {
        await docs.evaluate((s) => {
          location.hash = "#" + s;
        }, slug);
        await docs.waitForFunction(
          (s) => {
            const a = document.querySelector(`.nav-side a[data-slug="${s}"]`);
            const c = document.getElementById("content");
            return a?.classList.contains("active") && c && c.textContent.trim() !== "Loading…";
          },
          slug,
          { timeout: 10000 }
        );
        const bad = await docs.$(".err");
        if (bad) note(`docs.html#${slug} (${scheme}): rendered an error card`);
        const len = await docs.evaluate(
          () => document.getElementById("content").textContent.trim().length
        );
        if (len < 400) note(`docs.html#${slug} (${scheme}): only ${len} chars of content`);
        routeCount++;
      }

      await checkLinks(docs, "docs.html", docSlugs);
      await docs.close();
      await ctx.close();
    }
  } finally {
    await browser.close();
    await site.close();
  }

  if (problems.length) {
    console.error(`\n${problems.length} problem(s):`);
    for (const p of problems) console.error(`  ✗ ${p}`);
    process.exit(1);
  }
  console.log(
    `site OK — landing page (light + dark), ${routeCount / 2} doc routes per theme, ` +
      `no console errors, no broken links, vendored fonts applied, status caveats present.`
  );
}

main().catch((err) => {
  console.error(err.message ?? err);
  process.exit(1);
});
