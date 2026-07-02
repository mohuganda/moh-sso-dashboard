import { Button, InlineNotification, Tag, Tile } from "@carbon/react";
import { Launch, Locked, Security, UserAdmin } from "@carbon/react/icons";

import "./security.scss";

const DEFAULT_KEYCLOAK_ACCOUNT_URL = "http://localhost:8081/realms/moh-realm/account";

function getKeycloakAccountUrl() {
  return import.meta.env.VITE_KEYCLOAK_ACCOUNT_URL || DEFAULT_KEYCLOAK_ACCOUNT_URL;
}

export default function SecurityPage() {
  const keycloakAccountUrl = getKeycloakAccountUrl();

  return (
    <div className="security-page">
      <Tile className="security-card security-hero-card">
        <div className="security-hero">
          <div className="security-hero__icon" aria-hidden="true">
            <Security size={28} />
          </div>

          <div className="security-hero__content">
            <div className="security-hero__title-row">
              <h3 className="security-hero__title">Security Settings</h3>

              <Tag size="sm" type="blue">
                Identity managed
              </Tag>
            </div>

            <p className="security-hero__description">
              Sensitive account actions such as password changes, multi-factor authentication, and
              account recovery are managed through the MOH centralized identity provider.
            </p>
          </div>
        </div>
      </Tile>

      <div className="security-grid">
        <Tile className="security-card">
          <div className="security-card__heading">
            <Locked size={20} />
            <h4>Password & MFA</h4>
          </div>

          <p className="security-card__text">
            Update your password, manage multi-factor authentication, and review account-level
            security options in the secure account portal.
          </p>

          <div className="security-card__actions">
            <Button
              kind="primary"
              href={keycloakAccountUrl}
              target="_blank"
              rel="noopener noreferrer"
              renderIcon={Launch}
            >
              Manage Password & MFA
            </Button>
          </div>
        </Tile>

        <Tile className="security-card">
          <div className="security-card__heading">
            <UserAdmin size={20} />
            <h4>Account Management Portal</h4>
          </div>

          <p className="security-card__text">
            You will be redirected to the MOH Identity Management console. Some changes may require
            you to sign in again.
          </p>

          <InlineNotification
            kind="info"
            lowContrast
            hideCloseButton
            title="Secure redirect"
            subtitle="The account portal opens in a new tab to keep your current session active."
          />
        </Tile>
      </div>

      <Tile className="security-card security-note-card">
        <div className="security-card__heading">
          <Security size={20} />
          <h4>Security Recommendations</h4>
        </div>

        <ul className="security-list">
          <li>Use a strong password that is not shared with other systems.</li>
          <li>Enable multi-factor authentication where available.</li>
          <li>Review active sessions and end sessions you do not recognize.</li>
          <li>Report suspicious account activity to system support.</li>
        </ul>
      </Tile>
    </div>
  );
}
