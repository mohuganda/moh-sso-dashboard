import { readFileSync } from "node:fs";
import { join } from "node:path";
import {
  check,
  fileExists,
  finish,
  listStandaloneApps,
  readJson,
  rel,
  root,
  sourceFiles,
} from "./workspace-lib.mjs";

const failures = [];
const apps = listStandaloneApps();
const importMap = readJson(join(root, "public", "import-map.json"));
const imports = importMap.imports ?? {};

for (const app of apps) {
  const appRoot = join(root, "apps", app);
  const packageJsonPath = join(appRoot, "package.json");
  const pkg = readJson(packageJsonPath);
  const expectedName = `@moh-sso/${app}`;

  check(pkg.name === expectedName, `${rel(packageJsonPath)} should be named ${expectedName}`, failures);
  check(pkg.private === false, `${rel(packageJsonPath)} should be publishable with private=false`, failures);
  check(fileExists(join(appRoot, "src", "index.ts")), `${app} is missing src/index.ts`, failures);
  check(fileExists(join(appRoot, "src", "root.component.tsx")), `${app} is missing src/root.component.tsx`, failures);
  check(fileExists(join(appRoot, "src", "routes.tsx")), `${app} is missing src/routes.tsx`, failures);
  check(fileExists(join(appRoot, "src", "single-spa.tsx")), `${app} is missing src/single-spa.tsx`, failures);
  check(fileExists(join(appRoot, "vite.config.ts")), `${app} is missing vite.config.ts`, failures);
  check(Boolean(pkg.exports?.["."]), `${app} package is missing exports["."]`, failures);
  check(Boolean(pkg.exports?.["./single-spa"]), `${app} package is missing exports["./single-spa"]`, failures);
  check(Boolean(pkg.exports?.["./routes"]), `${app} package is missing exports["./routes"]`, failures);
  check(Boolean(imports[expectedName]), `public/import-map.json is missing ${expectedName}`, failures);
}

const forbiddenAliases = /@\/(features|store|shared|lib|ui|utils)\b/;
for (const file of sourceFiles([join(root, "apps"), join(root, "packages")])) {
  const source = readFileSync(file, "utf8");
  if (forbiddenAliases.test(source)) {
    failures.push(`${rel(file)} imports through a retired compatibility alias`);
  }
}

finish("Microfrontend doctor", failures);
