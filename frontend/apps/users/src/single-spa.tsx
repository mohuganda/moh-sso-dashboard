import { createReactMicrofrontendLifecycle } from "@moh-sso/microfrontend";

import { UsersRoot } from "./root.component";

export const { bootstrap, mount, unmount } = createReactMicrofrontendLifecycle(UsersRoot);
