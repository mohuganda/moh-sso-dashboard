import type { MicrofrontendRoute } from "@moh-sso/microfrontend";

export const reportBrowserRoute: MicrofrontendRoute = {
  appName: "@moh-sso/report-browser",
  path: "/apps/dwh/dashboards",
  paths: ["/apps/dwh/dashboards"],
  requiredPermissions: ["report_browser:read"],
  requiredSystems: ["data-statistics"],
};
