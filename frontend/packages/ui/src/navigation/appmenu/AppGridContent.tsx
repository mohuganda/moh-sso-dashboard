import React, { useMemo } from "react";
import { InlineLoading, Tile } from "@carbon/react";
import {
  App,
  Calendar,
  ChartLineData,
  CloudUpload,
  Document,
  Email,
  IbmJrs,
  Menu,
  Settings,
  Task,
  User,
  WatsonHealthTextAnnotationToggle,
} from "@carbon/react/icons";

import type { Client } from "@moh-sso/types";

import AppTile from "./AppMenuItem.component";
import "./AppMenu.scss";

const ICON_MAP: Record<string, React.ElementType> = {
  app: App,
  apps: App,
  applications: App,
  calendar: Calendar,
  chart: ChartLineData,
  dashboard: ChartLineData,
  document: Document,
  documents: Document,
  upload: CloudUpload,
  email: Email,
  mail: Email,
  menu: Menu,
  settings: Settings,
  user: User,
  profile: User,
  reporting: IbmJrs,
  reports: IbmJrs,
  tasks: Task,
  action: Task,
  tracker: Task,
  terminology: WatsonHealthTextAnnotationToggle,
};

type AppGridContentProps = {
  clients?: Client[];
  isLoading?: boolean;
  isError?: boolean;
  onSelect?: () => void;
  onRetry?: () => void;
  onOpenClient?: (href: string, clientId?: string) => void;
};

function getClientHref(client: Client): string {
  return client.baseUrl?.trim() || "#";
}

function getClientName(client: Client): string {
  return client.name || client.clientId || "Application";
}

function getClientIcon(client: Client): React.ElementType {
  const iconKey = client.attributes?.["ui.icon"]?.trim().toLowerCase();

  if (!iconKey) {
    return App;
  }

  return ICON_MAP[iconKey] ?? App;
}

function getClientOrder(client: Client): number {
  const rawOrder = client.attributes?.["ui.order"];
  const order = Number(rawOrder);

  return Number.isFinite(order) ? order : Number.MAX_SAFE_INTEGER;
}

const AppGridContent: React.FC<AppGridContentProps> = ({
  clients = [],
  isLoading = false,
  isError = false,
  onSelect,
  onRetry,
  onOpenClient,
}) => {
  const sortedClients = useMemo(() => {
    return [...clients].sort((a, b) => {
      const orderDiff = getClientOrder(a) - getClientOrder(b);

      if (orderDiff !== 0) {
        return orderDiff;
      }

      return getClientName(a).localeCompare(getClientName(b));
    });
  }, [clients]);

  if (isLoading) {
    return (
      <div className="app-grid-state">
        <InlineLoading description="Loading applications…" />
      </div>
    );
  }

  if (isError) {
    return (
      <div className="app-grid-state">
        <Tile className="app-grid-state__tile">
          <h5 className="app-grid-state__title">Failed to load applications</h5>
          <p className="app-grid-state__description">Please check your connection and try again.</p>

          <button type="button" className="app-grid-retry" onClick={onRetry}>
            Retry
          </button>
        </Tile>
      </div>
    );
  }

  if (sortedClients.length === 0) {
    return (
      <div className="app-grid-state">
        <Tile className="app-grid-state__tile">
          <h5 className="app-grid-state__title">No applications available</h5>
          <p className="app-grid-state__description">
            Applications assigned to your account will appear here.
          </p>
        </Tile>
      </div>
    );
  }

  return (
    <div className="app-grid" role="list">
      {sortedClients.map((client: Client) => {
        const Icon = getClientIcon(client);

        return (
          <div className="app-grid__item" role="listitem" key={client.clientId || client.id}>
            <AppTile
              icon={Icon}
              name={getClientName(client)}
              href={getClientHref(client)}
              clientId={client.clientId}
              onSelect={onSelect}
              onOpen={onOpenClient}
            />
          </div>
        );
      })}
    </div>
  );
};

export default AppGridContent;
