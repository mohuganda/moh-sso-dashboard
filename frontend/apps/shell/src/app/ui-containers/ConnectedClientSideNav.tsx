import { useSelector } from "react-redux";
import { useLocation, useNavigate } from "react-router-dom";

import { useAuthorization } from "@moh-sso/auth";
import { selectActiveClient, selectClients } from "@moh-sso/state";
import { ClientSideNav } from "@moh-sso/ui";

type ConnectedClientSideNavProps = {
  hasPermission?: (permission: string) => boolean;
};

export function ConnectedClientSideNav({ hasPermission }: ConnectedClientSideNavProps) {
  const clients = useSelector(selectClients);
  const activeClient = useSelector(selectActiveClient);
  const location = useLocation();
  const navigate = useNavigate();
  const { can, canLaunchSystem } = useAuthorization();
  const visibleClients = clients.filter((client) => client.clientId && canLaunchSystem(client.clientId));
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
