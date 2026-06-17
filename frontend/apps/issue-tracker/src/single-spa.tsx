import { createReactMicrofrontendLifecycle } from "@moh-sso/microfrontend";

import { IssueTrackerRoot } from "./root.component";

export const { bootstrap, mount, unmount } = createReactMicrofrontendLifecycle(IssueTrackerRoot);
