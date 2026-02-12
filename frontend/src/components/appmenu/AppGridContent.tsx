import { Document, Calendar, Email, Menu, User, App, IbmJrs } from "@carbon/react/icons";
import { InlineLoading, Tile } from "@carbon/react";
import React, { useEffect } from "react";
import { useDispatch } from "react-redux";

import { useListClientsQuery } from "../../store/api/clients.api";
import { setClients } from "../../store/clients/clients.slice";
import type { Client } from "../../store/types/client.types";

import AppTile from "./AppMenuItem.component";
import "./AppMenu.css";

const ICON_MAP: Record<string, React.ElementType> = {
  email: Email,
  calendar: Calendar,
  document: Document,
  settings: Menu,
  user: User,
  applications: App,
  reporting: IbmJrs,
};

const AppGridContent: React.FC = () => {
  const dispatch = useDispatch();

  const { data: clients = [], isLoading, isError, refetch } = useListClientsQuery();

  /* -----------------------------
   * 🔑 Sync clients into Redux
   * ----------------------------- */
  useEffect(() => {
    if (clients.length > 0) {
      dispatch(setClients(clients));
    }
  }, [clients, dispatch]);

  /* -----------------------------
   * Loading state
   * ----------------------------- */
  if (isLoading) {
    return (
      <div className="app-grid-state">
        <InlineLoading description="Loading applications…" />
      </div>
    );
  }

  /* -----------------------------
   * Error state
   * ----------------------------- */
  if (isError) {
    return (
      <div className="app-grid-state">
        <Tile>
          <p style={{ marginBottom: "0.5rem" }}>Failed to load applications.</p>
          <button type="button" className="app-grid-retry" onClick={() => refetch()}>
            Retry
          </button>
        </Tile>
      </div>
    );
  }

  /* -----------------------------
   * Empty state
   * ----------------------------- */
  if (clients.length === 0) {
    return (
      <div className="app-grid-state">
        <p style={{ opacity: 0.7 }}>No applications available.</p>
      </div>
    );
  }

  return (
    <>
      {clients.map((client: Client) => {
        // 🔑 Attribute-driven icon
        const iconKey = client.attributes?.["ui.icon"];
        const Icon = ICON_MAP[iconKey ?? ""] ?? App;
        return (
          <div className="app-tile-wrapper">
            <AppTile
              icon={Icon}
              name={client.name}
              href={client.baseUrl ?? ""}
              clientId={client.clientId}
            />
          </div>
        );
      })}
    </>
  );
};

export default AppGridContent;
