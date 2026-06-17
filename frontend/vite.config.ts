import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// En dev, le serveur Vite (5173) proxy les appels /api vers le backend Go (8080).
// En prod, on build dans dist/ et c'est le serveur Go qui sert ces fichiers.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api": "http://localhost:8080",
    },
  },
  build: {
    outDir: "dist",
  },
});
