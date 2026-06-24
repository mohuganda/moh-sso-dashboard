/* global console, process */
import { existsSync, readFileSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import { dirname } from "node:path";
import { fileURLToPath } from "node:url";

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const distDir = join(root, "dist");
const basePath = (process.env.FRONTEND_BASE_PATH ?? "/").replace(/^\/?/, "/").replace(/\/$/, "");
const importMapPath = join(distDir, "import-map.json");

function readImportMap() {
  if (!existsSync(importMapPath)) {
    throw new Error(`Missing import map: ${importMapPath}`);
  }

  return JSON.stringify(JSON.parse(readFileSync(importMapPath, "utf8")), null, 2);
}

function inlineImportMap(indexPath, importMapJson) {
  if (!existsSync(indexPath)) {
    return false;
  }

  const html = readFileSync(indexPath, "utf8");
  const inlineScript = `<script type="importmap">\n${importMapJson}\n    </script>`;
  const externalImportMapPattern =
    /<script\s+type=["']importmap["']\s+src=["'][^"']+["']\s*>\s*<\/script>/;

  let nextHtml = html;

  if (nextHtml.includes("<!-- import-map:generated -->")) {
    nextHtml = nextHtml.replace("<!-- import-map:generated -->", inlineScript);
  } else if (externalImportMapPattern.test(nextHtml)) {
    nextHtml = nextHtml.replace(externalImportMapPattern, inlineScript);
  } else if (!nextHtml.includes('type="importmap"')) {
    nextHtml = nextHtml.replace("</head>", `    ${inlineScript}\n  </head>`);
  }

  writeFileSync(indexPath, nextHtml);
  return true;
}

const importMapJson = readImportMap();
const targets = [join(distDir, "index.html")];

if (basePath && basePath !== "/") {
  targets.push(join(distDir, basePath.slice(1), "index.html"));
}

const updated = targets.filter((target) => inlineImportMap(target, importMapJson));

if (updated.length === 0) {
  throw new Error("No index.html files were found for import-map injection.");
}

console.log(`Inlined production import map into ${updated.length} HTML file(s).`);
