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

  return defineConfig({
    plugins: [react()],
    resolve: {
      alias: [
        { find: "@moh-sso/api", replacement: pathFromFrontend("packages/api/src") },
        { find: "@moh-sso/auth", replacement: pathFromFrontend("packages/auth/src") },
        { find: "@moh-sso/config", replacement: pathFromFrontend("packages/config/src") },
        { find: "@moh-sso/microfrontend", replacement: pathFromFrontend("packages/microfrontend/src") },
        { find: "@moh-sso/state", replacement: pathFromFrontend("packages/state/src") },
        { find: "@moh-sso/types", replacement: pathFromFrontend("packages/types/src") },
        { find: "@moh-sso/ui", replacement: pathFromFrontend("packages/ui/src") },
        { find: "@moh-sso/utils", replacement: pathFromFrontend("packages/utils/src") },
        ...extraAliases,
      ],
    },
    build: {
      outDir: pathFromApp("./dist"),
      emptyOutDir: true,
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
        external: ["react", "react-dom", "react-redux", "single-spa", ...extraExternal],
      },
    },
  });
}
