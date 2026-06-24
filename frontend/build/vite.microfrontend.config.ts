import react from "@vitejs/plugin-react";
import { fileURLToPath, URL } from "node:url";
import { defineConfig, type Alias } from "vite";

type MicrofrontendConfigOptions = {
  appUrl: string;
  name: string;
  extraAliases?: Alias[];
  extraExternal?: string[];
};

export function defineMicrofrontendConfig({
  appUrl,
  name,
  extraAliases = [],
  extraExternal = [],
}: MicrofrontendConfigOptions) {
  const pathFromApp = (path: string) => fileURLToPath(new URL(path, appUrl));
  const pathFromFrontend = (path: string) => fileURLToPath(new URL(`../../${path}`, appUrl));

  return defineConfig(({ mode }) => ({
    plugins: [react()],

    // Solves the esbuild vs rollup primitive evaluation conflicts
    define: {
      "process.env.NODE_ENV": JSON.stringify(mode),
      "process.env": JSON.stringify({ NODE_ENV: mode }),
      process: JSON.stringify({ env: { NODE_ENV: mode } }),
    },

    resolve: {
      alias: [
        { find: "@moh-sso/api", replacement: pathFromFrontend("packages/api/src") },
        { find: "@moh-sso/auth", replacement: pathFromFrontend("packages/auth/src") },
        { find: "@moh-sso/config", replacement: pathFromFrontend("packages/config/src") },
        {
          find: "@moh-sso/microfrontend",
          replacement: pathFromFrontend("packages/microfrontend/src"),
        },
        { find: "@moh-sso/state", replacement: pathFromFrontend("packages/state/src") },
        { find: "@moh-sso/types", replacement: pathFromFrontend("packages/types/src") },
        { find: "@moh-sso/ui", replacement: pathFromFrontend("packages/ui/src") },
        { find: "@moh-sso/utils", replacement: pathFromFrontend("packages/utils/src") },
        ...extraAliases,
      ],
    },

    build: {
      target: "es2020",
      outDir: pathFromApp("./dist"),
      emptyOutDir: true,
      sourcemap: true,

      commonjsOptions: {
        include: [/react-pivottable/, /node_modules/],
        transformMixedEsModules: true,
      },

      lib: {
        entry: {
          index: pathFromApp("./src/index.ts"),
          routes: pathFromApp("./src/routes.tsx"),
          "single-spa": pathFromApp("./src/single-spa.tsx"),
        },
        name,
        formats: ["es"],
        fileName: (_format, entryName) => `${entryName}.js`,
      },

      rollupOptions: {
        external: [
          "react",
          "react-dom",
          "react-dom/client", // Prevents internal React bundle pollution
          "react-redux",
          "react-router-dom",
          "@reduxjs/toolkit",
          "@carbon/react",
          "@carbon/react/icons",
          "single-spa",
          ...extraExternal,
        ],

        output: {
          entryFileNames: "[name].js",
          chunkFileNames: "assets/[name]-[hash].js",
          assetFileNames: "assets/[name]-[hash][extname]",
        },
      },
    },
  }));
}
