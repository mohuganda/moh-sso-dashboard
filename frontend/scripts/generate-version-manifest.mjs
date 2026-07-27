/* global console, process */
import { createHash } from "node:crypto";
import { execFileSync } from "node:child_process";
import { existsSync, readdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const modulesRoot = process.env.MOH_SSO_NPM_INSTALL_ROOT ?? join(root, "node_modules");

function gitCommit() {
  if (process.env.GIT_COMMIT) return process.env.GIT_COMMIT;
  try {
    return execFileSync("git", ["rev-parse", "--short=12", "HEAD"], {
      cwd: dirname(root),
      encoding: "utf8",
    }).trim();
  } catch {
    return "unknown";
  }
}

function importMapHash() {
  const path = join(root, "public", "import-map.json");
  if (!existsSync(path)) return "unknown";
  return createHash("sha256").update(readFileSync(path)).digest("hex");
}

function readJson(path) {
  return JSON.parse(readFileSync(path, "utf8"));
}

function collectPackages(group) {
  const groupPath = join(root, group);
  const entries = {};

  if (!existsSync(groupPath)) {
    return entries;
  }

  for (const dirent of readdirSync(groupPath, { withFileTypes: true })) {
    if (!dirent.isDirectory()) {
      continue;
    }

    const workspacePackagePath = join(groupPath, dirent.name, "package.json");
    const npmPackagePath = join(modulesRoot, "@moh-sso", dirent.name, "package.json");
    const packagePath =
      process.env.FRONTEND_MODULE_SOURCE === "npm" &&
      dirent.name !== "shell" &&
      existsSync(npmPackagePath)
        ? npmPackagePath
        : workspacePackagePath;
    if (!existsSync(packagePath)) {
      continue;
    }

    const packageJson = readJson(packagePath);
    entries[dirent.name] = {
      name: packageJson.name,
      version: packageJson.version,
    };
  }

  return Object.fromEntries(Object.entries(entries).sort(([a], [b]) => a.localeCompare(b)));
}

const appVersions = collectPackages("apps");
const { shell, ...standaloneApps } = appVersions;
const manifest = {
  build: {
    commit: gitCommit(),
    buildTime: process.env.BUILD_TIME ?? new Date().toISOString(),
    deploymentMode: process.env.FRONTEND_MODULE_SOURCE ?? "source",
    importMapSha256: importMapHash(),
  },
  shell,
  apps: standaloneApps,
  packages: collectPackages("packages"),
};

writeFileSync(
  join(root, "public", "version-manifest.json"),
  `${JSON.stringify(manifest, null, 2)}\n`,
);

console.log("Generated public/version-manifest.json");
