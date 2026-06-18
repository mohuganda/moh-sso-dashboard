import { useEffect, useMemo } from "react";
import { useDispatch } from "react-redux";
import { useNavigate } from "react-router-dom";

import { useAuthorization } from "@moh-sso/auth";
import { AppGridContent } from "@moh-sso/ui";
import { setActiveClient, setClients } from "@moh-sso/state";

import { buildAccessibleClients } from "@/app/access/accessClients";

type ConnectedAppGridContentProps = {
  onSelect?: () => void;
};

function isExternalUrl(url: string): boolean {
  return /^https?:\/\//i.test(url);
}

function normalizePortalPath(href: string): string {
  if (isExternalUrl(href)) {
    return href;
  }

  if (href === "/portal") {
    return "/apps";
  }

  if (href.startsWith("/portal/")) {
    return href.replace(/^\/portal/, "");
  }

  return href;
}

export function ConnectedAppGridContent({ onSelect }: ConnectedAppGridContentProps) {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { accessibleSystems } = useAuthorization();
  const visibleClients = useMemo(
    () =>
      buildAccessibleClients({
        accessibleSystems,
      }),
    [accessibleSystems],
  );

  useEffect(() => {
    dispatch(setClients(visibleClients));
  }, [dispatch, visibleClients]);

  const handleOpenClient = (href: string, clientId?: string) => {
    if (clientId) {
      dispatch(setActiveClient(clientId));
    }

    const targetHref = normalizePortalPath(href);

    if (isExternalUrl(targetHref)) {
      window.open(targetHref, "_blank", "noopener,noreferrer");
      return;
    }

    navigate(targetHref);
  };

  return (
    <AppGridContent
      clients={visibleClients}
      onSelect={onSelect}
      onOpenClient={handleOpenClient}
    />
  );
}
