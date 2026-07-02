import { useSelector } from "react-redux";
import { useLocation, useNavigate } from "react-router-dom";

import { PERMISSIONS, useAuthorization } from "@moh-sso/auth";
import { DEFAULT_CLIENT_ID, selectActiveClient, selectClients } from "@moh-sso/state";
import { ClientSideNav } from "@moh-sso/ui";

// These clients are local frontend-only modules with no Keycloak backing.
// They don't appear in ihp_systems so canLaunchSystem's hasSystem check
// always fails for them. Show them to anyone with systems:launch permission.
const LOCAL_FRONTEND_CLIENT_IDS = new Set([
  "__utilities__",
  "__settings__",
  "eservices",
  "research-studies",
  "case-registers",
  "outbreak-management",
  "reference-registers",
]);

type ConnectedClientSideNavProps = {
  hasPermission?: (permission: string) => boolean;
};

export function ConnectedClientSideNav({ hasPermission }: ConnectedClientSideNavProps) {
  const clients = useSelector(selectClients);
  const activeClient = useSelector(selectActiveClient);
  const location = useLocation();
  const navigate = useNavigate();
  const { can, canAny, canLaunchSystem } = useAuthorization();
  const canAccessDataStatistics = canAny([
    PERMISSIONS.dataQualityRead,
    PERMISSIONS.documentsRead,
    PERMISSIONS.reportBrowserRead,
    PERMISSIONS.surveillanceRead,
  ]);
  const visibleClients = clients.filter((client) => {
    if (!client.clientId) {
      return false;
    }

    if (client.clientId === DEFAULT_CLIENT_ID) {
      return canAccessDataStatistics;
    }

    if (LOCAL_FRONTEND_CLIENT_IDS.has(client.clientId)) {
      return can(PERMISSIONS.systemsLaunch);
    }

    return canLaunchSystem(client.clientId);
  });
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
