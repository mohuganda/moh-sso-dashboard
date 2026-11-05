import { MICROFRONTEND_NAVIGATION_EVENT } from "@moh-sso/microfrontend";
import { BreadcrumbBar } from "@moh-sso/ui";
import { useSyncExternalStore } from "react";
import { useHref, useLocation, useNavigate } from "react-router-dom";

import { buildBreadcrumbs } from "./breadcrumbs";

function subscribeToRouteChanges(listener: () => void) {
  window.addEventListener(MICROFRONTEND_NAVIGATION_EVENT, listener);
  window.addEventListener("popstate", listener);
  window.addEventListener("single-spa:routing-event", listener);
  return () => {
    window.removeEventListener(MICROFRONTEND_NAVIGATION_EVENT, listener);
    window.removeEventListener("popstate", listener);
    window.removeEventListener("single-spa:routing-event", listener);
  };
}

const currentPath = () => window.location.pathname;

export function RouteBreadcrumbBar() {
  const location = useLocation();
  const navigate = useNavigate();

  // Embedded routers can navigate without updating the shell router context.
  // Subscribe to the shared URL as well as normal shell location changes.
  const pathname = useSyncExternalStore(
    subscribeToRouteChanges,
    currentPath,
    () => location.pathname,
  );
  const deploymentBase = useHref("/").replace(/\/+$/, "");
  const routePath =
    deploymentBase && (pathname === deploymentBase || pathname.startsWith(`${deploymentBase}/`))
      ? pathname.slice(deploymentBase.length) || "/"
      : pathname;
  const items = buildBreadcrumbs(routePath);

  if (items.length <= 1) {
    return null;
  }

  return (
    <BreadcrumbBar items={items} onNavigate={(href) => navigate(href)} showBackButton={false} />
  );
}
