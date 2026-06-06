import react from "@vitejs/plugin-react";
import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";

const pathFromRoot = (path: string) => fileURLToPath(new URL(path, import.meta.url));

export default defineConfig({
  base: "/portal/",
  root: pathFromRoot("./apps/shell"),
  publicDir: pathFromRoot("./public"),
  plugins: [react()],
  resolve: {
    alias: [
      {
        find: /^@\/features\/([^/]+)\/(.*)$/,
        replacement: `${pathFromRoot("./apps")}/$1/src/$2`,
      },
      { find: "@/features", replacement: pathFromRoot("./apps") },
      { find: "@/store/api", replacement: pathFromRoot("./packages/api/src/api") },
      { find: "@/store/auth", replacement: pathFromRoot("./packages/auth/src/auth") },
      { find: "@/store/types", replacement: pathFromRoot("./packages/types/src/types") },
      { find: "@/store/ui", replacement: pathFromRoot("./packages/state/src/ui") },
      { find: "@/store/clients", replacement: pathFromRoot("./packages/state/src/clients") },
      { find: "@/store/document", replacement: pathFromRoot("./packages/state/src/document") },
      { find: "@/store", replacement: pathFromRoot("./packages/state/src") },
      { find: "@/shared", replacement: pathFromRoot("./packages/ui/src/shared") },
      { find: "@/config", replacement: pathFromRoot("./packages/config/src") },
      { find: "@/lib", replacement: pathFromRoot("./packages/config/src/lib") },
      { find: "@/ui", replacement: pathFromRoot("./packages/ui/src/ui") },
      { find: "@/utils", replacement: pathFromRoot("./packages/utils/src/utils") },
      { find: "@/types", replacement: pathFromRoot("./packages/types/src/global") },
      { find: "@moh-sso/api", replacement: pathFromRoot("./packages/api/src") },
      { find: "@moh-sso/auth", replacement: pathFromRoot("./packages/auth/src") },
      { find: "@moh-sso/config", replacement: pathFromRoot("./packages/config/src") },
      { find: "@moh-sso/state", replacement: pathFromRoot("./packages/state/src") },
      { find: "@moh-sso/types", replacement: pathFromRoot("./packages/types/src") },
      { find: "@moh-sso/ui", replacement: pathFromRoot("./packages/ui/src") },
      { find: "@moh-sso/utils", replacement: pathFromRoot("./packages/utils/src") },
      { find: "@", replacement: pathFromRoot("./apps/shell/src") },
    ],
  },
  server: {
    host: true,
    port: 3000,
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
