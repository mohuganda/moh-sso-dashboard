import React from "react";
import { useDispatch } from "react-redux";

import { setActiveClient } from "../../store/clients/clients.slice";

import "./AppMenu.css";
import { ClickableTile } from "@carbon/react";

interface AppTileProps {
  icon: React.ElementType;
  name: string;
  href: string;
  clientId?: string;
}

const AppTile: React.FC<AppTileProps> = ({ icon: Icon, name, href, clientId }) => {
  const dispatch = useDispatch();

  const handleClick = () => {
    // 🔑 Activate client if provided
    if (clientId) {
      dispatch(setActiveClient(clientId));
    }

    // 🌍 Open in new tab (safe defaults)
    window.open(href, "_blank", "noopener,noreferrer");
  };

  return (
    <ClickableTile
      id={`clickable-tile-${name}`}
      onClick={handleClick}
      className={`clickable-tile-moh`}
    >
      <div>
        <Icon size={25} className="app-menu-item__icon" />
      </div>
      <div>{name}</div>
    </ClickableTile>
  );
};

export default AppTile;
