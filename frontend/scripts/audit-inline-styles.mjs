/* global console, process */
import { readdirSync, readFileSync, statSync } from "node:fs";
import path from "node:path";

const root = process.cwd();
const scanRoots = ["apps", "packages"];
const ignoredDirs = new Set(["dist", "node_modules", ".git", ".husky"]);
const sourceExtensions = new Set([".ts", ".tsx"]);
const storyPattern = /\.stories\.(ts|tsx)$/;
const inlineStylePattern = /style=\{|\bstyle:\s*\{/g;

const matches = [];

function walk(dir) {
  const entries = readdirSync(dir, { withFileTypes: true });

  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name);

    if (entry.isDirectory()) {
      if (!ignoredDirs.has(entry.name)) {
        walk(fullPath);
      }
      continue;
    }

    if (!entry.isFile()) {
      continue;
    }

    const ext = path.extname(entry.name);
    if (!sourceExtensions.has(ext) || storyPattern.test(entry.name)) {
      continue;
    }

    const content = readFileSync(fullPath, "utf8");
    const lines = content.split(/\r?\n/);

    lines.forEach((line, index) => {
      if (inlineStylePattern.test(line)) {
        matches.push({
          file: path.relative(root, fullPath),
          line: index + 1,
          content: line.trim(),
        });
      }
      inlineStylePattern.lastIndex = 0;
    });
  }
}

for (const scanRoot of scanRoots) {
  const absoluteRoot = path.join(root, scanRoot);
  try {
    if (statSync(absoluteRoot).isDirectory()) {
      walk(absoluteRoot);
    }
  } catch {
    // Workspace roots are optional in package-level runs.
  }
}

if (matches.length === 0) {
  console.log("No inline style usage found in app/package runtime source.");
  process.exit(0);
}

console.log(`Found ${matches.length} inline style reference(s) in app/package runtime source:`);
for (const match of matches) {
  console.log(`${match.file}:${match.line}: ${match.content}`);
}

console.log(
  "\nMove static styles to SCSS. Keep inline styles only for runtime-computed values that cannot be represented safely in CSS classes.",
);
