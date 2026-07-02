import { useEffect, useState } from "react";
import { Button, Checkbox, InlineNotification, TextInput, Tile } from "@carbon/react";
import { Notification, Save } from "@carbon/react/icons";

import {
  useGetNotificationPreferencesQuery,
  useUpdateNotificationPreferencesMutation,
} from "@/app/api/notifications.api";

export default function NotificationSettingsPage() {
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
    if (!notificationPreferences) {
      return;
    }

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
    <div className="settings-page">
      <div className="settings-page__header">
        <h2>Notification preferences</h2>
        <p className="settings-page__description">
          Choose how the portal should contact you for operational, system, and announcement
          messages.
        </p>
      </div>

      <Tile className="settings-card">
        <div className="settings-card__heading">
          <Notification size={20} />
          <h3>Delivery channels</h3>
        </div>

        <p className="settings-card__text">
          Email and SMS preferences are used by announcements and notification delivery workflows.
        </p>

        <div className="settings-toggle-list">
          <Checkbox
            id="settings-notification-email-enabled"
            labelText="Email notifications"
            checked={emailEnabled}
            disabled={preferencesLoading || preferencesSaving}
            onChange={(_, { checked }) => setEmailEnabled(checked)}
          />

          <Checkbox
            id="settings-notification-sms-enabled"
            labelText="SMS notifications"
            checked={smsEnabled}
            disabled={preferencesLoading || preferencesSaving}
            onChange={(_, { checked }) => setSmsEnabled(checked)}
          />

          <TextInput
            id="settings-notification-phone-number"
            labelText="Mobile number"
            placeholder="+256..."
            value={phoneNumber}
            disabled={preferencesLoading || preferencesSaving}
            onChange={(event) => setPhoneNumber(event.target.value)}
          />

          <div className="settings-form-grid">
            <TextInput
              id="settings-notification-quiet-hours-start"
              labelText="Quiet hours start"
              placeholder="22:00"
              value={quietHoursStart}
              disabled={preferencesLoading || preferencesSaving}
              onChange={(event) => setQuietHoursStart(event.target.value)}
            />

            <TextInput
              id="settings-notification-quiet-hours-end"
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
              subtitle="SMS can be configured now. Verification can be enforced before production sends."
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
        </div>

        <div className="settings-card__actions">
          <Button
            kind="primary"
            renderIcon={Save}
            disabled={preferencesLoading || preferencesSaving}
            onClick={saveNotificationPreferences}
          >
            Save preferences
          </Button>
        </div>
      </Tile>
    </div>
  );
}
