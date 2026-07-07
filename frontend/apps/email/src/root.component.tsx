import type { MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";

import EmailOutbox from "./pages/email-outbox.component";

export function EmailRoot(props: MicrofrontendRuntimeProps) {
  void props;

  return (
    <div className="moh-microfrontend-root">
      <EmailOutbox />
    </div>
  );
}
