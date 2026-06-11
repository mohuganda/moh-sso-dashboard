import { useSelector } from "react-redux";
import { useLocation, useNavigate } from "react-router-dom";

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

  return (
    <ClientSideNav
      clients={clients}
      activeClient={activeClient}
      currentPath={location.pathname}
      hasPermission={hasPermission}
      onNavigate={navigate}
    />
  );
}
