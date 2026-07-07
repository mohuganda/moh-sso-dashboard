import type { MicrofrontendRoute } from "@moh-sso/microfrontend";

export const issueTrackerRoute: MicrofrontendRoute = {
  appName: "@moh-sso/issue-tracker",
  path: "/apps/dwh/issue-tracker",
  requiredPermissions: ["issue_tracker:read"],
  requiredSystems: ["data-statistics"],
};
