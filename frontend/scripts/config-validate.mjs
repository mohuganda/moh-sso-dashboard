import { readFileSync } from "node:fs";
import { join } from "node:path";
import { check, fileExists, finish, readJson, root } from "./workspace-lib.mjs";

const failures = [];
const configFiles = ["config.js", "config.development.js", "config.production.js", "config.local-remote.js"];

for (const file of configFiles) {
  const path = join(root, "public", file);
  check(fileExists(path), `public/${file} is missing`, failures);
  if (fileExists(path)) {
    const source = readFileSync(path, "utf8");
    check(source.includes("window.__APP_CONFIG__"), `public/${file} does not define window.__APP_CONFIG__`, failures);
    check(source.includes("API_BASE_URL"), `public/${file} is missing API_BASE_URL`, failures);
    check(source.includes("microfrontendMode"), `public/${file} is missing microfrontendMode`, failures);
    check(source.includes("singleSpaOrchestration"), `public/${file} is missing singleSpaOrchestration`, failures);
    check(source.includes("microfrontendMountMode"), `public/${file} is missing microfrontendMountMode`, failures);
  }
}

const production = readFileSync(join(root, "public", "config.production.js"), "utf8");
check(production.includes('microfrontendMode: "remote"'), "production config should use remote microfrontend mode", failures);
check(production.includes("singleSpaOrchestration: true"), "production config should enable single-spa orchestration", failures);
check(production.includes('microfrontendMountMode: "orchestrated"'), "production config should use orchestrated mount mode", failures);

readJson(join(root, "public", "import-map.json"));
readJson(join(root, "public", "version-manifest.json"));

finish("Frontend config validation", failures);
