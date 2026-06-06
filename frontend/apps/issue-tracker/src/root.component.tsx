import type { MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";

import IssueTracker from "./pages/issue-tracker";

export function IssueTrackerRoot(props: MicrofrontendRuntimeProps) {
  void props;

  return <IssueTracker />;
}
