/* global process */
import { readFileSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = dirname(dirname(fileURLToPath(import.meta.url)));

const apps = [
  "announcements",
  "audit",
  "clients",
  "data-visualizer",
  "documents",
  "e-services",
  "email",
  "issue-tracker",
  "report-browser",
  "surveillance",
  "users",
  "utilities",
];

const packages = ["api", "auth", "config", "microfrontend", "state", "types", "ui", "utils"];
const versionedImports = process.env.FRONTEND_VERSIONED_IMPORTS === "true";
const assetBaseUrl = (process.env.FRONTEND_ASSET_BASE_URL ?? "").replace(/\/$/, "");

function readVersion(group, name) {
  const packageJson = JSON.parse(readFileSync(join(root, group, name, "package.json"), "utf8"));
  return packageJson.version;
}

function appUrl(app) {
  if (!versionedImports || !assetBaseUrl) {
    return `/mf/${app}/single-spa.js`;
  }

  return `${assetBaseUrl}/mf/${app}/${readVersion("apps", app)}/single-spa.js`;
}

function packageUrl(pkg) {
  if (!versionedImports || !assetBaseUrl) {
    return `/packages/${pkg}/index.js`;
  }

  return `${assetBaseUrl}/packages/${pkg}/${readVersion("packages", pkg)}/index.js`;
}

const imports = {
  react: "https://esm.sh/react@19.2.0",
  "react-dom": "https://esm.sh/react-dom@19.2.0",
  "react-dom/client": "https://esm.sh/react-dom@19.2.0/client",
  "react-redux": "https://esm.sh/react-redux@9.2.0",
  "single-spa": "https://esm.sh/single-spa@6.0.3",
  "single-spa-react": "https://esm.sh/single-spa-react@6.0.2",
};

for (const app of apps) {
  imports[`@moh-sso/${app}`] = appUrl(app);
}

for (const pkg of packages) {
  imports[`@moh-sso/${pkg}`] = packageUrl(pkg);
}

writeFileSync(join(root, "public", "import-map.json"), `${JSON.stringify({ imports }, null, 2)}\n`);
