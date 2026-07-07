import type { MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";

import AuditLogsPage from "./pages/audit_component";

export function AuditRoot(props: MicrofrontendRuntimeProps) {
  void props;

  return (
    <div className="moh-microfrontend-root">
      <AuditLogsPage />
    </div>
  );
}
