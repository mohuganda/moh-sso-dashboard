import type { MicrofrontendRoute } from "@moh-sso/microfrontend";

export const documentsRoutes: MicrofrontendRoute[] = [
  {
    appName: "@moh-sso/documents",
    path: "/apps/dwh/documents",
    paths: [
      "/portal/apps/dwh/documents",
      "/portal/apps/dwh/documents/:id",
      "/portal/apps/dwh/documents/:id/preview",
    ],
    requiredSystems: ["data-statistics"],
    requiredAnyPermissions: ["documents:read"],
  },
];
