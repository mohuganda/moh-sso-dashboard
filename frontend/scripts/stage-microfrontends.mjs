import { copyFileSync, existsSync, mkdirSync, readdirSync, rmSync } from "node:fs";
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

function assertRequired(path) {
  if (!existsSync(path)) {
    throw new Error(`Missing build artifact: ${path}`);
  }
}

function copyDirectory(from, to) {
  if (!existsSync(from)) {
    throw new Error(`Missing build directory: ${from}`);
  }

  rmSync(to, { recursive: true, force: true });
  mkdirSync(to, { recursive: true });

  for (const entry of readdirSync(from, { withFileTypes: true })) {
    const source = join(from, entry.name);
    const target = join(to, entry.name);

    if (entry.isDirectory()) {
      if (entry.name !== "node_modules") {
        copyDirectory(source, target);
      }
      continue;
    }

    copyFileSync(source, target);
  }
}

for (const app of apps) {
  assertRequired(join(root, "apps", app, "dist", "single-spa.js"));
  copyDirectory(
    join(root, "apps", app, "dist"),
    join(root, "dist", "mf", app),
  );
}

for (const pkg of packages) {
  assertRequired(join(root, "packages", pkg, "dist", "index.js"));
  copyDirectory(
    join(root, "packages", pkg, "dist"),
    join(root, "dist", "packages", pkg),
  );
}

copyRequired(
  join(root, "public", "version-manifest.json"),
  join(root, "dist", "version-manifest.json"),
);
