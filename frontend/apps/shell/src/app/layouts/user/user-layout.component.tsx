import { Content } from "@carbon/react";
import { Outlet } from "react-router-dom";

import { HeaderPanelProvider, PublicFooter, ToastProvider } from "@moh-sso/ui";

import { RouteBreadcrumbBar } from "@/app/navigation/RouteBreadcrumbBar";
import { ConnectedClientSideNav, ConnectedUserHeader } from "@/app/ui-containers";

import "./user-layout.scss";

export default function UserLayout() {
  return (
    <ToastProvider>
      <HeaderPanelProvider>
        <div className="user-layout">
          <ConnectedUserHeader />

          <div className="user-layout__body">
            <ConnectedClientSideNav />

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
