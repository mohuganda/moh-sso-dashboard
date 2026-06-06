import { createReactMicrofrontendLifecycle } from "@moh-sso/microfrontend";

import { EmailRoot } from "./root.component";

export const { bootstrap, mount, unmount } = createReactMicrofrontendLifecycle(EmailRoot);
