import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

export default defineConfig({
  base: "/portal/",
  plugins: [react()],
  server: {
    host: true,
    port: 3000,
    watch: {
      usePolling: true,
      interval: 300,
    },
  },
  optimizeDeps: {
    include: ["react-pivottable/PivotTableUI"],
  },
  build: {
    commonjsOptions: {
      include: [/react-pivottable/, /node_modules/],
    },
  },
});
