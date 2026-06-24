import type { SystemAccess } from "@moh-sso/auth";
import type { Client } from "@moh-sso/types";

export function mapAccessibleSystemToClient(system: SystemAccess): Client {
  return {
    id: system.clientId,
    clientId: system.clientId,
    name: system.displayName || system.clientId,
    description: system.category ? `${system.category} application` : undefined,
    enabled: Boolean(system.launchUrl),
    publicClient: false,
    rootUrl: system.launchUrl,
    baseUrl: system.launchUrl,
    redirectUris: [],
    roles: [],
    attributes: {
      "ui.icon": system.icon ?? "",
      "ui.home": system.launchUrl ?? "",
      "ui.category": system.category ?? "",
      "ui.navigation": system.navigation ?? "",
      "ui.sidenav": system.navigation ?? "",
      "ui.systemType": system.systemType,
      "ui.displayInLauncher": String(system.displayInLauncher),
      "ui.displayInSideNav": String(system.displayInSideNav),
      "ui.launchMode": system.launchMode,
    },
  };
}

export function buildAccessibleClients({
  accessibleSystems,
}: {
  accessibleSystems: SystemAccess[];
}): Client[] {
  return accessibleSystems
    .filter(
      (system) =>
        system.clientId &&
        system.launchUrl &&
        system.displayInLauncher,
    )
    .map(mapAccessibleSystemToClient);
}
