import { Content } from "@carbon/react";
import { Outlet } from "react-router-dom";
import { useEffect, useMemo, useState } from "react";

import { useAuthorization } from "@moh-sso/auth";
import { HeaderPanelProvider, PublicFooter, ToastProvider, hasVisibleClientSideNav } from "@moh-sso/ui";
import { buildAccessibleSideNavClients } from "@/app/access/accessClients";

import { RouteBreadcrumbBar } from "@/app/navigation/RouteBreadcrumbBar";
import { ConnectedClientSideNav, ConnectedUserHeader } from "@/app/ui-containers";
import { useVersionInfo } from "@/app/version/useVersionInfo";

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
  const versionInfo = useVersionInfo();
  const clients = useMemo(() => buildAccessibleSideNavClients({ accessibleSystems }), [accessibleSystems]);
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

  return (
    <ToastProvider>
      <HeaderPanelProvider>
        <div className={layoutClassName}>
          <ConnectedUserHeader />

          <div className="user-layout__body">
            {hasSideNav && (
              <ConnectedClientSideNav
                visible={isSideNavVisible}
                onToggleVisibility={() => setIsSideNavVisible((visible) => !visible)}
              />
            )}

            <div className="user-layout__content-shell">
              <RouteBreadcrumbBar />

              <Content id="main-content" className="user-layout__content">
                <Outlet />
              </Content>
            </div>
          </div>

          <PublicFooter
            frontendVersion={versionInfo.frontend.version}
            backendVersion={versionInfo.backend?.version}
            backendUnavailable={!versionInfo.isLoading && !versionInfo.backend}
          />
        </div>
      </HeaderPanelProvider>
    </ToastProvider>
  );
}
