import { createReactMicrofrontendLifecycle } from "@moh-sso/microfrontend";

import { RbacRoot } from "./root.component";

export const { bootstrap, mount, unmount } = createReactMicrofrontendLifecycle(RbacRoot);
