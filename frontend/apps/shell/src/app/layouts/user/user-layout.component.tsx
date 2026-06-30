import { Content, HeaderGlobalAction } from "@carbon/react";
import { Menu } from "@carbon/react/icons";
import { Outlet } from "react-router-dom";
import { useEffect, useMemo, useState } from "react";

import { useAuthorization } from "@moh-sso/auth";
import { HeaderPanelProvider, PublicFooter, ToastProvider, hasVisibleClientSideNav } from "@moh-sso/ui";
import { mapAccessibleSystemToClient } from "@/app/access/accessClients";

import { RouteBreadcrumbBar } from "@/app/navigation/RouteBreadcrumbBar";
import { ConnectedClientSideNav, ConnectedUserHeader } from "@/app/ui-containers";

import "./user-layout.scss";

const USER_SIDENAV_STORAGE_KEY = "moh.userLayout.sideNavVisible";

function readStoredSideNavVisible(): boolean {
  if (typeof window === "undefined") {
    return true;
  }

  return window.localStorage.getItem(USER_SIDENAV_STORAGE_KEY) !== "false";
}

export default function UserLayout() {
  const { accessibleSystems, can } = useAuthorization();
  const [isSideNavVisible, setIsSideNavVisible] = useState(readStoredSideNavVisible);
  const clients = useMemo(
    () => accessibleSystems.map(mapAccessibleSystemToClient),
    [accessibleSystems],
  );
  const hasSideNav = hasVisibleClientSideNav(clients, (permission) => can(permission as never));
  const layoutClassName = [
    "user-layout",
    !hasSideNav ? "user-layout--without-sidenav" : "",
    hasSideNav && !isSideNavVisible ? "user-layout--sidenav-hidden" : "",
  ]
    .filter(Boolean)
    .join(" ");

  useEffect(() => {
    window.localStorage.setItem(USER_SIDENAV_STORAGE_KEY, String(isSideNavVisible));
  }, [isSideNavVisible]);

  const navigationToggle = hasSideNav ? (
    <HeaderGlobalAction
      aria-label={isSideNavVisible ? "Hide navigation" : "Show navigation"}
      aria-controls="user-sidenav"
      aria-expanded={isSideNavVisible}
      className="user-layout__sidenav-toggle"
      tooltipAlignment="start"
      onClick={() => setIsSideNavVisible((visible) => !visible)}
    >
      <Menu size={22} />
    </HeaderGlobalAction>
  ) : undefined;

  return (
    <ToastProvider>
      <HeaderPanelProvider>
        <div className={layoutClassName}>
          <ConnectedUserHeader navigationToggle={navigationToggle} />

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
