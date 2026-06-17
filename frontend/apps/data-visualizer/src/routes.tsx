import type { MicrofrontendRoute } from "@moh-sso/microfrontend";

export const dataVisualizerRoute: MicrofrontendRoute = {
  appName: "@moh-sso/data-visualizer",
  path: "/apps/dwh/data-visualizer",
  requiredAnyPermissions: ["documents:read", "data_quality:read", "surveillance:read"],
};
