import type { MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";

import MyTimeSheet from "./pages/my-timesheet.component";

export function UtilitiesRoot(props: MicrofrontendRuntimeProps) {
  void props;

  return <MyTimeSheet />;
}
