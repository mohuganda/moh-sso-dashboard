import { Content } from "@carbon/react";
import { Outlet } from "react-router-dom";

import {
  ClientSideNav,
  HeaderPanelProvider,
  PublicFooter,
  ToastProvider,
  UserHeader,
} from "@moh-sso/ui";

import { RouteBreadcrumbBar } from "@/app/navigation/RouteBreadcrumbBar";

import "./user-layout.scss";

export default function UserLayout() {
  return (
    <ToastProvider>
      <HeaderPanelProvider>
        <div className="user-layout">
          <UserHeader />

          <div className="user-layout__body">
            <ClientSideNav />

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
