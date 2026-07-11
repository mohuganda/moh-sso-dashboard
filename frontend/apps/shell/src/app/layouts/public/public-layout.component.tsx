import { Content } from "@carbon/react";
import { Outlet } from "react-router-dom";

import { HeaderPanelProvider, PublicFooter } from "@moh-sso/ui";

import { useVersionInfo } from "@/app/version/useVersionInfo";
import { ConnectedPublicHeader } from "@/app/ui-containers";
import "./public-layout.scss";

export default function PublicLayout() {
  const versionInfo = useVersionInfo();

  return (
    <HeaderPanelProvider>
      <div className="public-layout">
        <ConnectedPublicHeader />

        <main className="public-layout__main">
          <Content id="main-content" className="public-layout__content">
            <Outlet />
          </Content>
        </main>

        <PublicFooter
          frontendVersion={versionInfo.frontend.version}
          backendVersion={versionInfo.backend?.version}
          backendUnavailable={!versionInfo.isLoading && !versionInfo.backend}
        />
      </div>
    </HeaderPanelProvider>
  );
}
