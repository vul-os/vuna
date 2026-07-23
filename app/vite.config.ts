import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// Vuna desktop app — Vite config tuned for Tauri v2 (fixed port, no clearScreen so
// `cargo tauri dev` logs stay visible, ignores src-tauri for the watcher).
const host = process.env.TAURI_DEV_HOST;
const isDebug = !!process.env.TAURI_ENV_DEBUG;
const isWindows = process.env.TAURI_ENV_PLATFORM === "windows";

export default defineConfig({
  plugins: [react()],
  clearScreen: false,
  server: {
    port: 1420,
    strictPort: true,
    host: host || false,
    hmr: host ? { protocol: "ws", host, port: 1421 } : undefined,
    watch: {
      ignored: ["**/src-tauri/**"],
    },
  },
  envPrefix: ["VITE_", "TAURI_"],
  build: {
    target: isWindows ? "chrome105" : "safari13",
    minify: isDebug ? false : "esbuild",
    sourcemap: isDebug,
  },
});
