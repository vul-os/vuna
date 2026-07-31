#!/usr/bin/env node
//
// Responsive/visual verifier for site/, in both themes, across five
// breakpoints. Complements site-check.mjs, which already covers console
// errors, dead links, doc-route content and the vendored display face, but
// does not look at any of the following:
//
//   · outbound requests — site-check.mjs only fails on requestfailed or an
//     HTTP status >= 400, so a *successful* third-party fetch passes it
//     silently. The footer on both pages states "nothing on this page is
//     fetched from a third party." This check records every request the
//     page issues and fails if any of them leave the local origin, whether
//     or not it succeeds.
//   · sub-12px type — the dominant responsive defect across this suite.
//   · horizontal page overflow.
//   · image aspect-ratio preservation — max-width:100% without height:auto
//     silently stretched every screenshot in a sibling repo and was
//     invisible to visual review.
//   · breadth — 320/375/768/1024/1440, light and dark. Vuna has a theme
//     toggle, so both color schemes are real states, not a preference.
//
//   npm run site-check         (runs this after the console/link checks)
//   node scripts/check-responsive.mjs   (standalone)
//
import { chromium } from "playwright";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { mkdir } from "node:fs/promises";
import { startStaticServer } from "./lib/static-server.mjs";

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const SITE_DIR = path.join(ROOT, "site");
const OUT = path.join(
  "/private/tmp/claude-501/-Users-pc-code-vulos/8606e9de-dcba-4abc-8a2a-dc73a4da360b/scratchpad",
  "vuna-shots"
);

const BPS = [320, 375, 768, 1024, 1440];
const SCHEMES = ["light", "dark"];

let fail = 0;
let pagesExamined = 0;
let chaptersExamined = 0;
let imagesChecked = 0;
const say = (ok, msg) => {
  if (!ok) fail++;
  console.log(`${ok ? "PASS" : "FAIL"}  ${msg}`);
};

// vuna's theme is not just `prefers-color-scheme`: a pre-paint inline script
// in <head> reads `localStorage['vulos-theme']` first and only falls back to
// the media query if that key is absent. Setting Playwright's `colorScheme`
// alone would only ever exercise the fallback path — never the path real
// visitors hit after using the in-page toggle, which is the one that
// persists. Seed localStorage via addInitScript so it runs before the
// page's own inline script does, on every navigation in the context.
async function themeContext(browser, w, h, scheme) {
  const ctx = await browser.newContext({
    viewport: { width: w, height: h },
    colorScheme: scheme,
    deviceScaleFactor: 1,
    // html{scroll-behavior:smooth} on both pages races the driver: a script
    // that sets location.hash and then immediately measures reads a layout
    // that is still animating towards its target. Absent this, that shows
    // up as an intermittent failure, not a real one — and a check that is
    // flaky teaches people to re-run it until green instead of trusting it.
    reducedMotion: "reduce",
  });
  await ctx.addInitScript((t) => {
    try {
      localStorage.setItem("vulos-theme", t);
    } catch (e) {
      /* ignore */
    }
  }, scheme);
  return ctx;
}

function watchNetwork(page, base) {
  const external = [];
  const failed = [];
  const consoleErrs = [];
  page.on("request", (r) => {
    if (!r.url().startsWith(base) && !r.url().startsWith("data:")) external.push(r.url());
  });
  page.on("response", (r) => {
    if (r.status() >= 400) failed.push(`${r.status()} ${r.url().replace(base, "")}`);
  });
  page.on("requestfailed", (r) => failed.push(`requestfailed ${r.url().replace(base, "")}`));
  page.on("pageerror", (e) => consoleErrs.push(String(e).slice(0, 160)));
  return { external, failed, consoleErrs };
}

const overflowOf = (page) =>
  page.evaluate(() => {
    const docW = document.documentElement.scrollWidth,
      winW = window.innerWidth;
    const bad = [];
    if (docW > winW + 1) {
      for (const el of document.querySelectorAll("*")) {
        const r = el.getBoundingClientRect();
        if (r.right > winW + 1 && r.width > 0)
          bad.push(`${el.tagName.toLowerCase()}.${(el.className || "").toString().split(" ")[0]} right=${Math.round(r.right)}`);
      }
    }
    return { docW, winW, bad: bad.slice(0, 6) };
  });

const smallOf = (page) =>
  page.evaluate(() => {
    const out = [];
    for (const el of document.querySelectorAll("body *")) {
      if (!el.firstChild) continue;
      let hasText = false;
      for (const n of el.childNodes) if (n.nodeType === 3 && n.textContent.trim()) hasText = true;
      if (!hasText) continue;
      const cs = getComputedStyle(el);
      if (cs.display === "none" || cs.visibility === "hidden") continue;
      const fs = parseFloat(cs.fontSize);
      if (fs < 12) out.push(`${el.tagName.toLowerCase()}.${(el.className || "").toString().split(" ")[0]}=${fs}px "${el.textContent.trim().slice(0, 28)}"`);
    }
    return out;
  });

const stretchedOf = (page) =>
  page.evaluate(() => {
    const out = [];
    for (const img of document.querySelectorAll("img")) {
      if (!img.naturalWidth || !img.naturalHeight) {
        out.push(`${img.getAttribute("src")} DID NOT LOAD`);
        continue;
      }
      const r = img.getBoundingClientRect();
      if (r.width < 2 || r.height < 2) continue; // hidden light/dark twin
      const nat = img.naturalWidth / img.naturalHeight,
        ren = r.width / r.height;
      if (Math.abs(nat - ren) / nat > 0.02)
        out.push(`${img.getAttribute("src")} natural ${nat.toFixed(3)} vs rendered ${ren.toFixed(3)}`);
    }
    return out;
  });

async function settle(page) {
  // A fast auto-scroll does not fire IntersectionObserver and yields blank
  // captures that are an artifact, not a bug — so scroll slowly.
  await page.evaluate(async () => {
    for (const i of document.images) i.loading = "eager";
    const step = window.innerHeight * 0.7;
    for (let y = 0; y < document.body.scrollHeight; y += step) {
      window.scrollTo(0, y);
      await new Promise((r) => setTimeout(r, 80));
    }
    window.scrollTo(0, 0);
    await Promise.all(
      [...document.images].map((i) =>
        i.complete
          ? null
          : Promise.race([
              new Promise((r) => {
                i.onload = i.onerror = r;
              }),
              new Promise((r) => setTimeout(r, 4000)),
            ])
      )
    );
  });
  await page.waitForTimeout(400);
}

async function main() {
  await mkdir(OUT, { recursive: true });
  const site = await startStaticServer(SITE_DIR);
  const browser = await chromium.launch({ headless: true });

  try {
    // Learn the docs route table dynamically, same as site-check.mjs.
    const probeCtx = await themeContext(browser, 1440, 960, "light");
    const probe = await probeCtx.newPage();
    await probe.goto(`${site.base}/docs.html`, { waitUntil: "networkidle" });
    await probe.waitForSelector(".nav-side a");
    const docSlugs = await probe.$$eval(".nav-side a", (as) => as.map((a) => a.dataset.slug));
    await probeCtx.close();

    // The pre-paint script's actual priority is: `vulos-theme` in localStorage
    // wins if it is "light" or "dark"; only otherwise does it fall back to
    // `prefers-color-scheme`. Every other check in this file sets both
    // Playwright's `colorScheme` *and* the localStorage key to the same
    // value, which cannot tell the two paths apart — a context that ignored
    // localStorage entirely would still pass them, because the media-query
    // fallback would happen to agree. Set them to opposite values once, on
    // both pages, to prove localStorage is the one actually taking priority.
    for (const [file, sel] of [
      ["index.html", "html"],
      ["docs.html", "html"],
    ]) {
      const ctx = await browser.newContext({ colorScheme: "light", reducedMotion: "reduce" });
      await ctx.addInitScript(() => {
        try {
          localStorage.setItem("vulos-theme", "dark");
        } catch (e) {
          /* ignore */
        }
      });
      const page = await ctx.newPage();
      await page.goto(`${site.base}/${file}`, { waitUntil: "networkidle" });
      const theme = await page.evaluate((s) => document.querySelector(s).getAttribute("data-theme"), sel);
      say(theme === "dark", `${file}  localStorage("vulos-theme"="dark") outranks colorScheme:"light" (got data-theme="${theme}")`);
      await ctx.close();
    }

    // ── Pass A: index.html across every breakpoint, in both themes ──
    for (const scheme of SCHEMES) {
      for (const w of BPS) {
        const ctx = await themeContext(browser, w, 900, scheme);
        const page = await ctx.newPage();
        const { external, failed, consoleErrs } = watchNetwork(page, site.base);
        await page.goto(`${site.base}/index.html`, { waitUntil: "networkidle" });
        await page.evaluate(() => document.fonts.ready);

        const theme = await page.evaluate(() => document.documentElement.getAttribute("data-theme"));
        say(theme === scheme, `index.html ${scheme}/${w}  data-theme applied via localStorage (got "${theme}")`);

        await settle(page);
        const tag = `index.html ${scheme}/${w}`;

        say(external.length === 0, `${tag}  zero outbound requests${external.length ? " :: " + external.join(", ") : ""}`);
        say(failed.length === 0, `${tag}  no failed/4xx resources${failed.length ? " :: " + failed.join(", ") : ""}`);
        say(consoleErrs.length === 0, `${tag}  no uncaught page errors${consoleErrs.length ? " :: " + consoleErrs.join(" | ") : ""}`);

        const ov = await overflowOf(page);
        say(ov.docW <= ov.winW + 1, `${tag}  no horizontal overflow (scrollWidth ${ov.docW} vs ${ov.winW})${ov.bad.length ? " :: " + ov.bad.join(" | ") : ""}`);

        const sm = await smallOf(page);
        say(sm.length === 0, `${tag}  no text below 12px${sm.length ? " :: " + sm.slice(0, 8).join(" | ") : ""}`);

        const st = await stretchedOf(page);
        imagesChecked += await page.evaluate(() => document.querySelectorAll("img").length);
        say(st.length === 0, `${tag}  every image keeps its aspect ratio${st.length ? " :: " + st.join(" | ") : ""}`);

        const fonts = await page.evaluate(async () => {
          await document.fonts.ready;
          return {
            fraunces: document.fonts.check("16px Fraunces"),
            archivo: document.fonts.check("16px Archivo"),
            plex: document.fonts.check('13px "IBM Plex Mono"'),
          };
        });
        say(
          fonts.fraunces && fonts.archivo && fonts.plex,
          `${tag}  vendored fonts loaded (Fraunces=${fonts.fraunces}, Archivo=${fonts.archivo}, Plex Mono=${fonts.plex})`
        );

        await page.screenshot({ path: `${OUT}/index-${scheme}-${w}.png`, fullPage: false });
        pagesExamined++;
        await ctx.close();
      }
    }

    // ── Pass B: docs.html, every chapter, at a phone size and a desk size ──
    for (const [w, h, scheme] of [
      [375, 812, "light"],
      [1440, 960, "dark"],
    ]) {
      const ctx = await themeContext(browser, w, h, scheme);
      const page = await ctx.newPage();
      const { external, failed } = watchNetwork(page, site.base);
      await page.goto(`${site.base}/docs.html`, { waitUntil: "networkidle" });
      await page.evaluate(() => document.fonts.ready);

      for (const slug of docSlugs) {
        await page.evaluate((s) => {
          location.hash = "#" + s;
        }, slug);
        await page.waitForFunction(
          (s) => {
            const a = document.querySelector(`.nav-side a[data-slug="${s}"]`);
            const c = document.getElementById("content");
            return a?.classList.contains("active") && c && c.textContent.trim() !== "Loading…";
          },
          slug,
          { timeout: 10000 }
        );
        await page.waitForTimeout(120);

        const ptag = `docs.html#${slug} ${scheme}/${w}`;
        const ov = await overflowOf(page);
        say(ov.docW <= ov.winW + 1, `${ptag}  no horizontal overflow (${ov.docW} vs ${ov.winW})${ov.bad.length ? " :: " + ov.bad.join(" | ") : ""}`);
        const sm = await smallOf(page);
        say(sm.length === 0, `${ptag}  no text below 12px${sm.length ? " :: " + sm.slice(0, 6).join(" | ") : ""}`);
        chaptersExamined++;
      }

      say(
        external.length === 0,
        `docs.html ${scheme}/${w}  zero outbound requests across ${docSlugs.length} chapters${external.length ? " :: " + external.join(", ") : ""}`
      );
      say(failed.length === 0, `docs.html ${scheme}/${w}  no failed/4xx resources${failed.length ? " :: " + failed.join(", ") : ""}`);
      pagesExamined++;
      await ctx.close();
    }

    // ── Pass C: docs.html full breakpoint sweep on its richest chapter (tables + code blocks) ──
    const richChapter = docSlugs.includes("architecture") ? "architecture" : docSlugs[0];
    for (const scheme of SCHEMES) {
      for (const w of BPS) {
        const ctx = await themeContext(browser, w, 900, scheme);
        const page = await ctx.newPage();
        const { external, failed } = watchNetwork(page, site.base);
        await page.goto(`${site.base}/docs.html#${richChapter}`, { waitUntil: "networkidle" });
        await page.waitForFunction(
          () => {
            const c = document.getElementById("content");
            return c && c.textContent.trim() !== "Loading…";
          },
          null,
          { timeout: 10000 }
        );
        await settle(page);

        const tag = `docs.html#${richChapter} ${scheme}/${w}`;
        say(external.length === 0, `${tag}  zero outbound requests${external.length ? " :: " + external.join(", ") : ""}`);
        say(failed.length === 0, `${tag}  no failed/4xx resources${failed.length ? " :: " + failed.join(", ") : ""}`);
        const ov = await overflowOf(page);
        say(ov.docW <= ov.winW + 1, `${tag}  no horizontal overflow (${ov.docW} vs ${ov.winW})${ov.bad.length ? " :: " + ov.bad.join(" | ") : ""}`);
        const sm = await smallOf(page);
        say(sm.length === 0, `${tag}  no text below 12px${sm.length ? " :: " + sm.slice(0, 8).join(" | ") : ""}`);

        await page.screenshot({ path: `${OUT}/docs-${richChapter}-${scheme}-${w}.png`, fullPage: false });
        pagesExamined++;
        await ctx.close();
      }
    }

    // Coverage assertions: an exact expected count, not just ">0", so a loop
    // that silently examines fewer pages than it should (a shortened range, a
    // `break` that fires early) is itself a failure, not a quieter pass.
    const expectedPages = SCHEMES.length * BPS.length * 2 + 2; // Pass A + Pass C + Pass B(2)
    const expectedChapters = 2 * docSlugs.length; // Pass B
    const expectedImages = 6 * SCHEMES.length * BPS.length; // Pass A, 6 <img> on index.html
    say(pagesExamined === expectedPages, `examined ${pagesExamined}/${expectedPages} page renders across breakpoints/themes`);
    say(chaptersExamined === expectedChapters, `examined ${chaptersExamined}/${expectedChapters} chapter renders in the docs viewer`);
    say(imagesChecked === expectedImages, `checked aspect ratio on ${imagesChecked}/${expectedImages} image instances`);
  } finally {
    await browser.close();
    await site.close();
  }

  console.log(`\n${fail === 0 ? "ALL RESPONSIVE/VISUAL CHECKS PASSED" : fail + " CHECK(S) FAILED"}`);
  process.exit(fail === 0 ? 0 : 1);
}

main().catch((err) => {
  console.error(err.message ?? err);
  process.exit(1);
});
