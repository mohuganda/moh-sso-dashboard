import { createReactMicrofrontendLifecycle } from "@moh-sso/microfrontend";

import { UtilitiesRoot } from "./root.component";

export const { bootstrap, mount, unmount } = createReactMicrofrontendLifecycle(UtilitiesRoot);
