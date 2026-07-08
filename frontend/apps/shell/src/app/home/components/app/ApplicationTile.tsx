import { Launch } from "@carbon/react/icons";
import { Tile, Stack, Tag, Button } from "@carbon/react";
import { useNavigate } from "react-router-dom";
import "./application-tile.scss";

export type ApplicationTileProps = {
  clientId: string;
  name: string;
  description?: string | null;
  enabled: boolean;

  /** Application root URL */
  rootUrl?: string;
  launchMode?: "internal" | "new_tab" | "same_tab";

  /** Optional override click handler */
  onLaunch?: () => void;
};

export function ApplicationTile({
  clientId,
  name,
  description,
  enabled,
  rootUrl,
  launchMode = "internal",
  onLaunch,
}: ApplicationTileProps) {
  const navigate = useNavigate();
  const handleLaunch = () => {
    if (!enabled) return;

    if (onLaunch) {
      onLaunch();
      return;
    }

    if (rootUrl) {
      if (launchMode === "new_tab") {
        window.open(rootUrl, "_blank", "noopener,noreferrer");
      } else if (launchMode === "same_tab") {
        window.location.assign(rootUrl);
      } else {
        navigate(rootUrl);
      }
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
