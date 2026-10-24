import react from "@vitejs/plugin-react";
import type { ProxyOptions } from "vite";
import { defineConfig } from "vitest/config";

const backend = process.env.IM_BACKEND ?? "http://127.0.0.1:8765";
const session = process.env.IM_SESSION ?? "";

const proxy: ProxyOptions = {
  target: backend,
  changeOrigin: true,
  configure: (p) => {
    p.on("proxyReq", (req) => {
      req.removeHeader("origin");
      req.removeHeader("sec-fetch-site");
      if (session) req.setHeader("X-IM-Session", session);
    });
  },
};

export default defineConfig({
  plugins: [react()],
  base: "/",
  build: {
    outDir: "dist",
    emptyOutDir: true,
    assetsDir: "app",
    target: "es2022",
    sourcemap: false,
  },
  server: {
    proxy: {
      "/api": proxy,
      "/events": proxy,
      "/notification-sound": proxy,
    },
  },
  test: {
    environment: "jsdom",
    globals: false,
  },
});
