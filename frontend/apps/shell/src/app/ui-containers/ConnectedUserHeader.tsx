import { useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";

import { selectUser } from "@moh-sso/auth";
import { API } from "@moh-sso/config";
import { UserHeader } from "@moh-sso/ui";

import { ConnectedAppMenu } from "./ConnectedAppMenu";
import { HealthContextSelector } from "./HealthContextSelector";

export function ConnectedUserHeader() {
  const navigate = useNavigate();
  const user = useSelector(selectUser);

  const handleLogout = () => {
    window.location.replace(API.auth.logout());
  };

  return (
    <UserHeader
      username={user?.username}
      contextSelector={<HealthContextSelector />}
      appMenu={<ConnectedAppMenu />}
      onNavigateHome={() => navigate("/apps")}
      onLogout={handleLogout}
    />
  );
}
