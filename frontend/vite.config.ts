import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";

const backendTarget = process.env.VITE_API_PROXY_TARGET || "http://localhost:8000";

const isLocalApiAssetRequest = (url: string): boolean => {
  const pathname = (url.split("?")[0] || "").split("#")[0] || "";
  if (!pathname.startsWith("/api/")) {
    return false;
  }
  return /\.(ts|tsx|js|jsx|mjs|css|scss|sass|json|map|png|jpe?g|gif|svg|webp|avif|ico|woff2?|ttf|eot)$/.test(
    pathname
  );
};

export default defineConfig({
  plugins: [react()],
  server: {
    host: "0.0.0.0",
    port: 5180,
    proxy: {
      "/api": {
        target: backendTarget,
        changeOrigin: true,
        bypass(req, _res) {
          const url = req.url ? decodeURIComponent(req.url) : "";
          if (isLocalApiAssetRequest(url)) {
            return url;
          }
          return null;
        },
      },
      "/health": {
        target: backendTarget,
        changeOrigin: true,
      },
    },
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./"),
    },
  },
});