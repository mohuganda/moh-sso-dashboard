import type { MicrofrontendRoute } from "@moh-sso/microfrontend";

export const documentsRoute: MicrofrontendRoute = {
  appName: "@moh-sso/documents",
  path: "/apps/dwh/filesvr",
  paths: [
    "/apps/dwh/filesvr",
    "/apps/utilities/self-service/eservice/document-upload",
  ],
  requiredAnyPermissions: ["documents:read", "documents:write"],
};
