/* global console, process */
import { existsSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const [mode, command, ...args] = process.argv.slice(2);
const validModes = new Set(["development", "production", "local-remote"]);

if (!validModes.has(mode) || !command) {
  console.error(
    "Usage: node scripts/with-runtime-config.mjs <development|production|local-remote> <command> [...args]",
  );
  process.exit(1);
}

const source = join(root, "public", `config.${mode}.js`);
const target = join(root, "public", "config.js");

if (!existsSync(source)) {
  throw new Error(`Runtime config source not found: ${source}`);
}

const previous = existsSync(target) ? readFileSync(target) : null;

try {
  writeFileSync(target, readFileSync(source));
  console.log(`Temporarily activated ${mode} runtime config.`);

  const result = spawnSync(command, args, {
    cwd: root,
    env: process.env,
    stdio: "inherit",
  });
  process.exitCode = result.status ?? 1;
} finally {
  if (previous !== null) {
    writeFileSync(target, previous);
    console.log("Restored previous public/config.js.");
  }
}
