import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { viteSingleFile } from "vite-plugin-singlefile";

export default defineConfig({
  plugins: [react(), viteSingleFile()],
  server: {
    port: 5173,
    proxy: {
      "/api": process.env.VITE_API_PROXY ?? "http://localhost:4173",
    },
  },
});
