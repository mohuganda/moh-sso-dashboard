import { useMemo } from "react";
import { useSelector } from "react-redux";

import { selectUser } from "./auth.selectors";
import type { Permission } from "./auth.types";

export function useAuthorization() {
  const user = useSelector(selectUser);

  return useMemo(() => {
    const permissions = new Set(user?.permissions ?? []);
    const realmRoles = new Set((user?.realmRoles ?? []).map((role) => role.toLowerCase()));
    const clientRoles = user?.clientRoles ?? {};

    return {
      can: (permission: Permission) => permissions.has(permission),
      canAny: (requiredPermissions: Permission[]) =>
        requiredPermissions.some((permission) => permissions.has(permission)),
      hasRealmRole: (role: string) => realmRoles.has(role.trim().toLowerCase()),
      hasClientRole: (clientId: string, role: string) =>
        (clientRoles[clientId] ?? []).some(
          (currentRole) => currentRole.trim().toLowerCase() === role.trim().toLowerCase(),
        ),
      permissions: [...permissions],
    };
  }, [user]);
}

