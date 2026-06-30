import { useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";
import type { ReactNode } from "react";

import { selectUser } from "@moh-sso/auth";
import { API } from "@moh-sso/config";
import { UserHeader } from "@moh-sso/ui";

import { ConnectedAppMenu } from "./ConnectedAppMenu";

type ConnectedUserHeaderProps = {
  navigationToggle?: ReactNode;
};

export function ConnectedUserHeader({ navigationToggle }: ConnectedUserHeaderProps) {
  const navigate = useNavigate();
  const user = useSelector(selectUser);

  const handleLogout = () => {
    window.location.replace(API.auth.logout());
  };

  return (
    <UserHeader
      username={user?.username}
      appMenu={<ConnectedAppMenu />}
      navigationToggle={navigationToggle}
      onNavigateHome={() => navigate("/apps")}
      onLogout={handleLogout}
    />
  );
}
