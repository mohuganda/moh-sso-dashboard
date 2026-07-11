import type { MicrofrontendRoute } from "@moh-sso/microfrontend";

export const surveillanceRoute: MicrofrontendRoute = {
  appName: "@moh-sso/surveillance",
  path: "/apps/dwh/surveillance",
  requiredAnyPermissions: ["surveillance:read"],
  requiredSystems: ["data-statistics"],
};
