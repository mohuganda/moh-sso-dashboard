import react from "@vitejs/plugin-react";
import { readdirSync } from "node:fs";
import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";

const pathFromRoot = (path: string) => fileURLToPath(new URL(path, import.meta.url));

const getDirectories = (source: string) =>
  readdirSync(pathFromRoot(source), { withFileTypes: true })
    .filter((dirent) => dirent.isDirectory())
    .map((dirent) => dirent.name);

const apps = getDirectories("./apps");
const packages = getDirectories("./packages");

const dynamicAliases = [
  ...packages.map((pkg) => ({
    find: `@moh-sso/${pkg}`,
    replacement: pathFromRoot(`./packages/${pkg}/src`),
  })),
  ...apps.map((app) => ({
    find: `@moh-sso/${app}/single-spa`,
    replacement: pathFromRoot(`./apps/${app}/src/single-spa.tsx`),
  })),
  ...apps.map((app) => ({
    find: `@moh-sso/${app}`,
    replacement: pathFromRoot(`./apps/${app}/src`),
  })),
];

export default defineConfig({
  base: "/portal/",
  root: pathFromRoot("./apps/shell"),
  publicDir: pathFromRoot("./public"),
  plugins: [react()],
  resolve: {
    alias: [
      { find: "@/config", replacement: pathFromRoot("./packages/config/src") },
      { find: "@/types", replacement: pathFromRoot("./packages/types/src/global") },
      ...dynamicAliases,
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
    rollupOptions: {
      external: [
        "react",
        "react-dom",
        "react-redux",
        "react-router-dom",
        "@reduxjs/toolkit",
        "@carbon/react",
        "@carbon/react/icons",
        "single-spa",
        "single-spa-react",
      ],
    },
  },
});

