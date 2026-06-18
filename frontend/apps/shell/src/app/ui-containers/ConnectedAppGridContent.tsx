import { useEffect, useMemo } from "react";
import { useDispatch } from "react-redux";
import { useNavigate } from "react-router-dom";

import { PERMISSIONS, useAuthorization } from "@moh-sso/auth";
import { AppGridContent } from "@moh-sso/ui";
import { setActiveClient, setClients } from "@moh-sso/state";

import { buildAccessibleClients } from "@/app/access/accessClients";

type ConnectedAppGridContentProps = {
  onSelect?: () => void;
};

const DATA_STATISTICS_CLIENT_ID = "__default__";

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
  const { accessibleSystems, canAny } = useAuthorization();
  const canAccessDataStatistics = canAny([
    PERMISSIONS.dataQualityRead,
    PERMISSIONS.documentsRead,
    PERMISSIONS.reportBrowserRead,
    PERMISSIONS.surveillanceRead,
  ]);
  const visibleClients = useMemo(
    () =>
      buildAccessibleClients({
        accessibleSystems,
        includeDataStatistics: canAccessDataStatistics,
      }),
    [accessibleSystems, canAccessDataStatistics],
  );

  useEffect(() => {
    dispatch(
      setClients(
        visibleClients.filter((client) => client.clientId !== DATA_STATISTICS_CLIENT_ID),
      ),
    );
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
