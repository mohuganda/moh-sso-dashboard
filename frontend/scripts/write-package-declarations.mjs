import { mkdirSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import process from "node:process";

const [, , kind, name] = process.argv;

if (!kind || !name || !["app", "package"].includes(kind)) {
  throw new Error("Usage: node scripts/write-package-declarations.mjs <app|package> <name>");
}

const baseDir = kind === "app" ? join("apps", name) : join("packages", name);
const distDir = join(baseDir, "dist");

mkdirSync(distDir, { recursive: true });

if (kind === "app") {
  writeFileSync(join(distDir, "index.d.ts"), 'export * from "../src/index";\n');
  writeFileSync(join(distDir, "routes.d.ts"), 'export * from "../src/routes";\n');
  writeFileSync(join(distDir, "single-spa.d.ts"), 'export * from "../src/single-spa";\n');
} else {
  writeFileSync(join(distDir, "index.d.ts"), 'export * from "../src/index";\n');
}
