import path from "path"
import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import { defineConfig } from "vite"

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(import.meta.dirname, "./src"),
    },
  },
  server: {
    proxy: {
      // 本地开发免 CORS：/sponsor/* → https://api.axlmc.org/*
      "/sponsor": {
        target: "https://api.axlmc.org",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/sponsor/, ""),
      },
    },
  },
})
