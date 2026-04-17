import react from "@vitejs/plugin-react-swc";
// import { visualizer } from "rollup-plugin-visualizer";
import { defineConfig } from "vite";

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    // visualizer({ open: true })
  ],
  server: {
    proxy: {
      "/api": {
        target: "https://localhost",
        secure: false,
      },
    },
  },
});
