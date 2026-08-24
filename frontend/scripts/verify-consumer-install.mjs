/* global console, process */
import { existsSync, mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { spawnSync } from "node:child_process";
import { root } from "./workspace-lib.mjs";

const artifactsDir = join(root, ".artifacts", "npm");
const manifestPath = join(artifactsDir, "pack-manifest.json");
const consumerDir = process.env.NPM_CONSUMER_DIR ?? join(tmpdir(), "moh-sso-npm-consumer");
const npmCache = process.env.NPM_CONFIG_CACHE ?? join(tmpdir(), "moh-sso-npm-cache");

if (!existsSync(manifestPath)) {
  throw new Error("Pack manifest is missing; run npm run pack:all first");
}

const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
const tarballs = manifest.packages.map((pkg) => join(artifactsDir, pkg.tarball));

rmSync(consumerDir, { recursive: true, force: true });
mkdirSync(consumerDir, { recursive: true });
writeFileSync(
  join(consumerDir, "package.json"),
  `${JSON.stringify({ name: "moh-sso-npm-consumer", version: "1.0.0", private: true, type: "module" }, null, 2)}\n`,
);

const install = spawnSync(
  "npm",
  [
    "install",
    "--ignore-scripts",
    "--no-audit",
    "--no-fund",
    "--package-lock=false",
    ...tarballs,
  ],
  {
    cwd: consumerDir,
    encoding: "utf8",
    env: { ...process.env, NPM_CONFIG_CACHE: npmCache },
  },
);

if (install.status !== 0) {
  console.error(install.stdout);
  console.error(install.stderr);
  throw new Error("Clean consumer installation failed");
}

for (const pkg of manifest.packages) {
  const installed = join(consumerDir, "node_modules", ...pkg.name.split("/"), "package.json");
  if (!existsSync(installed)) {
    throw new Error(`${pkg.name} was not installed in the clean consumer`);
  }
}

console.log(`Clean consumer installed all ${manifest.packages.length} packages at ${consumerDir}.`);
