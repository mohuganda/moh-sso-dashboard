import { createReactMicrofrontendLifecycle } from "@moh-sso/microfrontend";

import { ClientsRoot } from "./root.component";

export const { bootstrap, mount, unmount } = createReactMicrofrontendLifecycle(ClientsRoot);
