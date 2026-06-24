import { createReactMicrofrontendLifecycle } from "@moh-sso/microfrontend";

import { DataValidationRoot } from "./root.component";

export const { bootstrap, mount, unmount } =
  createReactMicrofrontendLifecycle(DataValidationRoot);
