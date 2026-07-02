import { defineConfig } from "vite";

import { cloudflare } from "@cloudflare/vite-plugin";

export default defineConfig({
  plugins: [cloudflare()],
  base: "./",
  server: {
    port: 5173,
    host: true,
  },
});