/* global console, process */
import { existsSync, readFileSync } from "node:fs";
import { join } from "node:path";
import {
  listPackages,
  listStandaloneApps,
  root,
} from "./workspace-lib.mjs";

const failures = [];
const semverPattern = /^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$/;

function readJson(path) {
  return JSON.parse(readFileSync(path, "utf8"));
}

function assert(condition, path, message) {
  if (!condition) {
    failures.push(`${path}: ${message}`);
  }
}

function hasDistExport(pkg, key, file) {
  const exportEntry = pkg.exports?.[key];
  return exportEntry?.types === `./dist/${file}.d.ts` && exportEntry?.import === `./dist/${file}.js`;
}

function auditPackageJson(kind, name) {
  const baseDir = join(root, kind === "app" ? "apps" : "packages", name);
  const relativePath = `${kind === "app" ? "apps" : "packages"}/${name}/package.json`;
  const pkg = readJson(join(baseDir, "package.json"));

  assert(pkg.name?.startsWith("@moh-sso/"), relativePath, "name must start with @moh-sso/");
  assert(semverPattern.test(pkg.version ?? ""), relativePath, "version must be valid SemVer");
  assert(pkg.private === false, relativePath, "private must be false");
  assert(pkg.type === "module", relativePath, "type must be module");
  assert(Boolean(pkg.scripts?.build), relativePath, "build script is required");
  assert(pkg.main === "./dist/index.js", relativePath, "main must point to ./dist/index.js");
  assert(pkg.module === "./dist/index.js", relativePath, "module must point to ./dist/index.js");
  assert(pkg.types === "./dist/index.d.ts", relativePath, "types must point to ./dist/index.d.ts");
  assert(hasDistExport(pkg, ".", "index"), relativePath, "exports[\".\"] must point to dist index files");
  assert(
    Array.isArray(pkg.files) &&
      pkg.files.includes("dist") &&
      pkg.files.includes("README.md") &&
      pkg.files.includes("package.json"),
    relativePath,
    "files must include dist, README.md, and package.json",
  );
  assert(!pkg.files?.includes("src"), relativePath, "files must not include src");
  assert(JSON.stringify(pkg.exports ?? {}).includes("./src") === false, relativePath, "exports must not point to src");
  assert(pkg.publishConfig?.access === "restricted", relativePath, "publishConfig.access must be restricted");
  assert(
    pkg.publishConfig?.registry === "https://registry.npmjs.org/",
    relativePath,
    "publishConfig.registry must use npmjs",
  );
  assert(pkg.license === "UNLICENSED", relativePath, "license must be UNLICENSED for restricted packages");
  assert(pkg.repository?.type === "git", relativePath, "repository.type must be git");
  assert(
    pkg.repository?.url === "git+https://github.com/mohuganda/moh-sso-dashboard.git",
    relativePath,
    "repository.url must reference the canonical repository",
  );
  assert(
    pkg.repository?.directory === `${kind === "app" ? "frontend/apps" : "frontend/packages"}/${name}`,
    relativePath,
    "repository.directory must identify the workspace",
  );
  assert(existsSync(join(baseDir, "README.md")), relativePath, "README.md is missing");

  assert(existsSync(join(baseDir, "dist", "index.js")), relativePath, "dist/index.js is missing; run build first");
  assert(existsSync(join(baseDir, "dist", "index.d.ts")), relativePath, "dist/index.d.ts is missing; run build first");

  if (kind === "app") {
    assert(hasDistExport(pkg, "./single-spa", "single-spa"), relativePath, "apps must export ./single-spa from dist");
    assert(hasDistExport(pkg, "./routes", "routes"), relativePath, "apps must export ./routes from dist");
    assert(existsSync(join(baseDir, "dist", "single-spa.js")), relativePath, "dist/single-spa.js is missing");
    assert(existsSync(join(baseDir, "dist", "single-spa.d.ts")), relativePath, "dist/single-spa.d.ts is missing");
    assert(existsSync(join(baseDir, "dist", "routes.js")), relativePath, "dist/routes.js is missing");
    assert(existsSync(join(baseDir, "dist", "routes.d.ts")), relativePath, "dist/routes.d.ts is missing");

    for (const peer of ["react", "react-dom", "react-redux", "single-spa"]) {
      assert(Boolean(pkg.peerDependencies?.[peer]), relativePath, `apps must declare ${peer} as a peer dependency`);
    }
  }
}

for (const app of listStandaloneApps()) {
  auditPackageJson("app", app);
}

for (const packageName of listPackages()) {
  auditPackageJson("package", packageName);
}

if (failures.length > 0) {
  console.error("Package publishability audit failed:");
  for (const failure of failures) {
    console.error(`  - ${failure}`);
  }
  process.exit(1);
}

console.log("All publishable frontend apps/packages are npm-ready.");
