/* global process */

import { existsSync, readFileSync } from "node:fs";
import { join } from "node:path";
import { check, finish, root } from "./workspace-lib.mjs";

const manifestPath = join(root, ".artifacts", "npm", "pack-manifest.json");
const failures = [];
const maxTarballBytes = Number(process.env.NPM_PACKAGE_MAX_TARBALL_BYTES ?? 10 * 1024 * 1024);
const maxUnpackedBytes = Number(process.env.NPM_PACKAGE_MAX_UNPACKED_BYTES ?? 40 * 1024 * 1024);

check(
  existsSync(manifestPath),
  ".artifacts/npm/pack-manifest.json is missing; run npm run pack:all",
  failures,
);
check(
  Number.isFinite(maxTarballBytes) && maxTarballBytes > 0,
  "NPM_PACKAGE_MAX_TARBALL_BYTES must be a positive number",
  failures,
);
check(
  Number.isFinite(maxUnpackedBytes) && maxUnpackedBytes > 0,
  "NPM_PACKAGE_MAX_UNPACKED_BYTES must be a positive number",
  failures,
);

if (existsSync(manifestPath)) {
  const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
  const forbidden = [
    /(^|\/)\.env(?:\.|$)/,
    /(^|\/)\.npmrc$/,
    /(^|\/)node_modules\//,
    /(^|\/)src\//,
    /(^|\/)(?:coverage|tests?|__tests__)\//,
    /(?:^|\/)(?:config\.development|config\.local-remote)\.js$/,
    /\.(?:log|tgz)$/,
  ];

  check(
    manifest.packages?.length === 22,
    `expected 22 publishable packages, found ${manifest.packages?.length ?? 0}`,
    failures,
  );

  for (const pkg of manifest.packages ?? []) {
    const paths = new Set(pkg.files.map((file) => file.path));
    check(
      existsSync(join(root, ".artifacts", "npm", pkg.tarball)),
      `${pkg.name}: tarball is missing`,
      failures,
    );
    check(
      pkg.size <= maxTarballBytes,
      `${pkg.name}: tarball is ${(pkg.size / 1024 / 1024).toFixed(2)} MiB; limit is ${(maxTarballBytes / 1024 / 1024).toFixed(2)} MiB`,
      failures,
    );
    check(
      pkg.unpackedSize <= maxUnpackedBytes,
      `${pkg.name}: unpacked size is ${(pkg.unpackedSize / 1024 / 1024).toFixed(2)} MiB; limit is ${(maxUnpackedBytes / 1024 / 1024).toFixed(2)} MiB`,
      failures,
    );
    check(paths.has("package.json"), `${pkg.name}: package.json is missing`, failures);
    check(paths.has("README.md"), `${pkg.name}: README.md is missing`, failures);
    check(paths.has("dist/index.js"), `${pkg.name}: dist/index.js is missing`, failures);
    check(paths.has("dist/index.d.ts"), `${pkg.name}: dist/index.d.ts is missing`, failures);

    if (pkg.workspace.startsWith("apps/")) {
      check(paths.has("dist/single-spa.js"), `${pkg.name}: single-spa.js is missing`, failures);
      check(paths.has("dist/routes.js"), `${pkg.name}: routes.js is missing`, failures);
    }

    for (const path of paths) {
      if (forbidden.some((pattern) => pattern.test(path))) {
        failures.push(`${pkg.name}: forbidden tarball entry ${path}`);
      }
    }
  }
}

finish("Tarball verification", failures);
