import { InlineLoading } from "@carbon/react";
import type { JSX } from "react";
import { useSelector } from "react-redux";
import { Navigate } from "react-router-dom";

import {
  selectAuthenticated,
  selectAuthLoaded,
  type Permission,
  type System,
  useAuthorization,
} from "@moh-sso/auth";

type PermissionRouteProps = {
  permission?: Permission;
  anyOf?: Permission[];
  allOf?: Permission[];
  systems?: Array<System | string>;
  systemRoles?: Array<{ system: System | string; role: string }>;
  children: JSX.Element;
};

export function PermissionRoute({
  permission,
  anyOf,
  allOf,
  systems,
  systemRoles,
  children,
}: PermissionRouteProps) {
  const loaded = useSelector(selectAuthLoaded);
  const authenticated = useSelector(selectAuthenticated);
  const { can, canAny, canAll, hasSystem, hasSystemRole } = useAuthorization();

  if (!loaded) {
    return <InlineLoading description="Checking permissions..." />;
  }

  if (!authenticated) {
    return <Navigate to="/" replace />;
  }

  const hasRequiredPermission = permission ? can(permission) : true;
  const hasAnyRequiredPermission = anyOf && anyOf.length > 0 ? canAny(anyOf) : true;
  const hasAllRequiredPermissions = allOf && allOf.length > 0 ? canAll(allOf) : true;
  const hasRequiredSystems =
    systems && systems.length > 0 ? systems.every((system) => hasSystem(system)) : true;
  const hasRequiredSystemRoles =
    systemRoles && systemRoles.length > 0
      ? systemRoles.every(({ system, role }) => hasSystemRole(system, role))
      : true;

  if (
    !hasRequiredPermission ||
    !hasAnyRequiredPermission ||
    !hasAllRequiredPermissions ||
    !hasRequiredSystems ||
    !hasRequiredSystemRoles
  ) {
    return <Navigate to="/forbidden" replace />;
  }

  return children;
}
