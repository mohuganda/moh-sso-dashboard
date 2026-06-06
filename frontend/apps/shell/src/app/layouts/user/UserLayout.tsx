import { Content } from "@carbon/react";
import { Outlet } from "react-router-dom";

import { ClientSideNav, HeaderPanelProvider, PublicFooter, ToastProvider, UserHeader } from "@moh-sso/ui";

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
