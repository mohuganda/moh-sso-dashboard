import { useEffect } from "react";
import { useDispatch } from "react-redux";
import { useNavigate } from "react-router-dom";

import { useListClientsQuery } from "@moh-sso/api";
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

  useEffect(() => {
    dispatch(setClients(clients));
  }, [clients, dispatch]);

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
      clients={clients}
      isLoading={isLoading}
      isError={isError}
      onRetry={refetch}
      onSelect={onSelect}
      onOpenClient={handleOpenClient}
    />
  );
}
