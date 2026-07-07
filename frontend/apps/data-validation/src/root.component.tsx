import type { MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";

import DataValidationPage from "./pages/data-validation.page";

export function DataValidationRoot(props: MicrofrontendRuntimeProps) {
  void props;

  return (
    <div className="moh-microfrontend-root">
      <DataValidationPage />
    </div>
  );
}
