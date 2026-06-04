import { Content } from "@carbon/react";
import { Outlet } from "react-router-dom";

import { PublicFooter } from "@/shared/components/footer/PublicFooter";
import UserHeader from "@/shared/components/header/UserHeader.component";
import { ClientSideNav } from "@/shared/components/sidenav/ClientSideNav";
import { ToastProvider } from "@/shared/components/notifications/toast/ToastProvider";
import { HeaderPanelProvider } from "@/shared/components/header-panel/header-panel.context";

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
