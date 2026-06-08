import { UserAvatarFilled, Logout } from "@carbon/react/icons";
import {
  Header,
  HeaderName,
  HeaderGlobalBar,
  HeaderGlobalAction,
  SkipToContent,
} from "@carbon/react";
import type React from "react";
import { useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";

import { API } from "@moh-sso/config";
import { selectUser } from "@moh-sso/auth";
import { BRANDING } from "@moh-sso/config";
import AppMenuAction from "../appmenu/AppMenu.component";
import "./public-header.css";

const UserHeader: React.FC = () => {
  const navigate = useNavigate();
  const user = useSelector(selectUser);

  const handleLogout = () => {
    window.location.replace(API.auth.logout());
  };
  return (
    <Header aria-label="MOH Integrated Health Portal">
      <SkipToContent />

      {/* Brand / Home */}
      <img src={BRANDING.logo} className={`moh-image-style`} />
      <HeaderName prefix="MOH" onClick={() => navigate("/apps")} className="moh-header-name">
        {BRANDING.title}
      </HeaderName>

      {/* Global Actions */}
      <HeaderGlobalBar>
        {/* 🌱 Environment */}
        {/*<EnvironmentBadge />*/}
        {/* App Launcher */}
        <AppMenuAction />
        {/* User indicator */}
        <HeaderGlobalAction
          aria-label={`Signed in as ${user?.username ?? "user"}`}
          tooltipAlignment="end"
        >
          <UserAvatarFilled size={20} />
        </HeaderGlobalAction>
        {/* Logout */}
        <HeaderGlobalAction
          aria-label="Logout"
          tooltipAlignment="end"
          onClick={() => {
            handleLogout();
          }}
        >
          <Logout size={20} />
        </HeaderGlobalAction>
      </HeaderGlobalBar>
    </Header>
  );
};

export default UserHeader;
