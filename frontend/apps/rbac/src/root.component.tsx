import type { MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";

import RbacManagementPage from "./pages/rbac-management.page";

export function RbacRoot(props: MicrofrontendRuntimeProps) {
  void props;
  return <RbacManagementPage />;
}
