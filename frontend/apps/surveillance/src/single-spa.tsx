import { createReactMicrofrontendLifecycle } from "@moh-sso/microfrontend";

import { SurveillanceRoot } from "./root.component";

export const { bootstrap, mount, unmount } = createReactMicrofrontendLifecycle(SurveillanceRoot);
