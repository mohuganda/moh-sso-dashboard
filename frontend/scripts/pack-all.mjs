/* global console, process */
import { mkdirSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { spawnSync } from "node:child_process";
import {
  listPackages,
  listStandaloneApps,
  readJson,
  root,
} from "./workspace-lib.mjs";

const artifactsDir = join(root, ".artifacts", "npm");
const npmCache = process.env.NPM_CONFIG_CACHE ?? join(tmpdir(), "moh-sso-npm-cache");
const workspaces = [
  ...listPackages().map((name) => join("packages", name)),
  ...listStandaloneApps().map((name) => join("apps", name)),
];

rmSync(artifactsDir, { recursive: true, force: true });
mkdirSync(artifactsDir, { recursive: true });

const packages = [];

for (const workspace of workspaces) {
  const packageJson = readJson(join(root, workspace, "package.json"));
  const result = spawnSync(
    "npm",
    ["pack", "--json", "--pack-destination", artifactsDir],
    {
      cwd: join(root, workspace),
      encoding: "utf8",
      env: { ...process.env, NPM_CONFIG_CACHE: npmCache },
    },
  );

  if (result.status !== 0) {
    console.error(result.stdout);
    console.error(result.stderr);
    throw new Error(`npm pack failed for ${packageJson.name}`);
  }

  const [packed] = JSON.parse(result.stdout);
  packages.push({
    name: packageJson.name,
    version: packageJson.version,
    workspace,
    tarball: packed.filename,
    size: packed.size,
    unpackedSize: packed.unpackedSize,
    files: packed.files.map(({ path, size }) => ({ path, size })),
  });
  console.log(`Packed ${packageJson.name}@${packageJson.version}`);
}

writeFileSync(
  join(artifactsDir, "pack-manifest.json"),
  `${JSON.stringify({ generatedAt: new Date().toISOString(), packages }, null, 2)}\n`,
);

console.log(`Packed ${packages.length} publishable workspaces into .artifacts/npm.`);
