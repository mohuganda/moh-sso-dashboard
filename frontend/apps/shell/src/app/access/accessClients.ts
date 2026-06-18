import type { SystemAccess } from "@moh-sso/auth";
import type { Client } from "@moh-sso/types";

const DATA_STATISTICS_CLIENT_ID = "__default__";

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
      "ui.icon": system.icon,
      "ui.home": system.launchUrl,
      "ui.category": system.category,
    },
  };
}

export function buildAccessibleClients({
  accessibleSystems,
  includeDataStatistics,
}: {
  accessibleSystems: SystemAccess[];
  includeDataStatistics: boolean;
}): Client[] {
  const clients = accessibleSystems
    .filter((system) => system.clientId && system.launchUrl)
    .map(mapAccessibleSystemToClient);

  if (!includeDataStatistics) {
    return clients;
  }

  const hasDataStatistics = clients.some(
    (client) =>
      client.clientId === DATA_STATISTICS_CLIENT_ID ||
      client.baseUrl === "/apps/dwh" ||
      client.baseUrl === "/portal/apps/dwh",
  );

  if (hasDataStatistics) {
    return clients;
  }

  return [
    {
      id: DATA_STATISTICS_CLIENT_ID,
      clientId: DATA_STATISTICS_CLIENT_ID,
      name: "Data & Statistics",
      description: "Data quality, document upload, reports, and surveillance tools.",
      enabled: true,
      publicClient: false,
      rootUrl: "/apps/dwh",
      baseUrl: "/apps/dwh",
      redirectUris: [],
      roles: [],
      attributes: {
        "ui.icon": "dashboard",
        "ui.home": "/apps/dwh",
        "ui.order": "0",
        "ui.category": "platform",
      },
    },
    ...clients,
  ];
}
