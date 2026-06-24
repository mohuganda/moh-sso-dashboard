import type { MicrofrontendRoute } from "@moh-sso/microfrontend";

export const utilitiesRoute: MicrofrontendRoute = {
  appName: "@moh-sso/utilities",
  path: "/apps/utilities",
  requiredPermissions: ["portal:access"],
};
