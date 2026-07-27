/* global console */
import { join } from "node:path";
import {
  listPackages,
  listStandaloneApps,
  readJson,
  root,
} from "./workspace-lib.mjs";

const manifests = [
  ...listStandaloneApps().map((name) => join(root, "apps", name, "package.json")),
  ...listPackages().map((name) => join(root, "packages", name, "package.json")),
];

console.log(
  manifests
    .map(readJson)
    .map((pkg) => `${pkg.name}@${pkg.version}`)
    .sort()
    .join(" "),
);
