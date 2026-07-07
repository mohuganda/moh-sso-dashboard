import type { MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";

import IssueTracker from "./pages/issue-tracker.component";

export function IssueTrackerRoot(props: MicrofrontendRuntimeProps) {
  void props;

  return (
    <div className="moh-microfrontend-root">
      <IssueTracker />
    </div>
  );
}
