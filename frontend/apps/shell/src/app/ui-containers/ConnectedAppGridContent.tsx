import { useEffect, useMemo } from "react";
import { useDispatch } from "react-redux";
import { useNavigate } from "react-router-dom";

import { useListClientsQuery } from "@moh-sso/api";
import { useAuthorization } from "@moh-sso/auth";
import { AppGridContent } from "@moh-sso/ui";
import { setActiveClient, setClients } from "@moh-sso/state";

type ConnectedAppGridContentProps = {
  onSelect?: () => void;
};

function isExternalUrl(url: string): boolean {
  return /^https?:\/\//i.test(url);
}

export function ConnectedAppGridContent({ onSelect }: ConnectedAppGridContentProps) {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { data: clients = [], isLoading, isError, refetch } = useListClientsQuery();
  const { canLaunchSystem } = useAuthorization();
  const visibleClients = useMemo(
    () =>
      clients.filter((client) => {
        if (!client.clientId) {
          return false;
        }
        return canLaunchSystem(client.clientId);
      }),
    [canLaunchSystem, clients],
  );

  useEffect(() => {
    dispatch(setClients(visibleClients));
  }, [dispatch, visibleClients]);

  const handleOpenClient = (href: string, clientId?: string) => {
    if (clientId) {
      dispatch(setActiveClient(clientId));
    }

    if (isExternalUrl(href)) {
      window.open(href, "_blank", "noopener,noreferrer");
      return;
    }

    navigate(href);
  };

  return (
    <AppGridContent
      clients={visibleClients}
      isLoading={isLoading}
      isError={isError}
      onRetry={refetch}
      onSelect={onSelect}
      onOpenClient={handleOpenClient}
    />
  );
}
