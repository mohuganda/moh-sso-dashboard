import type { SystemAccess } from "@moh-sso/auth";
import type { Client } from "@moh-sso/types";

function normalizePath(path?: string): string {
  const trimmed = path?.trim();
  if (!trimmed) {
    return "";
  }
  return trimmed;
}

function inferSystemType(system: SystemAccess): "platform" | "external" {
  if (system.systemType === "platform" || system.systemType === "external") {
    return system.systemType;
  }

  const launchUrl = normalizePath(system.launchUrl);
  const navigation = normalizePath(system.navigation);

  if (/^https?:\/\//i.test(launchUrl) && !navigation) {
    return "external";
  }

  return "platform";
}

function inferLaunchMode(system: SystemAccess, systemType: "platform" | "external") {
  if (system.launchMode === "internal" || system.launchMode === "new_tab" || system.launchMode === "same_tab") {
    return system.launchMode;
  }

  return systemType === "external" ? "new_tab" : "internal";
}

function inferDisplayInLauncher(system: SystemAccess): boolean {
  return system.displayInLauncher ?? true;
}

function inferDisplayInSideNav(system: SystemAccess, systemType: "platform" | "external"): boolean {
  if (system.displayInSideNav !== undefined) {
    return system.displayInSideNav;
  }

  if (systemType === "external") {
    return false;
  }

  return normalizePath(system.navigation) !== "";
}

export function mapAccessibleSystemToClient(system: SystemAccess): Client {
  const systemType = inferSystemType(system);
  const displayInLauncher = inferDisplayInLauncher(system);
  const displayInSideNav = inferDisplayInSideNav(system, systemType);
  const launchMode = inferLaunchMode(system, systemType);

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
      "ui.navigation": normalizePath(system.navigation),
      "ui.sidenav": normalizePath(system.navigation),
      "ui.systemType": systemType,
      "ui.displayInLauncher": String(displayInLauncher),
      "ui.displayInSideNav": String(displayInSideNav),
      "ui.launchMode": launchMode,
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
        (system.displayInLauncher ?? true),
    )
    .map(mapAccessibleSystemToClient);
}
