import React from "react";
import { useDispatch } from "react-redux";
import { ClickableTile } from "@carbon/react";
import { Launch } from "@carbon/react/icons";
import { useNavigate } from "react-router-dom";

import { setActiveClient } from "@moh-sso/state";

import "./AppMenu.scss";

interface AppTileProps {
  icon: React.ElementType;
  name: string;
  href: string;
  clientId?: string;
  onSelect?: () => void;
}

function isExternalUrl(url: string): boolean {
  return /^https?:\/\//i.test(url);
}

function isValidHref(href?: string | null): href is string {
  return Boolean(href && href.trim().length > 0 && href.trim() !== "#");
}

function getTileId(name: string) {
  return `app-menu-tile-${name
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/(^-|-$)/g, "")}`;
}

const AppTile: React.FC<AppTileProps> = ({ icon: Icon, name, href, clientId, onSelect }) => {
  const dispatch = useDispatch();
  const navigate = useNavigate();

  const disabled = !isValidHref(href);
  const external = isValidHref(href) && isExternalUrl(href);

  const handleClick = () => {
    if (disabled) {
      return;
    }

    if (clientId) {
      dispatch(setActiveClient(clientId));
    }

    onSelect?.();

    if (external) {
      window.open(href, "_blank", "noopener,noreferrer");
      return;
    }

    navigate(href);
  };

  return (
    <ClickableTile
      id={getTileId(name)}
      className={["app-menu-tile", disabled ? "app-menu-tile--disabled" : ""]
        .filter(Boolean)
        .join(" ")}
      onClick={handleClick}
      disabled={disabled}
      aria-label={external ? `Open ${name} in a new tab` : `Open ${name}`}
    >
      <span className="app-menu-tile__icon-wrap" aria-hidden="true">
        <Icon size={24} className="app-menu-tile__icon" />
      </span>

      <span className="app-menu-tile__label">{name}</span>

      {external && (
        <span className="app-menu-tile__external" aria-hidden="true">
          <Launch size={14} />
        </span>
      )}
    </ClickableTile>
  );
};

export default AppTile;
