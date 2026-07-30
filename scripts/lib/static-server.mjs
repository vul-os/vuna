// A ~40-line static file server on an ephemeral port, shared by scripts/screenshots.mjs and
// scripts/site-check.mjs. Both need a real origin: the docs viewer `fetch`es markdown, which
// file:// blocks, and woff2 faces have to be served with a font MIME type or the browser
// silently falls back and the screenshots come out in the wrong typeface.

import { createServer } from "node:http";
import { readFile, stat } from "node:fs/promises";
import path from "node:path";

const MIME = {
  ".html": "text/html; charset=utf-8",
  ".js": "text/javascript; charset=utf-8",
  ".mjs": "text/javascript; charset=utf-8",
  ".css": "text/css; charset=utf-8",
  ".json": "application/json; charset=utf-8",
  ".md": "text/markdown; charset=utf-8",
  ".svg": "image/svg+xml",
  ".png": "image/png",
  ".jpg": "image/jpeg",
  ".webp": "image/webp",
  ".ico": "image/x-icon",
  ".woff2": "font/woff2",
  ".woff": "font/woff",
  ".txt": "text/plain; charset=utf-8",
  ".webmanifest": "application/manifest+json",
};

/** Serves `rootDir` on 127.0.0.1:<ephemeral>. Resolves to `{ base, close }`. */
export function startStaticServer(rootDir) {
  const root = path.resolve(rootDir);

  const server = createServer(async (req, res) => {
    try {
      const url = new URL(req.url, "http://127.0.0.1");
      let filePath = path.join(root, decodeURIComponent(url.pathname));

      // Never serve outside the root, however the path is spelled.
      if (!filePath.startsWith(root)) {
        res.writeHead(403).end("forbidden");
        return;
      }

      let info = await stat(filePath).catch(() => null);
      if (info?.isDirectory()) {
        filePath = path.join(filePath, "index.html");
        info = await stat(filePath).catch(() => null);
      }
      if (!info) {
        res.writeHead(404).end("not found");
        return;
      }

      const body = await readFile(filePath);
      res.writeHead(200, {
        "content-type": MIME[path.extname(filePath).toLowerCase()] ?? "application/octet-stream",
        "cache-control": "no-store",
      });
      res.end(body);
    } catch (err) {
      res.writeHead(500).end(String(err));
    }
  });

  return new Promise((resolve) => {
    server.listen(0, "127.0.0.1", () => {
      const { port } = server.address();
      resolve({
        base: `http://127.0.0.1:${port}`,
        close: () => new Promise((done) => server.close(done)),
      });
    });
  });
}
