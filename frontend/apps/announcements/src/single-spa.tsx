import { createReactMicrofrontendLifecycle } from "@moh-sso/microfrontend";

import { AnnouncementsRoot } from "./root.component";

export const { bootstrap, mount, unmount } = createReactMicrofrontendLifecycle(AnnouncementsRoot);
