import { fileURLToPath, URL } from "node:url";
import { defineMicrofrontendConfig } from "../../build/vite.microfrontend.config";

const pathFromFrontend = (path: string) => fileURLToPath(new URL(`../../${path}`, import.meta.url));

export default defineMicrofrontendConfig({
  appUrl: import.meta.url,
  name: "MohSsoUsers",
  extraAliases: [{ find: "@moh-sso/clients", replacement: pathFromFrontend("apps/clients/src") }],
  extraExternal: ["@moh-sso/clients"]
});
