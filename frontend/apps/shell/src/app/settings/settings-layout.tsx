import { NavLink, Outlet } from "react-router-dom";
import { Tile } from "@carbon/react";
import {
  Accessibility,
  Application,
  Information,
  Notification,
  Security,
  SettingsAdjust,
  Time,
  UserAvatar,
} from "@carbon/react/icons";

import "./settings.scss";

const settingsSections = [
  {
    id: "notifications",
    label: "Notifications",
    description: "Email, SMS, and quiet-hours preferences.",
    path: "/apps/settings/notifications",
    icon: Notification,
  },
  {
    id: "preferences",
    label: "Preferences",
    description: "Interface, accessibility, and table defaults.",
    path: "/apps/settings/preferences",
    icon: SettingsAdjust,
  },

  {
    id: "apps",
    label: "Applications",
    description: "Favorites and default landing app.",
    path: "/apps/settings/apps",
    icon: Application,
  },
  {
    id: "about",
    label: "About",
    description: "Version, support, and privacy information.",
    path: "/apps/settings/about",
    icon: Information,
  },
];

export function SettingsLayout() {
  return (
    <div className="settings-module">
      <div className="settings-module__header">
        <p className="settings-module__eyebrow">Account settings</p>
        <h1>Settings</h1>
        <p>
          Manage your profile, security, notifications, application preferences, and portal
          experience.
        </p>
      </div>

      <div className="settings-module__grid">
        <Tile className="settings-module__nav">
          <nav aria-label="Settings sections">
            {settingsSections.map((section) => {
              const Icon = section.icon;

              return (
                <NavLink
                  key={section.id}
                  to={section.path}
                  className={({ isActive }) =>
                    `settings-module__nav-link${isActive ? " settings-module__nav-link--active" : ""}`
                  }
                >
                  <Icon size={18} />
                  <span>
                    <strong>{section.label}</strong>
                    <small>{section.description}</small>
                  </span>
                </NavLink>
              );
            })}
          </nav>
        </Tile>

        <div className="settings-module__content">
          <Outlet />
        </div>
      </div>
    </div>
  );
}
