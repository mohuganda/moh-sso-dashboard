import type { MicrofrontendRoute } from "@moh-sso/microfrontend";

export const rbacRoute: MicrofrontendRoute = {
  appName: "@moh-sso/rbac",
  path: "/admin/rbac",
  requiredPermissions: ["rbac:read"],
};
