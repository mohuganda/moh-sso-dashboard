import { join } from "node:path";
import {
  check,
  finish,
  listApps,
  listPackages,
  readJson,
  root,
} from "./workspace-lib.mjs";

const singletonPackages = [
  "@carbon/react",
  "react",
  "react-dom",
  "react-redux",
  "react-router-dom",
  "single-spa",
  "single-spa-react",
];
const failures = [];
const ranges = new Map();

for (const [folder, names] of [["apps", listApps()], ["packages", listPackages()]]) {
  for (const name of names) {
    const pkg = readJson(join(root, folder, name, "package.json"));
    for (const singleton of singletonPackages) {
      const range =
        pkg.peerDependencies?.[singleton] ??
        pkg.dependencies?.[singleton] ??
        pkg.devDependencies?.[singleton];
      if (!range) continue;

      const existing = ranges.get(singleton);
      if (existing && existing.range !== range) {
        failures.push(
          `${pkg.name} declares ${singleton}@${range}, but ${existing.name} declares ${existing.range}`,
        );
      } else if (!existing) {
        ranges.set(singleton, { name: pkg.name, range });
      }
    }
  }
}

const shell = readJson(join(root, "apps", "shell", "package.json"));
check(shell.private === true, "@moh-sso/shell must remain private", failures);
finish("Package compatibility check", failures);
