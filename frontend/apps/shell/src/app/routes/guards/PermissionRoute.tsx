import { InlineLoading } from "@carbon/react";
import type { JSX } from "react";
import { useSelector } from "react-redux";
import { Navigate } from "react-router-dom";

import {
  selectAuthenticated,
  selectAuthLoaded,
  selectHasAnyPermission,
  selectHasPermission,
  type Permission,
} from "@moh-sso/auth";

type PermissionRouteProps = {
  permission?: Permission;
  anyOf?: Permission[];
  children: JSX.Element;
};

export function PermissionRoute({ permission, anyOf, children }: PermissionRouteProps) {
  const loaded = useSelector(selectAuthLoaded);
  const authenticated = useSelector(selectAuthenticated);
  const hasPermission = useSelector(
    permission ? selectHasPermission(permission) : () => true,
  );
  const hasAnyPermission = useSelector(
    anyOf && anyOf.length > 0 ? selectHasAnyPermission(anyOf) : () => true,
  );

  if (!loaded) {
    return <InlineLoading description="Checking permissions..." />;
  }

  if (!authenticated) {
    return <Navigate to="/" replace />;
  }

  if (!hasPermission || !hasAnyPermission) {
    return <Navigate to="/" replace />;
  }

  return children;
}

