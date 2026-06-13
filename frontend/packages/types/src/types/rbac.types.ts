export type RbacPermission = {
  id: string;
  key: string;
  displayName?: string;
  description?: string;
  category?: string;
};

export type RbacSystem = {
  id: string;
  clientId: string;
  displayName: string;
  description?: string;
  icon?: string;
  launchUrl?: string;
  category?: string;
  enabled: boolean;
  sortOrder: number;
};

export type RbacSystemRole = {
  id: string;
  systemId: string;
  clientId: string;
  name: string;
  displayName?: string;
  description?: string;
  enabled: boolean;
  permissions: RbacPermission[];
};

export type RbacSystemDetail = RbacSystem & {
  accessRoles: string[];
  roles: RbacSystemRole[];
};

export type RbacRealmRolePermissionGroup = {
  realmRole: string;
  permissions: RbacPermission[];
};

export type UpsertRbacSystemPayload = {
  displayName: string;
  description?: string;
  icon?: string;
  launchUrl?: string;
  category?: string;
  enabled?: boolean;
  sortOrder?: number;
};

export type RbacRolePayload = {
  name: string;
  displayName?: string;
  description?: string;
  enabled?: boolean;
};

export type RbacPermissionPayload = {
  permissionKey: string;
};

export type RbacAccessRolePayload = {
  roleName: string;
};
