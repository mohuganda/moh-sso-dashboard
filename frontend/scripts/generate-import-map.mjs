import { writeFileSync } from "node:fs";
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

const imports = {
  react: "/vendor/react.js",
  "react-dom": "/vendor/react-dom.js",
  "react-redux": "/vendor/react-redux.js",
  "single-spa": "/vendor/single-spa.js",
  "single-spa-react": "/vendor/single-spa-react.js",
};

for (const app of apps) {
  imports[`@moh-sso/${app}`] = `/mf/${app}/single-spa.js`;
}

for (const pkg of packages) {
  imports[`@moh-sso/${pkg}`] = `/packages/${pkg}/index.js`;
}

writeFileSync(join(root, "public", "import-map.json"), `${JSON.stringify({ imports }, null, 2)}\n`);
