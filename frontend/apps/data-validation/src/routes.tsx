import type { MicrofrontendRoute } from "@moh-sso/microfrontend";

export const dataValidationRoute: MicrofrontendRoute = {
  appName: "@moh-sso/data-validation",
  path: "/apps/dwh/data-validation",
  requiredPermissions: ["data_quality:read"],
  requiredSystems: ["data-statistics"],
};
