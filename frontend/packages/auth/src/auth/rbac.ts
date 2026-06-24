import type { AuthUser, Permission, System, SystemAccess } from "./auth.types";
import { PERMISSIONS } from "./permissions";

function normalize(value: string) {
  return value.trim().toLowerCase();
}

export function hasPermission(user: AuthUser | null | undefined, permission: Permission) {
  return user?.permissions?.includes(permission) ?? false;
}

export function hasAnyPermission(user: AuthUser | null | undefined, permissions: Permission[]) {
  return permissions.some((permission) => hasPermission(user, permission));
}

export function hasAllPermissions(user: AuthUser | null | undefined, permissions: Permission[]) {
  return permissions.every((permission) => hasPermission(user, permission));
}

export function hasSystem(user: AuthUser | null | undefined, systemClientId: System | string) {
  const expected = normalize(systemClientId);
  return (
    user?.systems?.some((system) => normalize(system) === expected) ||
    user?.accessibleSystems?.some((system) => normalize(system.clientId) === expected) ||
    false
  );
}

export function hasSystemRole(
  user: AuthUser | null | undefined,
  systemClientId: System | string,
  role: string,
) {
  const expected = normalize(role);
  return (user?.clientRoles?.[systemClientId] ?? []).some(
    (currentRole) => normalize(currentRole) === expected,
  );
}

export function canLaunchSystem(user: AuthUser | null | undefined, systemClientId: System | string) {
  return hasPermission(user, PERMISSIONS.systemsLaunch) && hasSystem(user, systemClientId);
}

export function getAccessibleSystems(user: AuthUser | null | undefined): SystemAccess[] {
  return user?.accessibleSystems ?? [];
}
