import { Content } from "@carbon/react";
import { Outlet } from "react-router-dom";
import { useMemo } from "react";

import { useAuthorization } from "@moh-sso/auth";
import { HeaderPanelProvider, PublicFooter, ToastProvider, hasVisibleClientSideNav } from "@moh-sso/ui";
import { mapAccessibleSystemToClient } from "@/app/access/accessClients";

import { RouteBreadcrumbBar } from "@/app/navigation/RouteBreadcrumbBar";
import { ConnectedClientSideNav, ConnectedUserHeader } from "@/app/ui-containers";

import "./user-layout.scss";

export default function UserLayout() {
  const { accessibleSystems, can } = useAuthorization();
  const clients = useMemo(
    () => accessibleSystems.map(mapAccessibleSystemToClient),
    [accessibleSystems],
  );
  const hasSideNav = hasVisibleClientSideNav(clients, (permission) => can(permission as never));

  return (
    <ToastProvider>
      <HeaderPanelProvider>
        <div className={`user-layout${hasSideNav ? "" : " user-layout--without-sidenav"}`}>
          <ConnectedUserHeader />

          <div className="user-layout__body">
            {hasSideNav && <ConnectedClientSideNav />}

            <div className="user-layout__content-shell">
              <RouteBreadcrumbBar />

              <Content id="main-content" className="user-layout__content">
                <Outlet />
              </Content>
            </div>
          </div>

          <PublicFooter />
        </div>
      </HeaderPanelProvider>
    </ToastProvider>
  );
}
