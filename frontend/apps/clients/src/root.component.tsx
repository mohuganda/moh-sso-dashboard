import type { MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";

import ClientsPage from "./pages/client.component";

export function ClientsRoot(props: MicrofrontendRuntimeProps) {
  void props;

  return (
    <div className="moh-microfrontend-root">
      <ClientsPage />
    </div>
  );
}
