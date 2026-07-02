import { useState } from "react";
import { Button, Checkbox, InlineNotification, Select, SelectItem, TextInput, Tile } from "@carbon/react";
import { Accessibility, Save, SettingsAdjust } from "@carbon/react/icons";

import {
  readPortalPreferences,
  savePortalPreferences,
  type PortalPreferences,
} from "../preferences";

export default function PreferenceSettingsPage() {
  const [preferences, setPreferences] = useState<PortalPreferences>(() => readPortalPreferences());
  const [saved, setSaved] = useState(false);

  const updatePreference = <Key extends keyof PortalPreferences>(
    key: Key,
    value: PortalPreferences[Key],
  ) => {
    setSaved(false);
    setPreferences((current) => ({
      ...current,
      [key]: value,
    }));
  };

  const handleSave = () => {
    savePortalPreferences(preferences);
    setSaved(true);
  };

  return (
    <div className="settings-page">
      <div className="settings-page__header">
        <h2>Portal preferences</h2>
        <p className="settings-page__description">
          Tune how the portal displays tables, navigation, time, and accessibility defaults on this
          device.
        </p>
      </div>

      <Tile className="settings-card">
        <div className="settings-card__heading">
          <SettingsAdjust size={20} />
          <h3>Interface defaults</h3>
        </div>

        <div className="settings-toggle-list">
          <Checkbox
            id="settings-compact-tables"
            labelText="Use compact table density where supported"
            checked={preferences.compactTables}
            onChange={(_, { checked }) => updatePreference("compactTables", checked)}
          />

          <Checkbox
            id="settings-side-nav-collapsed"
            labelText="Start with side navigation collapsed"
            checked={preferences.sideNavCollapsed}
            onChange={(_, { checked }) => updatePreference("sideNavCollapsed", checked)}
          />

          <div className="settings-form-grid">
            <TextInput
              id="settings-items-per-page"
              labelText="Default items per page"
              type="number"
              min={5}
              max={100}
              value={String(preferences.itemsPerPage)}
              onChange={(event) =>
                updatePreference("itemsPerPage", Number(event.target.value) || 20)
              }
            />

            <TextInput
              id="settings-timezone"
              labelText="Timezone"
              value={preferences.timezone}
              onChange={(event) => updatePreference("timezone", event.target.value)}
            />
          </div>

          <Select
            id="settings-date-format"
            labelText="Date format"
            value={preferences.dateFormat}
            onChange={(event) =>
              updatePreference("dateFormat", event.target.value as PortalPreferences["dateFormat"])
            }
          >
            <SelectItem value="system" text="System default" />
            <SelectItem value="short" text="Short" />
            <SelectItem value="medium" text="Medium" />
          </Select>
        </div>
      </Tile>

      <Tile className="settings-card">
        <div className="settings-card__heading">
          <Accessibility size={20} />
          <h3>Accessibility</h3>
        </div>

        <p className="settings-card__text">
          These preferences are stored locally and can be used by modules as they adopt the shared
          settings contract.
        </p>

        <div className="settings-toggle-list">
          <Checkbox
            id="settings-reduced-motion"
            labelText="Reduce motion"
            checked={preferences.reducedMotion}
            onChange={(_, { checked }) => updatePreference("reducedMotion", checked)}
          />

          <Checkbox
            id="settings-larger-text"
            labelText="Prefer larger text"
            checked={preferences.largerText}
            onChange={(_, { checked }) => updatePreference("largerText", checked)}
          />

          <Checkbox
            id="settings-high-contrast"
            labelText="Prefer high contrast"
            checked={preferences.highContrast}
            onChange={(_, { checked }) => updatePreference("highContrast", checked)}
          />
        </div>
      </Tile>

      {saved ? (
        <InlineNotification
          kind="success"
          lowContrast
          hideCloseButton
          title="Preferences saved"
          subtitle="Your portal preferences were saved on this device."
        />
      ) : null}

      <div className="settings-card__actions">
        <Button kind="primary" renderIcon={Save} onClick={handleSave}>
          Save preferences
        </Button>
      </div>
    </div>
  );
}
