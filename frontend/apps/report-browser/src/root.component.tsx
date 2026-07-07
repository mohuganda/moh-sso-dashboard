import type { MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";

import ReportBrowser from "./pages/report-broswer";

export function ReportBrowserRoot(props: MicrofrontendRuntimeProps) {
  void props;

  return (
    <div className="moh-microfrontend-root">
      <ReportBrowser />
    </div>
  );
}
