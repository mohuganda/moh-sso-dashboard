import { fileURLToPath, URL } from "node:url";
import { defineMicrofrontendConfig } from "../../build/vite.microfrontend.config";

const pathFromFrontend = (path: string) => fileURLToPath(new URL(`../../${path}`, import.meta.url));

export default defineMicrofrontendConfig({
  appUrl: import.meta.url,
  name: "MohSsoIssueTracker",
  extraAliases: [{ find: "@moh-sso/data-visualizer", replacement: pathFromFrontend("apps/data-visualizer/src") }],
  extraExternal: ["@moh-sso/data-visualizer"]
});
