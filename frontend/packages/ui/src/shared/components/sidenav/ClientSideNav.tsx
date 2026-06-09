import { useMemo } from "react";
import { SideNav, SideNavItems, SideNavLink, SideNavMenu } from "@carbon/react";
import { useSelector } from "react-redux";
import { useLocation, useNavigate } from "react-router-dom";

import { selectActiveClient, selectClients } from "@moh-sso/state";
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

    return parsed.filter(isValidSideNavItem);
  } catch {
    return [];
  }
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
  hasPermission?: (permission: string) => boolean;
};

export function ClientSideNav({ hasPermission }: ClientSideNavProps) {
  const clients = useSelector(selectClients);
  const activeClient = useSelector(selectActiveClient);

  const navigate = useNavigate();
  const location = useLocation();

  const navClients = useMemo(() => {
    return clients
      .map((client) => {
        const items = filterByPermission(parseSideNav(client), hasPermission);

        return {
          client,
          items: sortItems(items),
        };
      })
      .filter(({ items }) => items.length > 0);
  }, [clients, hasPermission]);

  const handleNavigate = (path: string) => {
    if (location.pathname === path) {
      return;
    }

    navigate(path);
  };

  return (
    <SideNav isFixedNav expanded aria-label="Application navigation">
      <SideNavItems>
        {navClients.map(({ client, items }) => {
          const clientActive =
            client.clientId === activeClient?.clientId ||
            items.some((item) => hasActiveChild(item, location.pathname));

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
                  currentPath={location.pathname}
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
