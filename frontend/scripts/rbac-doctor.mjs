import { readFileSync } from "node:fs";
import { join } from "node:path";
import { check, fileExists, finish, listStandaloneApps, rel, root } from "./workspace-lib.mjs";

const failures = [];
const allowedWithoutMetadata = new Set([]);

for (const app of listStandaloneApps()) {
  const routeFile = join(root, "apps", app, "src", "routes.tsx");
  check(fileExists(routeFile), `${app} is missing src/routes.tsx`, failures);
  if (!fileExists(routeFile)) {
    continue;
  }

  const source = readFileSync(routeFile, "utf8");
  const hasPermissionMetadata =
    source.includes("requiredPermissions") || source.includes("requiredAnyPermissions");
  const hasSystemMetadata = source.includes("requiredSystems");

  if (!allowedWithoutMetadata.has(app)) {
    check(
      hasPermissionMetadata || hasSystemMetadata,
      `${rel(routeFile)} should define RBAC route metadata`,
      failures,
    );
  }
}

finish("RBAC doctor", failures);
