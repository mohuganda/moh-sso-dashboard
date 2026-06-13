import { useMemo } from "react";
import { useSelector } from "react-redux";

import { selectUser } from "./auth.selectors";
import type { Permission, System } from "./auth.types";
import {
  canLaunchSystem,
  getAccessibleSystems,
  hasAllPermissions,
  hasAnyPermission,
  hasPermission,
  hasSystem,
  hasSystemRole,
} from "./rbac";

export function useAuthorization() {
  const user = useSelector(selectUser);

  return useMemo(() => {
    const realmRoles = new Set((user?.realmRoles ?? []).map((role) => role.toLowerCase()));
    const clientRoles = user?.clientRoles ?? {};

    return {
      can: (permission: Permission) => hasPermission(user, permission),
      canAny: (requiredPermissions: Permission[]) => hasAnyPermission(user, requiredPermissions),
      canAll: (requiredPermissions: Permission[]) => hasAllPermissions(user, requiredPermissions),
      hasRealmRole: (role: string) => realmRoles.has(role.trim().toLowerCase()),
      hasClientRole: (clientId: string, role: string) =>
        (clientRoles[clientId] ?? []).some(
          (currentRole) => currentRole.trim().toLowerCase() === role.trim().toLowerCase(),
        ),
      hasSystem: (system: System | string) => hasSystem(user, system),
      hasSystemRole: (system: System | string, role: string) => hasSystemRole(user, system, role),
      canLaunchSystem: (system: System | string) => canLaunchSystem(user, system),
      permissions: user?.permissions ?? [],
      systems: user?.systems ?? [],
      accessibleSystems: getAccessibleSystems(user),
    };
  }, [user]);
}
