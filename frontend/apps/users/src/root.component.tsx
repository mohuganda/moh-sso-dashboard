import type { MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";

import UsersPage from "./pages/user.component";

export function UsersRoot(props: MicrofrontendRuntimeProps) {
  void props;

  return <UsersPage />;
}
