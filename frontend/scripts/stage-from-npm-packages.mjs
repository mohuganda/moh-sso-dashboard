/* global console, process */
import { copyFileSync, existsSync, mkdirSync, readFileSync, readdirSync, rmSync, statSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const modulesRoot = process.env.MOH_SSO_NPM_INSTALL_ROOT ?? join(root, "node_modules");
const basePath = (process.env.FRONTEND_BASE_PATH ?? "/").replace(/^\/?/, "/").replace(/\/$/, "");

const apps = [
  "announcements",
  "audit",
  "clients",
  "data-validation",
  "data-visualizer",
  "documents",
  "e-services",
  "email",
  "issue-tracker",
  "report-browser",
  "rbac",
  "surveillance",
  "users",
  "utilities",
];

const packages = ["api", "auth", "config", "microfrontend", "state", "types", "ui", "utils"];
const expectedPackages = [...apps, ...packages];
const requestedSpecs = (process.env.MOH_SSO_NPM_MODULES ?? "").trim().split(/\s+/).filter(Boolean);
const requestedVersions = new Map();

for (const spec of requestedSpecs) {
  const match = spec.match(/^@moh-sso\/([^@]+)@(\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?)$/);
  if (!match) {
    throw new Error(`npm deployment requires an exact SemVer package spec, received: ${spec}`);
  }
  requestedVersions.set(match[1], match[2]);
}

for (const name of expectedPackages) {
  if (!requestedVersions.has(name)) {
    throw new Error(`MOH_SSO_NPM_MODULES is missing an exact version for @moh-sso/${name}`);
  }
}

function packageDist(name) {
  return join(modulesRoot, "@moh-sso", name, "dist");
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

assertRequired(join(root, "apps", "shell", "dist", "index.html"));
copyDirectory(join(root, "apps", "shell", "dist"), join(root, "dist"));

for (const app of apps) {
  const dist = packageDist(app);
  const installed = JSON.parse(
    readFileSync(join(modulesRoot, "@moh-sso", app, "package.json"), "utf8"),
  );
  if (installed.version !== requestedVersions.get(app)) {
    throw new Error(`Installed ${installed.name}@${installed.version}, expected ${requestedVersions.get(app)}`);
  }
  assertRequired(join(dist, "single-spa.js"));
  copyDirectory(dist, join(root, "dist", "mf", app));
}

for (const packageName of packages) {
  const dist = packageDist(packageName);
  const installed = JSON.parse(
    readFileSync(join(modulesRoot, "@moh-sso", packageName, "package.json"), "utf8"),
  );
  if (installed.version !== requestedVersions.get(packageName)) {
    throw new Error(
      `Installed ${installed.name}@${installed.version}, expected ${requestedVersions.get(packageName)}`,
    );
  }
  assertRequired(join(dist, "index.js"));
  copyDirectory(dist, join(root, "dist", "packages", packageName));
}

if (basePath && basePath !== "/") {
  const baseDir = join(root, "dist", basePath.slice(1));
  rmSync(baseDir, { recursive: true, force: true });
  mkdirSync(baseDir, { recursive: true });

  for (const entry of [
    "assets",
    "mf",
    "packages",
    "config.js",
    "config.production.js",
    "import-map.json",
    "import-map.local.json",
    "index.html",
    "logo.png",
    "version-manifest.json",
  ]) {
    const source = join(root, "dist", entry);
    if (!existsSync(source)) {
      continue;
    }

    const target = join(baseDir, entry);
    if (statSync(source).isDirectory()) {
      copyDirectory(source, target);
    } else {
      mkdirSync(dirname(target), { recursive: true });
      copyFileSync(source, target);
    }
  }
}

console.log("Staged microfrontend apps/packages from installed @moh-sso npm packages.");
