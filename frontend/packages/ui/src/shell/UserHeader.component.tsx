import { UserAvatarFilled, Logout } from "@carbon/react/icons";
import {
  Header,
  HeaderName,
  HeaderGlobalBar,
  HeaderGlobalAction,
  SkipToContent,
} from "@carbon/react";
import type React from "react";
import type { ReactNode } from "react";

import AppMenuAction from "../navigation/appmenu/AppMenu.component";
import "./public-header.css";

import imagePath from "../assets/logo.png";

type UserHeaderProps = {
  username?: string | null;
  appMenu?: ReactNode;
  onNavigateHome?: () => void;
  onLogout?: () => void;
};

const UserHeader: React.FC<UserHeaderProps> = ({ username, appMenu, onNavigateHome, onLogout }) => {
  return (
    <Header aria-label="MOH Integrated Health Portal">
      <SkipToContent />

      <img src={imagePath} className={`moh-image-style`} />
      <HeaderName prefix="MOH" onClick={onNavigateHome} className="moh-header-name">
        Integrated Health Portal
      </HeaderName>

      <HeaderGlobalBar>
        {appMenu ?? <AppMenuAction />}
        <HeaderGlobalAction
          aria-label={`Signed in as ${username ?? "user"}`}
          tooltipAlignment="end"
        >
          <UserAvatarFilled size={20} />
        </HeaderGlobalAction>
        <HeaderGlobalAction
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
