/* global console, process */
import { spawnSync } from "node:child_process";
import { join } from "node:path";
import { listPackages, listStandaloneApps, readJson, root } from "./workspace-lib.mjs";

function run(command, args, options = {}) {
  return spawnSync(command, args, {
    cwd: root,
    encoding: "utf8",
    env: process.env,
    ...options,
  });
}

const status = run("git", ["status", "--porcelain"]);
if (status.status !== 0 || status.stdout.trim()) {
  throw new Error("Refusing to publish from a dirty worktree");
}

if (process.env.CI === "true" && process.env.GITHUB_REPOSITORY !== "mohuganda/moh-sso-dashboard") {
  throw new Error("Refusing to publish outside mohuganda/moh-sso-dashboard");
}

if (process.env.CI === "true" && process.env.GITHUB_REF_NAME !== "main") {
  throw new Error(`Refusing to publish from CI branch ${process.env.GITHUB_REF_NAME ?? "unknown"}`);
}

const registry = run("npm", ["config", "get", "registry"]);
if (registry.status !== 0 || registry.stdout.trim() !== "https://registry.npmjs.org/") {
  throw new Error(
    `npm registry must be https://registry.npmjs.org/, got ${registry.stdout.trim()}`,
  );
}

const tag = process.env.NPM_TAG ?? "latest";
const allowedTags = new Set(["latest", "next", "beta"]);
if (!allowedTags.has(tag)) {
  throw new Error(`NPM_TAG must be one of ${[...allowedTags].join(", ")}, got ${tag}`);
}

const packageJsons = [
  ...listStandaloneApps().map((name) => join(root, "apps", name, "package.json")),
  ...listPackages().map((name) => join(root, "packages", name, "package.json")),
];
const prereleases = packageJsons.map(readJson).filter((pkg) => pkg.version.includes("-"));

if (tag === "latest" && prereleases.length > 0) {
  throw new Error(
    `Refusing to publish prerelease versions with latest: ${prereleases.map((pkg) => `${pkg.name}@${pkg.version}`).join(", ")}`,
  );
}

const releaseCheck = run("npm", ["run", "release:check"], { stdio: "inherit" });
if (releaseCheck.status !== 0) {
  throw new Error("Release checks failed; no packages were published");
}

const result = run("npx", ["changeset", "publish", "--tag", tag], { stdio: "inherit" });
if (result.status !== 0) {
  process.exit(result.status ?? 1);
}

console.log(`Published npm packages with dist-tag ${tag}.`);
