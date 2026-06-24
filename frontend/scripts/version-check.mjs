import { join } from "node:path";
import { check, fileExists, finish, listPackages, listStandaloneApps, readJson, root } from "./workspace-lib.mjs";

const failures = [];
const semverPattern = /^\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?$/;

for (const group of [
  ["apps", listStandaloneApps()],
  ["packages", listPackages()],
]) {
  const [folder, names] = group;
  for (const name of names) {
    const path = join(root, folder, name, "package.json");
    const pkg = readJson(path);
    check(pkg.name === `@moh-sso/${name}`, `${folder}/${name}/package.json has unexpected package name`, failures);
    check(semverPattern.test(pkg.version ?? ""), `${folder}/${name}/package.json has invalid version ${pkg.version ?? ""}`, failures);
    check(pkg.private === false, `${folder}/${name}/package.json should have private=false for npm deployment`, failures);
    check(Array.isArray(pkg.files) && pkg.files.includes("dist"), `${folder}/${name}/package.json should publish dist`, failures);
    check(Boolean(pkg.publishConfig?.access), `${folder}/${name}/package.json should define publishConfig.access`, failures);
  }
}

check(fileExists(join(root, ".changeset", "config.json")), ".changeset/config.json is missing", failures);
finish("Version tooling check", failures);
