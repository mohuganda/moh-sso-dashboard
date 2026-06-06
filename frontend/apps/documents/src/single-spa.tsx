import { createReactMicrofrontendLifecycle } from "@moh-sso/microfrontend";

import { DocumentsRoot } from "./root.component";

export const { bootstrap, mount, unmount } = createReactMicrofrontendLifecycle(DocumentsRoot);
