import { Content } from "@carbon/react";
import { Outlet } from "react-router-dom";

import { HeaderPanelProvider, PublicFooter, ToastProvider } from "@moh-sso/ui";

import { ConnectedPublicHeader } from "@/app/ui-containers";
import "./public-layout.scss";

export default function PublicLayout() {
  return (
    <ToastProvider>
      <HeaderPanelProvider>
        <div className="public-layout">
          <ConnectedPublicHeader />

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
