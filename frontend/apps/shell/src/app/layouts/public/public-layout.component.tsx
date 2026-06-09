import { Content } from "@carbon/react";
import { Outlet } from "react-router-dom";

import { HeaderPanelProvider, PublicFooter, PublicHeader, ToastProvider } from "@moh-sso/ui";

import "./public-layout.scss";

export default function PublicLayout() {
  return (
    <ToastProvider>
      <HeaderPanelProvider>
        <div className="public-layout">
          <PublicHeader />

          <main className="public-layout__main">
            <Content id="main-content" className="public-layout__content">
              <Outlet />
            </Content>
          </main>

          <PublicFooter />
        </div>
      </HeaderPanelProvider>
    </ToastProvider>
  );
}
