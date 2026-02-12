import { SideNav, SideNavItems, SideNavLink, SideNavMenu } from "@carbon/react";
import { useSelector } from "react-redux";
import { useLocation, useNavigate } from "react-router-dom";

import { selectClients, selectActiveClient } from "../../store/clients/clients.selectors";
import type { Client } from "../../store/types/client.types";

type SideNavItem = {
  id: string;
  label: string;
  path: string;
  permission?: string;
  children?: SideNavItem[];
};

/* --------------------------------
 * Helpers
 * -------------------------------- */

function parseSideNav(client: Client): SideNavItem[] {
  try {
    return JSON.parse(client.attributes?.["ui.sidenav"] ?? "[]");
  } catch {
    return [];
  }
}

function RenderSideNavItem({
  item,
  location,
  navigate,
}: {
  item: SideNavItem;
  location: ReturnType<typeof useLocation>;
  navigate: ReturnType<typeof useNavigate>;
}) {
  // Group / submenu
  if (item.children && item.children.length > 0) {
    return (
      <SideNavMenu key={item.id} title={item.label}>
        {item.children.map((child) => (
          <RenderSideNavItem key={child.id} item={child} location={location} navigate={navigate} />
        ))}
      </SideNavMenu>
    );
  }

  // Leaf link
  if (!item.path) return null;

  return (
    <SideNavLink
      key={item.id}
      isActive={location.pathname.startsWith(item.path)}
      onClick={() => navigate(item.path)}
    >
      {item.label}
    </SideNavLink>
  );
}

/* --------------------------------
 * Component
 * -------------------------------- */

export function ClientSideNav() {
  const clients = useSelector(selectClients);
  const activeClient = useSelector(selectActiveClient);

  const navigate = useNavigate();
  const location = useLocation();

  return (
    <SideNav isFixedNav expanded aria-label="Application navigation">
      <SideNavItems>
        {clients.map((client) => {
          const items = parseSideNav(client);
          if (items.length === 0) return null;

          return (
            <SideNavMenu
              key={client.clientId}
              title={client.name}
              isActive={client.clientId === activeClient?.clientId}
            >
              {items.map((item) => (
                <RenderSideNavItem
                  key={item.id}
                  item={item}
                  location={location}
                  navigate={navigate}
                />
              ))}
            </SideNavMenu>
          );
        })}
      </SideNavItems>
    </SideNav>
  );
}
