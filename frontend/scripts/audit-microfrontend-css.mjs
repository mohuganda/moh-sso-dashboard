/* global process */
import { existsSync, readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";
import { check, finish, listStandaloneApps, root } from "./workspace-lib.mjs";

const failures = [];
const stagedMfPath = join(root, "dist", "mf");
const target = process.env.MF_CSS_AUDIT_TARGET ?? "apps";

function auditAppDist(app, distPath) {
  const assetsPath = join(distPath, "assets");
  const singleSpaPath = join(distPath, "single-spa.js");

  if (!existsSync(distPath)) {
    return;
  }

  const cssFiles = existsSync(assetsPath)
    ? readdirSync(assetsPath).filter((file) => file.endsWith(".css"))
    : [];

  if (cssFiles.length === 0) {
    return;
  }

  check(
    existsSync(singleSpaPath),
    `${app} has CSS assets but is missing dist/single-spa.js`,
    failures,
  );

  if (!existsSync(singleSpaPath)) {
    return;
  }

  const source = readFileSync(singleSpaPath, "utf8");
  check(
    source.includes("__mohCssFiles") && source.includes("data-moh-css"),
    `${app} has CSS assets but single-spa.js does not inject them`,
    failures,
  );
}

const useStagedDist = target === "staged";

check(
  target === "apps" || target === "staged",
  `unsupported MF_CSS_AUDIT_TARGET "${target}"`,
  failures,
);

if (useStagedDist) {
  check(
    existsSync(stagedMfPath),
    "staged microfrontend directory dist/mf does not exist",
    failures,
  );
}

for (const app of listStandaloneApps()) {
  auditAppDist(
    app,
    useStagedDist ? join(stagedMfPath, app) : join(root, "apps", app, "dist"),
  );
}

finish("microfrontend CSS audit", failures);
