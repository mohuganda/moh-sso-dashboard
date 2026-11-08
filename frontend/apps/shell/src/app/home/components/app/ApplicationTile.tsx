import { Application, ArrowRight, Launch } from "@carbon/react/icons";
import { Tile, Tag, Button } from "@carbon/react";
import { useNavigate } from "react-router-dom";
import "./application-tile.scss";

export type ApplicationTileProps = {
  clientId: string;
  name: string;
  description?: string | null;
  enabled: boolean;
  rootUrl?: string;
  launchMode?: "internal" | "new_tab" | "same_tab";
  onLaunch?: () => void;
};

export function ApplicationTile({
  name,
  description,
  enabled,
  rootUrl,
  launchMode = "internal",
  onLaunch,
}: ApplicationTileProps) {
  const navigate = useNavigate();
  const available = enabled && Boolean(rootUrl);
  const external = launchMode === "new_tab";
  const handleLaunch = () => {
    if (!available || !rootUrl) return;
    onLaunch?.();
    if (external) window.open(rootUrl, "_blank", "noopener,noreferrer");
    else if (launchMode === "same_tab") window.location.assign(rootUrl);
    else navigate(rootUrl);
  };
  return (
    <Tile
      role="article"
      aria-label={name}
      className={`app-tile ${!available ? "app-tile--disabled" : ""}`}
    >
      <div className="app-tile__header">
        <span className="app-tile__icon" aria-hidden="true">
          <Application size={20} />
        </span>
        <h5 className="app-tile__title">{name}</h5>
        {!available && (
          <Tag size="sm" type="gray">
            Unavailable
          </Tag>
        )}
      </div>
      <p className="app-tile__description">
        {description || "Access tools and resources for your work."}
      </p>
      <div className="app-tile__footer">
        <Button
          size="sm"
          kind="ghost"
          renderIcon={external ? Launch : ArrowRight}
          disabled={!available}
          onClick={handleLaunch}
          aria-label={external ? `Open ${name} in a new tab` : `Open ${name}`}
        >
          Open application
        </Button>
        {external && <span className="app-tile__hint">New tab</span>}
      </div>
    </Tile>
  );
}
