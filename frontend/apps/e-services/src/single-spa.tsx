import { createReactMicrofrontendLifecycle } from "@moh-sso/microfrontend";

import { EServicesRoot } from "./root.component";

export const { bootstrap, mount, unmount } = createReactMicrofrontendLifecycle(EServicesRoot);
