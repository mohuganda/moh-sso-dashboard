import type { SystemAccess } from "@moh-sso/auth";
import type { Client, SystemNavigationItem } from "@moh-sso/types";

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
  if (
    system.launchMode === "internal" ||
    system.launchMode === "new_tab" ||
    system.launchMode === "same_tab"
  ) {
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

function getSystemSortOrder(system: SystemAccess): number {
  const order = Number(system.sortOrder);

  return Number.isFinite(order) ? order : Number.MAX_SAFE_INTEGER;
}

function isNavigationItem(value: unknown): value is SystemNavigationItem {
  if (!value || typeof value !== "object") {
    return false;
  }

  const item = value as Partial<SystemNavigationItem>;
  return typeof item.id === "string" && typeof item.label === "string";
}

export function parseSystemNavigation(navigation?: string): SystemNavigationItem[] {
  const rawNavigation = normalizePath(navigation);

  if (!rawNavigation) {
    return [];
  }

  try {
    const parsed: unknown = JSON.parse(rawNavigation);

    if (!Array.isArray(parsed)) {
      return [];
    }

    return parsed.filter(isNavigationItem);
  } catch {
    return [];
  }
}

function canAccessNavigationItem(
  item: SystemNavigationItem,
  can?: (permission: string) => boolean,
): boolean {
  if (!can) {
    return true;
  }

  if (item.permission && !can(item.permission)) {
    return false;
  }

  if (item.requiredPermissions?.some((permission) => !can(permission))) {
    return false;
  }

  if (
    item.requiredAnyPermissions?.length &&
    !item.requiredAnyPermissions.some((permission) => can(permission))
  ) {
    return false;
  }

  return true;
}

type NavigationEntry = {
  item: SystemNavigationItem;
  sequence: number;
};

function collectLauncherItems(
  items: SystemNavigationItem[],
  can?: (permission: string) => boolean,
  collected: NavigationEntry[] = [],
): NavigationEntry[] {
  for (const item of items) {
    if (!canAccessNavigationItem(item, can)) {
      continue;
    }

    if (item.displayInLauncher === true && item.path) {
      collected.push({ item, sequence: collected.length });
    }

    if (item.children?.length) {
      collectLauncherItems(item.children, can, collected);
    }
  }

  return collected;
}

function mapNavigationItemToClient(
  system: SystemAccess,
  item: SystemNavigationItem,
  sequence: number,
): Client {
  const systemType = inferSystemType(system);
  const launchMode = item.launchMode ?? inferLaunchMode(system, systemType);
  const parentOrder = getSystemSortOrder(system);
  const itemOrder = Number.isFinite(item.order) ? Number(item.order) : sequence;
  const effectiveOrder =
    parentOrder === Number.MAX_SAFE_INTEGER ? itemOrder : parentOrder * 1000 + itemOrder;
  const moduleClientId = `${system.clientId}:${item.id}`;

  return {
    id: moduleClientId,
    clientId: moduleClientId,
    name: item.label,
    description: item.description,
    enabled: Boolean(item.path),
    publicClient: false,
    rootUrl: item.path,
    baseUrl: item.path,
    redirectUris: [],
    roles: [],
    attributes: {
      "ui.icon": item.icon ?? system.icon ?? "",
      "ui.home": item.path,
      "ui.category":
        (typeof item.category === "string" ? item.category.trim() : "") || system.category || "",
      "ui.navigation": "",
      "ui.sidenav": "",
      "ui.systemType": systemType,
      "ui.displayInLauncher": "true",
      "ui.displayInSideNav": "false",
      "ui.launchMode": launchMode,
      "ui.order": String(effectiveOrder),
      "ui.entryType": "module",
      "ui.parentClientId": system.clientId,
      "ui.moduleId": item.id,
    },
  };
}

export function compareAccessibleSystems(a: SystemAccess, b: SystemAccess): number {
  const orderDiff = getSystemSortOrder(a) - getSystemSortOrder(b);

  if (orderDiff !== 0) {
    return orderDiff;
  }

  const nameDiff = (a.displayName || a.clientId).localeCompare(b.displayName || b.clientId);

  if (nameDiff !== 0) {
    return nameDiff;
  }

  return a.clientId.localeCompare(b.clientId);
}

export function mapAccessibleSystemToClient(system: SystemAccess): Client {
  const systemType = inferSystemType(system);
  const displayInLauncher = inferDisplayInLauncher(system);
  const displayInSideNav = inferDisplayInSideNav(system, systemType);
  const launchMode = inferLaunchMode(system, systemType);
  const sortOrder = getSystemSortOrder(system);

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
      "ui.authenticatedLaunchUrl": system.authenticatedLaunchUrl ?? "",
      "ui.category": system.category ?? "",
      "ui.navigation": normalizePath(system.navigation),
      "ui.sidenav": normalizePath(system.navigation),
      "ui.systemType": systemType,
      "ui.displayInLauncher": String(displayInLauncher),
      "ui.displayInSideNav": String(displayInSideNav),
      "ui.launchMode": launchMode,
      "ui.order": String(sortOrder),
      "ui.entryType": "system",
      "ui.parentClientId": system.clientId,
    },
  };
}

export function buildAccessibleClients({
  accessibleSystems,
}: {
  accessibleSystems: SystemAccess[];
}): Client[] {
  return [...accessibleSystems]
    .sort(compareAccessibleSystems)
    .filter((system) => system.clientId && system.launchUrl && (system.displayInLauncher ?? true))
    .map(mapAccessibleSystemToClient);
}

export function buildAccessibleLauncherEntries({
  accessibleSystems,
  can,
}: {
  accessibleSystems: SystemAccess[];
  can?: (permission: string) => boolean;
}): Client[] {
  const entries: Client[] = [];
  const seen = new Set<string>();

  for (const system of [...accessibleSystems].sort(compareAccessibleSystems)) {
    if (!system.clientId) {
      continue;
    }

    if (system.launchUrl && inferDisplayInLauncher(system)) {
      const client = mapAccessibleSystemToClient(system);
      const key = `${system.clientId}:${normalizePath(system.launchUrl)}`;

      if (!seen.has(key)) {
        seen.add(key);
        entries.push(client);
      }
    }

    const launcherItems = collectLauncherItems(parseSystemNavigation(system.navigation), can);

    for (const { item, sequence } of launcherItems) {
      const key = `${system.clientId}:${normalizePath(item.path)}`;

      if (seen.has(key)) {
        continue;
      }

      seen.add(key);
      entries.push(mapNavigationItemToClient(system, item, sequence));
    }
  }

  return entries;
}

export function buildAccessibleSideNavClients({
  accessibleSystems,
}: {
  accessibleSystems: SystemAccess[];
}): Client[] {
  return [...accessibleSystems]
    .sort(compareAccessibleSystems)
    .filter((system) => system.clientId)
    .map(mapAccessibleSystemToClient);
}
