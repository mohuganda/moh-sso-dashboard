import { Launch } from "@carbon/react/icons";
import { Tile, Stack, Tag, Button } from "@carbon/react";
import "./application-tile.css";

export type ApplicationTileProps = {
  clientId: string;
  name: string;
  description?: string | null;
  enabled: boolean;

  /** Application root URL */
  rootUrl?: string;

  /** Optional override click handler */
  onLaunch?: () => void;
};

export function ApplicationTile({
  clientId,
  name,
  description,
  enabled,
  rootUrl,
  onLaunch,
}: ApplicationTileProps) {
  const handleLaunch = () => {
    if (!enabled) return;

    if (onLaunch) {
      onLaunch();
      return;
    }

    if (rootUrl) {
      window.location.href = rootUrl;
    }
  };

  return (
    <Tile key={clientId} className={`app-tile ${!enabled ? "app-tile--disabled" : ""}`}>
      <Stack gap={4}>
        {/* Header */}
        <div className="app-header">
          <strong>{name}</strong>

          {!enabled && (
            <Tag size="sm" type="gray">
              Disabled
            </Tag>
          )}
        </div>

        {/* Description */}
        <p className="app-description">{description || "No description provided"}</p>

        {/* Action */}
        <Button
          size="sm"
          kind={enabled ? "primary" : "secondary"}
          renderIcon={Launch}
          disabled={!enabled}
          onClick={handleLaunch}
        >
          {enabled ? "Launch" : "Unavailable"}
        </Button>
      </Stack>
    </Tile>
  );
}
