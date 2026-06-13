import { useMemo } from "react";
import { useSelector } from "react-redux";

import { selectUser } from "./auth.selectors";
import type { Permission, System } from "./auth.types";

export function useAuthorization() {
  const user = useSelector(selectUser);

  return useMemo(() => {
    const permissions = new Set(user?.permissions ?? []);
    const realmRoles = new Set((user?.realmRoles ?? []).map((role) => role.toLowerCase()));
    const clientRoles = user?.clientRoles ?? {};
    const systems = new Set(user?.systems ?? []);

    return {
      can: (permission: Permission) => permissions.has(permission),
      canAny: (requiredPermissions: Permission[]) =>
        requiredPermissions.some((permission) => permissions.has(permission)),
      hasRealmRole: (role: string) => realmRoles.has(role.trim().toLowerCase()),
      hasClientRole: (clientId: string, role: string) =>
        (clientRoles[clientId] ?? []).some(
          (currentRole) => currentRole.trim().toLowerCase() === role.trim().toLowerCase(),
        ),
      hasSystem: (system: System) => systems.has(system),
      hasSystemRole: (system: System, role: string) =>
        (clientRoles[system] ?? []).some(
          (currentRole) => currentRole.trim().toLowerCase() === role.trim().toLowerCase(),
        ),
      permissions: [...permissions],
      systems: [...systems],
    };
  }, [user]);
}
