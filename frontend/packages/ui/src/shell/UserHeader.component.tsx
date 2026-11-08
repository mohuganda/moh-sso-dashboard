import { UserAvatarFilled, Logout } from "@carbon/react/icons";
import {
  Header,
  HeaderName,
  HeaderMenuButton,
  HeaderGlobalBar,
  HeaderGlobalAction,
  SkipToContent,
} from "@carbon/react";
import type React from "react";
import type { ReactNode } from "react";

import AppMenuAction from "../navigation/appmenu/AppMenu.component";
import "./public-header.scss";

import imagePath from "../assets/logo.png";

export type UserHeaderProps = {
  appName?: string;
  homeHref?: string;
  username?: string | null;
  appMenu?: ReactNode;
  contextSelector?: ReactNode;
  onNavigateHome?: () => void;
  onLogout?: () => void;
  onProfile?: () => void;
  navigationExpanded?: boolean;
  onToggleNavigation?: () => void;
  onGlobalAction?: () => void;
};

const UserHeader: React.FC<UserHeaderProps> = ({
  appName = "Integrated Health Portal",
  homeHref = "/portal/apps",
  username,
  appMenu,
  contextSelector,
  onNavigateHome,
  onLogout,
  onProfile,
  navigationExpanded = false,
  onToggleNavigation,
  onGlobalAction,
}) => {
  return (
    <Header aria-label={`MOH ${appName}`} className="moh-shell-header">
      <SkipToContent />
      {onToggleNavigation && (
        <HeaderMenuButton
          isCollapsible
          aria-label={navigationExpanded ? "Close navigation" : "Open navigation"}
          aria-controls="user-sidenav"
          aria-expanded={navigationExpanded}
          isActive={navigationExpanded}
          onClick={onToggleNavigation}
        />
      )}

      <img src={imagePath} className="moh-image-style" alt="Uganda coat of arms" />
      <HeaderName
        prefix="MOH"
        href={homeHref}
        onClick={(event) => {
          if (
            onNavigateHome &&
            !event.metaKey &&
            !event.ctrlKey &&
            !event.shiftKey &&
            !event.altKey
          ) {
            event.preventDefault();
            onNavigateHome();
          }
        }}
        className="moh-header-name"
        title={appName}
      >
        {appName}
      </HeaderName>

      <HeaderGlobalBar onClickCapture={onGlobalAction}>
        {contextSelector}
        {appMenu ?? <AppMenuAction />}
        <HeaderGlobalAction
          aria-label={`Signed in as ${username ?? "user"}`}
          tooltipAlignment="end"
          onClick={onProfile}
        >
          <UserAvatarFilled size={20} />
        </HeaderGlobalAction>
        <HeaderGlobalAction
          className={onProfile ? "moh-header-logout" : undefined}
          aria-label="Logout"
          tooltipAlignment="end"
          onClick={onLogout}
        >
          <Logout size={20} />
        </HeaderGlobalAction>
      </HeaderGlobalBar>
    </Header>
  );
};

export default UserHeader;
