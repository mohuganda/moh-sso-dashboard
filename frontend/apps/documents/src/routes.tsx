import type { MicrofrontendRoute } from "@moh-sso/microfrontend";

export const documentsRoute: MicrofrontendRoute = {
  appName: "@moh-sso/documents",
  path: "/apps/dwh/documents",
  requiredSystems: ["data-statistics"],
  requiredAnyPermissions: ["documents:read"],
};

export const documentsRoutes: MicrofrontendRoute[] = [documentsRoute];
