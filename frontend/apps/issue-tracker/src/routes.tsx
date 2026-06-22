import type { MicrofrontendRoute } from "@moh-sso/microfrontend";

export const issueTrackerRoute: MicrofrontendRoute = {
  appName: "@moh-sso/issue-tracker",
  path: "/apps/dwh/issue-tracker",
  requiredAnyPermissions: ["data_quality:read", "issue_tracker:resolve"],
};
