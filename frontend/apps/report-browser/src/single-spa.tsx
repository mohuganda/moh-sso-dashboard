import { createReactMicrofrontendLifecycle } from "@moh-sso/microfrontend";

import { ReportBrowserRoot } from "./root.component";

export const { bootstrap, mount, unmount } = createReactMicrofrontendLifecycle(ReportBrowserRoot);
