import { Button, Tile } from "@carbon/react";
import { ArrowRight, Notification, Security, SettingsAdjust, UserAvatar } from "@carbon/react/icons";
import { useNavigate } from "react-router-dom";

const cards = [
  {
    title: "Profile",
    description: "Review your account details and safe profile fields.",
    path: "/apps/settings/profile",
    icon: UserAvatar,
  },
  {
    title: "Security",
    description: "Open password, MFA, and account security actions.",
    path: "/apps/settings/security",
    icon: Security,
  },
  {
    title: "Notifications",
    description: "Manage email, SMS, phone number, and quiet hours.",
    path: "/apps/settings/notifications",
    icon: Notification,
  },
  {
    title: "Preferences",
    description: "Tune table density, accessibility, and display defaults.",
    path: "/apps/settings/preferences",
    icon: SettingsAdjust,
  },
];

export default function SettingsHomePage() {
  const navigate = useNavigate();

  return (
    <div className="settings-page">
      <div className="settings-page__header">
        <h2>Settings overview</h2>
        <p className="settings-page__description">
          Choose a section to manage how your account and portal workspace behave.
        </p>
      </div>

      <div className="settings-card-grid">
        {cards.map((card) => {
          const Icon = card.icon;

          return (
            <Tile key={card.path} className="settings-card">
              <div className="settings-card__heading">
                <Icon size={20} />
                <h3>{card.title}</h3>
              </div>

              <p className="settings-card__text">{card.description}</p>

              <div className="settings-card__actions">
                <Button size="sm" kind="ghost" renderIcon={ArrowRight} onClick={() => navigate(card.path)}>
                  Open
                </Button>
              </div>
            </Tile>
          );
        })}
      </div>
    </div>
  );
}
