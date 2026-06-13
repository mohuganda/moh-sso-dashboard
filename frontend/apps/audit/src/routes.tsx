import type { MicrofrontendRoute } from "@moh-sso/microfrontend";

export const auditRoute: MicrofrontendRoute = {
  appName: "@moh-sso/audit",
  path: "/admin/audit-logs",
  requiredPermissions: ["audit:read"],
};
