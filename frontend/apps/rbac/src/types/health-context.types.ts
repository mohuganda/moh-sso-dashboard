export type HealthContextType =
  | "NATIONAL"
  | "REGION"
  | "DISTRICT"
  | "CITY"
  | "DIVISION"
  | "MUNICIPALITY"
  | "COUNTY"
  | "SUB_COUNTY"
  | "PARISH"
  | "FACILITY"
  | "PROGRAM"
  | "DEPARTMENT"
  | "TEAM"
  | "CUSTOM";

export type HealthContextScopeMode = "NODE_ONLY" | "NODE_AND_DESCENDANTS";

export type HealthContextNode = {
  id: string;
  code: string;
  name: string;
  contextType: HealthContextType;
  parentId?: string;
  source: string;
  metadata: Record<string, unknown>;
  enabled: boolean;
  version: number;
};

export type HealthContextAlias = {
  id: string;
  contextNodeId: string;
  namespace: string;
  externalId: string;
  metadata: Record<string, unknown>;
  createdAt: string;
};

export type HealthContextAssignment = {
  id: string;
  userId?: string;
  groupId?: string;
  contextNodeId: string;
  scopeMode: HealthContextScopeMode;
  isDefault: boolean;
  validFrom?: string;
  validUntil?: string;
  source: string;
  sourceReference?: string;
};

export type HealthContextNodePayload = {
  code: string;
  name: string;
  contextType: HealthContextType;
  parentId?: string | null;
  source: string;
  metadata: Record<string, unknown>;
  enabled: boolean;
  version?: number;
};

export type HealthContextAssignmentPayload = {
  contextNodeId: string;
  scopeMode: HealthContextScopeMode;
  isDefault: boolean;
  validFrom?: string;
  validUntil?: string;
  source: string;
  sourceReference?: string;
};

export type HealthContextGroupMappingDrift = {
  groupId: string;
  groupPath: string;
  keycloakGroupId?: string;
  groupEnabled: boolean;
  contextNodeId?: string;
  contextCode?: string;
  contextName?: string;
  contextEnabled: boolean;
  scopeMode?: HealthContextScopeMode;
  status:
    | "IN_SYNC"
    | "MISSING_KEYCLOAK_LINK"
    | "DISABLED_GROUP"
    | "UNMAPPED"
    | "INACTIVE_CONTEXT";
  recommendedAction: string;
};

export type HealthContextDriftReport = {
  generatedAt: string;
  total: number;
  inSync: number;
  drifted: number;
  items: HealthContextGroupMappingDrift[];
};

export type HealthContextSyncMapping = {
  groupId: string;
  contextNodeId: string;
  scopeMode: HealthContextScopeMode;
};

export type HealthContextSyncRequest = {
  mappings: HealthContextSyncMapping[];
  replaceExisting: boolean;
};

export type HealthContextSyncPreview = {
  valid: boolean;
  mappingCount: number;
  affectedGroups: number;
  replaceExisting: boolean;
  drift: HealthContextDriftReport;
};

export type HealthContextSyncResult = {
  appliedMappings: number;
  affectedGroups: number;
};
