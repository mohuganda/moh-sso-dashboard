import { useSelector } from "react-redux";
import { useLocation, useNavigate } from "react-router-dom";

import { useAuthorization } from "@moh-sso/auth";
import { selectActiveClient } from "@moh-sso/state";
import { ClientSideNav } from "@moh-sso/ui";
import { mapAccessibleSystemToClient } from "@/app/access/accessClients";

type ConnectedClientSideNavProps = {
  hasPermission?: (permission: string) => boolean;
};

export function ConnectedClientSideNav({ hasPermission }: ConnectedClientSideNavProps) {
  const activeClient = useSelector(selectActiveClient);
  const location = useLocation();
  const navigate = useNavigate();
  const { accessibleSystems, can } = useAuthorization();
  const visibleClients = accessibleSystems.map(mapAccessibleSystemToClient);
  const checkPermission = hasPermission ?? ((permission: string) => can(permission as never));

  return (
    <ClientSideNav
      clients={visibleClients}
      activeClient={activeClient}
      currentPath={location.pathname}
      hasPermission={checkPermission}
      onNavigate={navigate}
    />
  );
}
