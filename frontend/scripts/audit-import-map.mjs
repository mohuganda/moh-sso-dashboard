/* global console, process */
import { existsSync, readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";

const root = process.cwd();
const importMapPath = join(root, "public", "import-map.json");
const scanRoots = [join(root, "dist", "mf"), join(root, "dist", "packages")];

function walkJsFiles(dir, files = []) {
  if (!existsSync(dir)) {
    return files;
  }

  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const path = join(dir, entry.name);

    if (entry.isDirectory()) {
      walkJsFiles(path, files);
      continue;
    }

    if (entry.name.endsWith(".js")) {
      files.push(path);
    }
  }

  return files;
}

function isBareSpecifier(specifier) {
  return (
    specifier &&
    !specifier.startsWith(".") &&
    !specifier.startsWith("/") &&
    !specifier.startsWith("http://") &&
    !specifier.startsWith("https://")
  );
}

function collectSpecifiers(file) {
  const source = readFileSync(file, "utf8");
  const specifiers = new Set();

  for (const line of source.split(/\r?\n/)) {
    const trimmed = line.trim();

    if (trimmed.startsWith("import ")) {
      const sideEffectMatch = trimmed.match(/^import\s+["']([^"']+)["']/);
      const fromMatch = trimmed.match(/\sfrom\s+["']([^"']+)["']/);
      const specifier = sideEffectMatch?.[1] ?? fromMatch?.[1];

      if (isBareSpecifier(specifier)) {
        specifiers.add(specifier);
      }
    }
  }

  for (const match of source.matchAll(/\bimport\(\s*["']([^"']+)["']\s*\)/g)) {
    if (isBareSpecifier(match[1])) {
      specifiers.add(match[1]);
    }
  }

  return specifiers;
}

if (!existsSync(importMapPath)) {
  console.error(`Missing import map: ${importMapPath}`);
  process.exit(1);
}

const importMap = JSON.parse(readFileSync(importMapPath, "utf8"));
const imports = importMap.imports ?? {};
const missing = new Map();

for (const scanRoot of scanRoots) {
  for (const file of walkJsFiles(scanRoot)) {
    for (const specifier of collectSpecifiers(file)) {
      if (!imports[specifier]) {
        const files = missing.get(specifier) ?? [];
        files.push(file.replace(`${root}/`, ""));
        missing.set(specifier, files);
      }
    }
  }
}

if (missing.size > 0) {
  console.error("Missing import-map entries:");

  for (const [specifier, files] of [...missing.entries()].sort(([a], [b]) => a.localeCompare(b))) {
    console.error(`\n${specifier}`);
    for (const file of files.slice(0, 5)) {
      console.error(`  - ${file}`);
    }
    if (files.length > 5) {
      console.error(`  - ...and ${files.length - 5} more`);
    }
  }

  process.exit(1);
}

console.log("All staged bare imports are present in public/import-map.json.");
