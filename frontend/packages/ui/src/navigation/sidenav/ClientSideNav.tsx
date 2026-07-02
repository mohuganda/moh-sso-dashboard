import { useMemo } from "react";
import { SideNav, SideNavItems, SideNavLink, SideNavMenu } from "@carbon/react";

import type { Client } from "@moh-sso/types";

type SideNavItem = {
  id: string;
  label: string;
  path?: string;
  permission?: string;
  order?: number;
  children?: SideNavItem[];
};

/* --------------------------------
 * Helpers
 * -------------------------------- */

function normalizePath(path?: string): string {
  if (!path) {
    return "";
  }

  const trimmed = path.trim();

  if (!trimmed) {
    return "";
  }

  return trimmed.startsWith("/") ? trimmed : `/${trimmed}`;
}

function isActivePath(currentPath: string, itemPath?: string): boolean {
  const normalizedItemPath = normalizePath(itemPath);

  if (!normalizedItemPath) {
    return false;
  }

  return currentPath === normalizedItemPath || currentPath.startsWith(`${normalizedItemPath}/`);
}

function sortItems(items: SideNavItem[]): SideNavItem[] {
  return [...items].sort((a, b) => {
    const aOrder = typeof a.order === "number" ? a.order : Number.MAX_SAFE_INTEGER;
    const bOrder = typeof b.order === "number" ? b.order : Number.MAX_SAFE_INTEGER;

    if (aOrder !== bOrder) {
      return aOrder - bOrder;
    }

    return a.label.localeCompare(b.label);
  });
}

function isValidSideNavItem(value: unknown): value is SideNavItem {
  if (!value || typeof value !== "object") {
    return false;
  }

  const item = value as Partial<SideNavItem>;

  return typeof item.id === "string" && typeof item.label === "string";
}

function normalizeLegacySideNavItems(items: SideNavItem[]): SideNavItem[] {
  const normalizedItems = items
    .filter((item) => item.id !== "data-exports" && normalizePath(item.path) !== "/apps/dwh/exports")
    .map((item) => {
      const children = item.children ? normalizeLegacySideNavItems(item.children) : undefined;
      const itemPath = normalizePath(item.path);

      if (itemPath === "/apps/dwh" || itemPath === "/apps/dwh/reports" || itemPath === "/apps/dwh/dashboards") {
        return {
          ...item,
          id: "dashboards",
          label: "Dashboards",
          path: "/apps/dwh/dashboards",
          children,
        };
      }

      if (itemPath === "/apps/dwh/filesvr") {
        return {
          ...item,
          id: item.id === "file-svr" ? "documents" : item.id,
          label: item.label === "File Upload" ? "Documents" : item.label,
          path: "/apps/dwh/documents",
          children,
        };
      }

      return {
        ...item,
        children,
      };
    });

  const seen = new Set<string>();

  return normalizedItems.filter((item) => {
    const key = normalizePath(item.path) || item.id;

    if (seen.has(key)) {
      return false;
    }

    seen.add(key);
    return true;
  });
}

function parseSideNav(client: Client): SideNavItem[] {
  const rawSideNav = client.attributes?.["ui.sidenav"];

  if (!rawSideNav) {
    return [];
  }

  try {
    const parsed = JSON.parse(rawSideNav);

    if (!Array.isArray(parsed)) {
      return [];
    }

    return normalizeLegacySideNavItems(parsed.filter(isValidSideNavItem));
  } catch {
    return [];
  }
}

function isPlatformSideNavClient(client: Client): boolean {
  const systemType = client.attributes?.["ui.systemType"];
  const displayInSideNav = client.attributes?.["ui.displayInSideNav"];
  const navigation = parseSideNav(client);

  if (systemType === "external") {
    return false;
  }

  if (displayInSideNav === "true") {
    return navigation.length > 0;
  }

  if (displayInSideNav === "false") {
    return false;
  }

  return navigation.length > 0;
}

function hasActiveChild(item: SideNavItem, currentPath: string): boolean {
  if (isActivePath(currentPath, item.path)) {
    return true;
  }

  return item.children?.some((child) => hasActiveChild(child, currentPath)) ?? false;
}

function filterByPermission(
  items: SideNavItem[],
  hasPermission?: (permission: string) => boolean,
): SideNavItem[] {
  return items
    .filter((item) => {
      if (!item.permission) {
        return true;
      }

      return hasPermission?.(item.permission) ?? true;
    })
    .map((item) => ({
      ...item,
      children: item.children ? filterByPermission(item.children, hasPermission) : undefined,
    }))
    .filter((item) => {
      const hasChildren = item.children && item.children.length > 0;
      const hasPath = Boolean(normalizePath(item.path));

      return hasPath || hasChildren;
    });
}

function getClientLabel(client: Client): string {
  return client.name || client.clientId || "Application";
}

/* --------------------------------
 * Recursive item
 * -------------------------------- */

type RenderSideNavItemProps = {
  item: SideNavItem;
  currentPath: string;
  onNavigate: (path: string) => void;
};

function RenderSideNavItem({ item, currentPath, onNavigate }: RenderSideNavItemProps) {
  const children = sortItems(item.children ?? []);
  const hasChildren = children.length > 0;
  const itemPath = normalizePath(item.path);
  const active = isActivePath(currentPath, itemPath);
  const childActive = hasActiveChild(item, currentPath);

  if (hasChildren) {
    return (
      <SideNavMenu
        key={item.id}
        title={item.label}
        defaultExpanded={childActive}
        isActive={childActive}
      >
        {children.map((child) => (
          <RenderSideNavItem
            key={child.id}
            item={child}
            currentPath={currentPath}
            onNavigate={onNavigate}
          />
        ))}
      </SideNavMenu>
    );
  }

  if (!itemPath) {
    return null;
  }

  return (
    <SideNavLink
      key={item.id}
      isActive={active}
      aria-current={active ? "page" : undefined}
      onClick={() => onNavigate(itemPath)}
    >
      {item.label}
    </SideNavLink>
  );
}

/* --------------------------------
 * Component
 * -------------------------------- */

type ClientSideNavProps = {
  clients?: Client[];
  activeClient?: Client | null;
  currentPath: string;
  onNavigate: (path: string) => void;
  hasPermission?: (permission: string) => boolean;
};

export function ClientSideNav({
  clients = [],
  activeClient,
  currentPath,
  onNavigate,
  hasPermission,
}: ClientSideNavProps) {
  const navClients = useMemo(() => {
    return clients
      .filter(isPlatformSideNavClient)
      .map((client) => {
        const items = filterByPermission(parseSideNav(client), hasPermission);

        return {
          client,
          items: sortItems(items),
        };
      })
      .filter(({ items }) => items.length > 0);
  }, [clients, hasPermission]);

  if (navClients.length === 0) {
    return null;
  }

  const handleNavigate = (path: string) => {
    if (currentPath === path) {
      return;
    }

    onNavigate(path);
  };

  return (
    <SideNav
      id="user-sidenav"
      isFixedNav
      expanded
      aria-label="Application navigation"
      className="moh-client-sidenav"
    >
      <SideNavItems>
        <div className="moh-client-sidenav__header">
          <span className="moh-client-sidenav__kicker">Applications</span>
          <span className="moh-client-sidenav__title">Workspace</span>
        </div>

        {navClients.map(({ client, items }) => {
          const clientActive =
            client.clientId === activeClient?.clientId ||
            items.some((item) => hasActiveChild(item, currentPath));

          return (
            <SideNavMenu
              key={client.clientId}
              title={getClientLabel(client)}
              defaultExpanded={clientActive}
              isActive={clientActive}
            >
              {items.map((item) => (
                <RenderSideNavItem
                  key={item.id}
                  item={item}
                  currentPath={currentPath}
                  onNavigate={handleNavigate}
                />
              ))}
            </SideNavMenu>
          );
        })}
      </SideNavItems>
    </SideNav>
  );
}

export function hasVisibleClientSideNav(
  clients: Client[],
  hasPermission?: (permission: string) => boolean,
): boolean {
  return clients.some((client) => {
    if (!isPlatformSideNavClient(client)) {
      return false;
    }
    return filterByPermission(parseSideNav(client), hasPermission).length > 0;
  });
}
