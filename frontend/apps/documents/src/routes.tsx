import type { MicrofrontendRoute } from "@moh-sso/microfrontend";

export const documentsRoutes: MicrofrontendRoute[] = [
  {
    appName: "@moh-sso/documents",
    path: "/apps/dwh/filesvr",
    paths: ["/apps/dwh/filesvr", "/apps/dwh/filesvr/*"],
    requiredAnyPermissions: ["documents:read"],
  },
  {
    appName: "@moh-sso/documents",
    path: "/apps/utilities/self-service/eservice/document-upload",
    paths: ["/apps/utilities/self-service/eservice/document-upload"],
    requiredAnyPermissions: ["documents:write"],
  },
];
