import { createReactMicrofrontendLifecycle } from "@moh-sso/microfrontend";

import { AuditRoot } from "./root.component";

export const { bootstrap, mount, unmount } = createReactMicrofrontendLifecycle(AuditRoot);
