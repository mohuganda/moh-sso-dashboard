import { Button } from "@carbon/react";
import { ArrowRight, Logout, Settings, UserAvatar } from "@carbon/react/icons";
import { PERMISSIONS, selectUser, useAuthorization } from "@moh-sso/auth";
import { API } from "@moh-sso/config";
import { useHeaderPanel } from "@moh-sso/ui";
import { useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";

import "./shell-account.scss";

export function ConnectedAccountPanel() {
  const user = useSelector(selectUser);
  const { can, hasSystem } = useAuthorization();
  const { closePanel } = useHeaderPanel();
  const navigate = useNavigate();
  const name =
    user?.fullName ||
    [user?.firstName, user?.lastName].filter(Boolean).join(" ") ||
    user?.username ||
    "Your account";
  const canOpenSettings = can(PERMISSIONS.systemsLaunch) && hasSystem("settings");
  const navigateTo = (path: string) => {
    closePanel();
    navigate(path);
  };

  return (
    <section className="shell-account" aria-label="Your account">
      <div className="shell-account__identity">
        <span className="shell-account__avatar" aria-hidden="true">
          <UserAvatar size={32} />
        </span>
        <div>
          <h2>{name}</h2>
          <p>{user?.email || user?.username}</p>
        </div>
      </div>
      <p className="shell-account__hint">Manage your account or switch workspaces.</p>
      <div className="shell-account__links">
        <Button kind="tertiary" renderIcon={ArrowRight} onClick={() => navigateTo("/apps")}>
          My workspace
        </Button>
        {canOpenSettings && (
          <Button
            kind="tertiary"
            renderIcon={Settings}
            onClick={() => navigateTo("/apps/settings/profile")}
          >
            Profile and settings
          </Button>
        )}
        {user?.isAdmin && (
          <Button kind="tertiary" renderIcon={ArrowRight} onClick={() => navigateTo("/admin")}>
            Administration
          </Button>
        )}
      </div>
      <Button
        kind="danger--ghost"
        renderIcon={Logout}
        onClick={() => window.location.replace(API.auth.logout())}
      >
        Sign out
      </Button>
    </section>
  );
}
