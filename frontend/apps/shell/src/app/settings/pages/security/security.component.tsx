import { useEffect, useState } from "react";
import { Button, Checkbox, InlineNotification, Tag, TextInput, Tile } from "@carbon/react";
import { Launch, Locked, Notification, Security, UserAdmin } from "@carbon/react/icons";

import {
  useGetNotificationPreferencesQuery,
  useUpdateNotificationPreferencesMutation,
} from "@/app/api/notifications.api";

import "./security.scss";

const DEFAULT_KEYCLOAK_ACCOUNT_URL = "http://localhost:8081/realms/moh-realm/account";

function getKeycloakAccountUrl() {
  return import.meta.env.VITE_KEYCLOAK_ACCOUNT_URL || DEFAULT_KEYCLOAK_ACCOUNT_URL;
}

export default function SecurityPage() {
  const keycloakAccountUrl = getKeycloakAccountUrl();
  const { data: notificationPreferences, isLoading: preferencesLoading } =
    useGetNotificationPreferencesQuery();
  const [updateNotificationPreferences, { isLoading: preferencesSaving, error: preferencesError }] =
    useUpdateNotificationPreferencesMutation();
  const [emailEnabled, setEmailEnabled] = useState(true);
  const [smsEnabled, setSmsEnabled] = useState(false);
  const [phoneNumber, setPhoneNumber] = useState("");
  const [quietHoursStart, setQuietHoursStart] = useState("");
  const [quietHoursEnd, setQuietHoursEnd] = useState("");
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    if (!notificationPreferences) return;

    setEmailEnabled(notificationPreferences.email_enabled);
    setSmsEnabled(notificationPreferences.sms_enabled);
    setPhoneNumber(notificationPreferences.phone_number ?? "");
    setQuietHoursStart(notificationPreferences.quiet_hours_start ?? "");
    setQuietHoursEnd(notificationPreferences.quiet_hours_end ?? "");
  }, [notificationPreferences]);

  const saveNotificationPreferences = async () => {
    setSaved(false);
    await updateNotificationPreferences({
      email_enabled: emailEnabled,
      sms_enabled: smsEnabled,
      phone_number: phoneNumber,
      quiet_hours_start: quietHoursStart,
      quiet_hours_end: quietHoursEnd,
    }).unwrap();
    setSaved(true);
  };

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
          <Notification size={20} />
          <h4>Notification Preferences</h4>
        </div>

        <p className="security-card__text">
          Choose how the portal should contact you for account, system, and operational alerts.
        </p>

        <div className="notification-preferences">
          <Checkbox
            id="notification-email-enabled"
            labelText="Email notifications"
            checked={emailEnabled}
            disabled={preferencesLoading || preferencesSaving}
            onChange={(_, { checked }) => setEmailEnabled(checked)}
          />

          <Checkbox
            id="notification-sms-enabled"
            labelText="SMS notifications"
            checked={smsEnabled}
            disabled={preferencesLoading || preferencesSaving}
            onChange={(_, { checked }) => setSmsEnabled(checked)}
          />

          <TextInput
            id="notification-phone-number"
            labelText="Mobile number"
            placeholder="+256..."
            value={phoneNumber}
            disabled={preferencesLoading || preferencesSaving}
            onChange={(event) => setPhoneNumber(event.target.value)}
          />

          <div className="notification-preferences__quiet-hours">
            <TextInput
              id="notification-quiet-hours-start"
              labelText="Quiet hours start"
              placeholder="22:00"
              value={quietHoursStart}
              disabled={preferencesLoading || preferencesSaving}
              onChange={(event) => setQuietHoursStart(event.target.value)}
            />

            <TextInput
              id="notification-quiet-hours-end"
              labelText="Quiet hours end"
              placeholder="06:00"
              value={quietHoursEnd}
              disabled={preferencesLoading || preferencesSaving}
              onChange={(event) => setQuietHoursEnd(event.target.value)}
            />
          </div>

          {notificationPreferences?.phone_verified ? (
            <InlineNotification
              kind="success"
              lowContrast
              hideCloseButton
              title="Phone verified"
              subtitle="SMS delivery is allowed for this number."
            />
          ) : smsEnabled ? (
            <InlineNotification
              kind="warning"
              lowContrast
              hideCloseButton
              title="Phone verification pending"
              subtitle="SMS can be configured now. Verification can be added before production sends."
            />
          ) : null}

          {preferencesError ? (
            <InlineNotification
              kind="error"
              lowContrast
              title="Unable to save preferences"
              subtitle="Check the phone number and quiet hours format, then try again."
            />
          ) : null}

          {saved ? (
            <InlineNotification
              kind="success"
              lowContrast
              hideCloseButton
              title="Preferences saved"
              subtitle="Your notification preferences were updated."
            />
          ) : null}

          <div className="security-card__actions">
            <Button
              kind="primary"
              disabled={preferencesLoading || preferencesSaving}
              onClick={saveNotificationPreferences}
            >
              Save preferences
            </Button>
          </div>
        </div>
      </Tile>

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
