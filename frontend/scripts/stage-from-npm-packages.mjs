/* global console */
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

function packageDist(name) {
  return join(root, "node_modules", "@moh-sso", name, "dist");
}

function assertRequired(path) {
  if (!existsSync(path)) {
    throw new Error(`Missing npm package build artifact: ${path}`);
  }
}

function copyDirectory(from, to) {
  if (!existsSync(from)) {
    throw new Error(`Missing npm package build directory: ${from}`);
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
  const dist = packageDist(app);
  assertRequired(join(dist, "single-spa.js"));
  copyDirectory(dist, join(root, "dist", "mf", app));
}

for (const packageName of packages) {
  const dist = packageDist(packageName);
  assertRequired(join(dist, "index.js"));
  copyDirectory(dist, join(root, "dist", "packages", packageName));
}

console.log("Staged microfrontend apps/packages from installed @moh-sso npm packages.");
