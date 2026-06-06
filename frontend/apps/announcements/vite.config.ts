import react from "@vitejs/plugin-react";
import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";

const pathFromApp = (path: string) => fileURLToPath(new URL(path, import.meta.url));
const pathFromFrontend = (path: string) => fileURLToPath(new URL(`../../${path}`, import.meta.url));

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: [
      { find: /^@\/features\/([^/]+)\/(.*)$/, replacement: `${pathFromFrontend("apps")}/$1/src/$2` },
      { find: "@/features", replacement: pathFromFrontend("apps") },
      { find: "@/store/api", replacement: pathFromFrontend("packages/api/src/api") },
      { find: "@/store/auth", replacement: pathFromFrontend("packages/auth/src/auth") },
      { find: "@/store/types", replacement: pathFromFrontend("packages/types/src/types") },
      { find: "@/shared", replacement: pathFromFrontend("packages/ui/src/shared") },
      { find: "@/lib", replacement: pathFromFrontend("packages/config/src/lib") },
      { find: "@/ui", replacement: pathFromFrontend("packages/ui/src/ui") },
      { find: "@/utils", replacement: pathFromFrontend("packages/utils/src/utils") },
      { find: "@moh-sso/api", replacement: pathFromFrontend("packages/api/src") },
      { find: "@moh-sso/microfrontend", replacement: pathFromFrontend("packages/microfrontend/src") },
      { find: "@moh-sso/state", replacement: pathFromFrontend("packages/state/src") },
      { find: "@moh-sso/ui", replacement: pathFromFrontend("packages/ui/src") },
    ],
  },
  build: {
    outDir: pathFromApp("./dist"),
    emptyOutDir: true,
    lib: {
      entry: pathFromApp("./src/single-spa.tsx"),
      name: "MohSsoAnnouncements",
      formats: ["es"],
      fileName: () => "single-spa.js",
    },
    rollupOptions: {
      external: ["react", "react-dom", "react-redux", "single-spa"],
    },
  },
});
