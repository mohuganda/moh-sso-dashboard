export type RbacPermission = {
  id: string;
  key: string;
  displayName?: string;
  description?: string;
  category?: string;
  status?: string;
  systemRoleUsageCount?: number;
  realmRoleUsageCount?: number;
};

export type RbacSystem = {
  id: string;
  clientId: string;
  displayName: string;
  description?: string;
  icon?: string;
  launchUrl?: string;
  category?: string;
  ownerTeam?: string;
  ownerName?: string;
  ownerEmail?: string;
  supportUrl?: string;
  documentationUrl?: string;
  environment?: string;
  criticality?: string;
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
  ownerTeam?: string;
  ownerName?: string;
  ownerEmail?: string;
  supportUrl?: string;
  documentationUrl?: string;
  environment?: string;
  criticality?: string;
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

export type RbacDriftSummary = {
  systemsInSync: number;
  systemsMissingInRbac: number;
  systemsStaleInRbac: number;
  rolesMissingInRbac: number;
  rolesStaleInRbac: number;
  realmRolesMissing: number;
};

export type RbacDriftSystem = {
  clientId: string;
  displayName: string;
  keycloakStatus: string;
  rbacStatus: string;
  missingRolesInRbac: string[];
  staleRolesInRbac: string[];
  missingAccessRoles: string[];
};

export type RbacDriftRealmRole = {
  name: string;
  status: string;
  rbacStatus: string;
};

export type RbacDriftReport = {
  source: string;
  summary: RbacDriftSummary;
  systems: RbacDriftSystem[];
  realmRoles: RbacDriftRealmRole[];
  warnings?: string[];
};

export type RbacSyncPayload = {
  source?: string;
  realmExport: unknown;
};

export type RbacSyncPreview = {
  report: RbacDriftReport;
  systemsToCreate: number;
  systemsToUpdate: number;
  rolesToCreate: number;
  accessRolesToAdd: number;
  realmRolesToTrack: number;
};

export type RbacSyncApplyResult = {
  preview: RbacSyncPreview;
  systemsSynced: number;
  rolesSynced: number;
  accessRolesSynced: number;
};

export type RbacEffectiveAccessUser = {
  id: string;
  username: string;
  email: string;
  fullName: string;
  enabled: boolean;
  isAdmin: boolean;
  isResolved: boolean;
};

export type RbacPermissionGrantSource = {
  permissionKey: string;
  grantedByType: "realmRole" | "clientRole" | string;
  role: string;
  systemClientId?: string;
  systemName?: string;
};

export type RbacSystemAccessSummary = {
  clientId: string;
  displayName: string;
  launchUrl?: string;
  icon?: string;
  category?: string;
  roles: string[];
};

export type RbacEffectiveAccess = {
  user: RbacEffectiveAccessUser;
  realmRoles: string[];
  clientRoles: Record<string, string[]>;
  permissions: RbacPermission[];
  accessibleSystems: RbacSystemAccessSummary[];
  grantSources: RbacPermissionGrantSource[];
};

export type RbacPermissionMetadataPayload = {
  displayName?: string;
  description?: string;
  category?: string;
  status?: string;
};

export type RbacRoleUsage = {
  role: RbacSystemRole;
  permissionsCount: number;
  assignedUserCount: number;
  warnings: string[];
};

export type RbacRealmRoleUsage = {
  realmRole: string;
  permissions: RbacPermission[];
  permissionsCount: number;
  warnings: string[];
};

export type RbacChangePreviewPayload = {
  action: string;
  resourceType: string;
  resourceId?: string;
  systemClientId?: string;
  roleName?: string;
  permissionKey?: string;
};

export type RbacChangePreview = {
  action: string;
  resourceType: string;
  resourceId?: string;
  highRisk: boolean;
  riskLevel: string;
  warnings: string[];
  permissionsAdded: string[];
  permissionsRemoved: string[];
  affectedSystems: string[];
  affectedUsersCount: number;
};

export type RbacImportPayload = {
  payload: unknown;
  format?: string;
  prune?: boolean;
};

export type RbacImportPreview = {
  systemsToCreate: number;
  systemsToUpdate: number;
  rolesToCreate: number;
  warnings: string[];
};

export type RbacImportApplyResult = {
  preview: RbacImportPreview;
  applied: boolean;
};

export type RbacRoleTemplate = {
  name: string;
  displayName: string;
  description: string;
  permissions: string[];
};

export type RbacRoleFromTemplatePayload = {
  templateName: string;
  roleName?: string;
  displayName?: string;
};

export type RbacCopyPermissionsPayload = {
  sourceRoleId: string;
};

export type RbacBulkPermissionPayload = {
  roleIds: string[];
  permissionKey: string;
};

export type RbacAuditEvent = {
  id: string;
  actorUserId?: string;
  action: string;
  resourceType: string;
  resourceId?: string;
  systemClientId?: string;
  roleName?: string;
  permissionKey?: string;
  details?: unknown;
  createdAt: string;
};

export type RbacAuditFilter = {
  actor?: string;
  systemClientId?: string;
  roleName?: string;
  permissionKey?: string;
  action?: string;
  from?: string;
  to?: string;
  limit?: number;
};

export type RbacAccessRequest = {
  id: string;
  userId?: string;
  username?: string;
  email?: string;
  systemClientId: string;
  requestedRole: string;
  reason?: string;
  status: string;
  requestedBy?: string;
  reviewedBy?: string;
  reviewedAt?: string;
  decisionNote?: string;
  createdAt: string;
  updatedAt: string;
};

export type RbacAccessRequestPayload = {
  userId?: string;
  username?: string;
  email?: string;
  systemClientId: string;
  requestedRole: string;
  reason?: string;
};

export type RbacDecisionPayload = {
  note?: string;
};

export type RbacChangeRequest = {
  id: string;
  requestedBy?: string;
  reviewedBy?: string;
  status: string;
  action: string;
  resourceType: string;
  resourceId?: string;
  payload?: unknown;
  riskLevel: string;
  reason?: string;
  decisionNote?: string;
  reviewedAt?: string;
  createdAt: string;
  updatedAt: string;
};

export type RbacChangeRequestPayload = {
  action: string;
  resourceType: string;
  resourceId?: string;
  payload?: unknown;
  riskLevel?: string;
  reason?: string;
};

export type RbacSimulationPayload = {
  userId?: string;
  username?: string;
  email?: string;
  realmRoles: string[];
  clientRoles: Record<string, string[]>;
  addRealmRoles?: string[];
  removeRealmRoles?: string[];
  addClientRoles?: Record<string, string[]>;
  removeClientRoles?: Record<string, string[]>;
  addPermissions?: string[];
  removePermissions?: string[];
};

export type RbacSimulationResult = {
  baselinePermissions: RbacPermission[];
  permissions: RbacPermission[];
  addedPermissions: string[];
  removedPermissions: string[];
  accessibleSystems: RbacSystemAccessSummary[];
  grantSources: RbacPermissionGrantSource[];
  warnings: string[];
};
