/* global console, process */
import { copyFileSync, existsSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const mode = process.argv[2];
const validModes = new Set(["development", "production", "local-remote"]);

if (!validModes.has(mode)) {
  console.error("Usage: node scripts/use-runtime-config.mjs <development|production|local-remote>");
  process.exit(1);
}

const source = join(root, "public", `config.${mode}.js`);
const target = join(root, "public", "config.js");

if (!existsSync(source)) {
  console.error(`Runtime config source not found: ${source}`);
  process.exit(1);
}

copyFileSync(source, target);
console.log(`Activated ${mode} runtime config: public/config.js`);
