import { AppMenu } from "@moh-sso/ui";

import { ConnectedAppGridContent } from "./ConnectedAppGridContent";

export function ConnectedAppMenu() {
  return (
    <AppMenu>
      {(closeMenu) => <ConnectedAppGridContent onSelect={closeMenu} />}
    </AppMenu>
  );
}
