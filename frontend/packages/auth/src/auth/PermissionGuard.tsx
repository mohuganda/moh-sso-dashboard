import type { ReactNode } from "react";

import type { Permission } from "./auth.types";
import { useAuthorization } from "./useAuthorization";

type PermissionGuardProps = {
  permission?: Permission;
  anyOf?: Permission[];
  fallback?: ReactNode;
  children: ReactNode;
};

export function PermissionGuard({
  permission,
  anyOf,
  fallback = null,
  children,
}: PermissionGuardProps) {
  const { can, canAny } = useAuthorization();

  const allowed =
    (permission ? can(permission) : true) && (anyOf && anyOf.length > 0 ? canAny(anyOf) : true);

  return allowed ? children : fallback;
}

