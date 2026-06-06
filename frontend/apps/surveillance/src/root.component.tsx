import type { MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";

import SurveillanceDashboardPage from "./pages/surveillance.component";

export function SurveillanceRoot(props: MicrofrontendRuntimeProps) {
  void props;

  return <SurveillanceDashboardPage />;
}
