import { Content } from "@carbon/react";
import { Outlet } from "react-router-dom";

import { PublicFooter } from "../../components/footer/PublicFooter";
import UserHeader from "../../components/header/UserHeader.component";
import { ClientSideNav } from "../../components/sidenav/ClientSideNav";
import { ToastProvider } from "../../components/notifications/toast/ToastProvider";
import { HeaderPanelProvider } from "../../components/header-panel/header-panel.context";

export default function PublicLayout() {
  return (
    <ToastProvider>
      <HeaderPanelProvider>
        <div
          style={{
            minHeight: "100vh",
            display: "flex",
            flexDirection: "column",
          }}
        >
          <UserHeader />

          <div
            style={{
              display: "flex",
              flex: 1,
            }}
          >
            <ClientSideNav />

            <Content
              id="main-content"
              style={{
                marginTop: "3rem",
                flex: 1,
                minWidth: 0,
                overflowX: "hidden",
                background: "#f9fafb",
              }}
            >
              <Outlet />
            </Content>
          </div>
          <PublicFooter />
        </div>
      </HeaderPanelProvider>
    </ToastProvider>
  );
}
