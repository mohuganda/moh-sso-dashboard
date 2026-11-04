import { PageShell } from "@moh-sso/ui";
import type { MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";

import ReportBrowser from "./pages/report-broswer";

export function ReportBrowserRoot(props: MicrofrontendRuntimeProps) {
  void props;

  return (
    <div className="moh-microfrontend-root">
      <PageShell
        title="Report Browser"
        subtitle="Browse health dashboards and reports."
      >
        <ReportBrowser />
      </PageShell>
    </div>
  );
}
