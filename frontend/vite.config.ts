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
      { find: "@moh-sso/microfrontend", replacement: pathFromRoot("./packages/microfrontend/src") },
      { find: "@moh-sso/state", replacement: pathFromRoot("./packages/state/src") },
      { find: "@moh-sso/types", replacement: pathFromRoot("./packages/types/src") },
      { find: "@moh-sso/ui", replacement: pathFromRoot("./packages/ui/src") },
      { find: "@moh-sso/utils", replacement: pathFromRoot("./packages/utils/src") },
      { find: "@moh-sso/announcements/single-spa", replacement: pathFromRoot("./apps/announcements/src/single-spa.tsx") },
      { find: "@moh-sso/audit/single-spa", replacement: pathFromRoot("./apps/audit/src/single-spa.tsx") },
      { find: "@moh-sso/clients/single-spa", replacement: pathFromRoot("./apps/clients/src/single-spa.tsx") },
      { find: "@moh-sso/data-visualizer/single-spa", replacement: pathFromRoot("./apps/data-visualizer/src/single-spa.tsx") },
      { find: "@moh-sso/documents/single-spa", replacement: pathFromRoot("./apps/documents/src/single-spa.tsx") },
      { find: "@moh-sso/e-services/single-spa", replacement: pathFromRoot("./apps/e-services/src/single-spa.tsx") },
      { find: "@moh-sso/email/single-spa", replacement: pathFromRoot("./apps/email/src/single-spa.tsx") },
      { find: "@moh-sso/issue-tracker/single-spa", replacement: pathFromRoot("./apps/issue-tracker/src/single-spa.tsx") },
      { find: "@moh-sso/report-browser/single-spa", replacement: pathFromRoot("./apps/report-browser/src/single-spa.tsx") },
      { find: "@moh-sso/surveillance/single-spa", replacement: pathFromRoot("./apps/surveillance/src/single-spa.tsx") },
      { find: "@moh-sso/users/single-spa", replacement: pathFromRoot("./apps/users/src/single-spa.tsx") },
      { find: "@moh-sso/utilities/single-spa", replacement: pathFromRoot("./apps/utilities/src/single-spa.tsx") },
      { find: "@moh-sso/announcements", replacement: pathFromRoot("./apps/announcements/src") },
      { find: "@moh-sso/audit", replacement: pathFromRoot("./apps/audit/src") },
      { find: "@moh-sso/clients", replacement: pathFromRoot("./apps/clients/src") },
      { find: "@moh-sso/data-visualizer", replacement: pathFromRoot("./apps/data-visualizer/src") },
      { find: "@moh-sso/documents", replacement: pathFromRoot("./apps/documents/src") },
      { find: "@moh-sso/e-services", replacement: pathFromRoot("./apps/e-services/src") },
      { find: "@moh-sso/email", replacement: pathFromRoot("./apps/email/src") },
      { find: "@moh-sso/issue-tracker", replacement: pathFromRoot("./apps/issue-tracker/src") },
      { find: "@moh-sso/report-browser", replacement: pathFromRoot("./apps/report-browser/src") },
      { find: "@moh-sso/surveillance", replacement: pathFromRoot("./apps/surveillance/src") },
      { find: "@moh-sso/users", replacement: pathFromRoot("./apps/users/src") },
      { find: "@moh-sso/utilities", replacement: pathFromRoot("./apps/utilities/src") },
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
