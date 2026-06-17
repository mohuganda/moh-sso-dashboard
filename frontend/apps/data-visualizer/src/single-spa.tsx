import { createReactMicrofrontendLifecycle } from "@moh-sso/microfrontend";

import { DataVisualizerRoot } from "./root.component";

export const { bootstrap, mount, unmount } =
  createReactMicrofrontendLifecycle(DataVisualizerRoot);
