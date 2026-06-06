import { copyFileSync, existsSync, mkdirSync } from "node:fs";
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

function copyRequired(from, to) {
  if (!existsSync(from)) {
    throw new Error(`Missing build artifact: ${from}`);
  }

  mkdirSync(dirname(to), { recursive: true });
  copyFileSync(from, to);
}

for (const app of apps) {
  copyRequired(
    join(root, "apps", app, "dist", "single-spa.js"),
    join(root, "dist", "mf", app, "single-spa.js"),
  );
}

for (const pkg of packages) {
  copyRequired(
    join(root, "packages", pkg, "dist", "index.js"),
    join(root, "dist", "packages", pkg, "index.js"),
  );
}

copyRequired(
  join(root, "public", "version-manifest.json"),
  join(root, "dist", "version-manifest.json"),
);
