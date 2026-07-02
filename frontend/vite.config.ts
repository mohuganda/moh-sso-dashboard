import react from "@vitejs/plugin-react";
import { readdirSync } from "node:fs";
import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";

const pathFromRoot = (path: string) => fileURLToPath(new URL(path, import.meta.url));

const getDirectories = (source: string) =>
  readdirSync(pathFromRoot(source), { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => entry.name);

const apps = getDirectories("./apps");
const packages = getDirectories("./packages");

const dynamicAliases = [
  ...packages.map((pkg) => ({
    find: new RegExp(`^@moh-sso/${pkg}/(.+)$`),
    replacement: pathFromRoot(`./packages/${pkg}/src/$1`),
  })),

  ...packages.map((pkg) => ({
    find: `@moh-sso/${pkg}`,
    replacement: pathFromRoot(`./packages/${pkg}/src/index.ts`),
  })),

  ...apps.map((app) => ({
    find: `@moh-sso/${app}/single-spa`,
    replacement: pathFromRoot(`./apps/${app}/src/single-spa.tsx`),
  })),

  ...apps.map((app) => ({
    find: new RegExp(`^@moh-sso/${app}/(.+)$`),
    replacement: pathFromRoot(`./apps/${app}/src/$1`),
  })),

  ...apps.map((app) => ({
    find: `@moh-sso/${app}`,
    replacement: pathFromRoot(`./apps/${app}/src/index.ts`),
  })),
];

export default defineConfig(({ mode }) => ({
  base: "/portal/",
  root: pathFromRoot("./apps/shell"),
  publicDir: pathFromRoot("./public"),

  plugins: [react()],

  define: {
    "process.env.NODE_ENV": JSON.stringify(mode === "production" ? "production" : "development"),
  },

  resolve: {
    dedupe: [
      "react",
      "react-dom",
      "react-router-dom",
      "react-redux",
      "@reduxjs/toolkit",
      "@carbon/react",
      "single-spa",
      "single-spa-react",
    ],

    alias: [
      {
        find: "@/config",
        replacement: pathFromRoot("./packages/config/src"),
      },
      {
        find: "@/types",
        replacement: pathFromRoot("./packages/types/src/global"),
      },
      ...dynamicAliases,
      {
        find: "@",
        replacement: pathFromRoot("./apps/shell/src"),
      },
    ],
  },

  server: {
    host: true,
    port: 3000,
    watch: {
      usePolling: true,
      interval: 300,
    },
  },

  optimizeDeps: {
    include: [
      "react",
      "react-dom",
      "react-dom/client",
      "react/jsx-runtime",
      "react/jsx-dev-runtime",
      "react-redux",
      "react-router-dom",
      "@reduxjs/toolkit",
      "@carbon/react",
      "@carbon/react/icons",
      "single-spa",
      "single-spa-react",
      "react-pivottable/PivotTableUI",
    ],
  },

  build: {
    outDir: pathFromRoot("./apps/shell/dist"),
    emptyOutDir: true,
    sourcemap: true,

    commonjsOptions: {
      include: [/react-pivottable/, /node_modules/],
      transformMixedEsModules: true,
    },
  },
}));
