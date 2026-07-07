import type { MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";

import { AnnouncementsPage } from "./pages/announcements.components";

export function AnnouncementsRoot(props: MicrofrontendRuntimeProps) {
  void props;

  return (
    <div className="moh-microfrontend-root">
      <AnnouncementsPage />
    </div>
  );
}
