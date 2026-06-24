import type { MicrofrontendRoute } from "@moh-sso/microfrontend";

export const reportBrowserRoute: MicrofrontendRoute = {
  appName: "@moh-sso/report-browser",
  path: "/apps/dwh/reports",
  requiredPermissions: ["report_browser:read"],
  requiredSystems: ["data-statistics"],
};
