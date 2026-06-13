import type {
  RbacAccessRolePayload,
  RbacAccessRequest,
  RbacAccessRequestPayload,
  RbacAuditFilter,
  RbacAuditEvent,
  RbacBulkPermissionPayload,
  RbacChangePreview,
  RbacChangePreviewPayload,
  RbacChangeRequest,
  RbacChangeRequestPayload,
  RbacCopyPermissionsPayload,
  RbacDecisionPayload,
  RbacDriftReport,
  RbacEffectiveAccess,
  RbacImportApplyResult,
  RbacImportPayload,
  RbacImportPreview,
  RbacPermission,
  RbacPermissionMetadataPayload,
  RbacPermissionPayload,
  RbacRealmRolePermissionGroup,
  RbacRealmRoleUsage,
  RbacRoleFromTemplatePayload,
  RbacRolePayload,
  RbacRoleTemplate,
  RbacRoleUsage,
  RbacSimulationPayload,
  RbacSimulationResult,
  RbacSyncApplyResult,
  RbacSyncPayload,
  RbacSyncPreview,
  RbacSystem,
  RbacSystemDetail,
  RbacSystemRole,
  UpsertRbacSystemPayload,
} from "@moh-sso/types";

import { baseApi } from "./baseApi";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

const base = "/admin/rbac";

export const rbacApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    listRbacSystems: builder.query<RbacSystem[], void>({
      query: () => `${base}/systems`,
      transformResponse: (res: ApiEnvelope<RbacSystem[]>) => res.data,
      providesTags: [{ type: "RbacSystem", id: "LIST" }],
    }),
    getRbacSystem: builder.query<RbacSystemDetail, string>({
      query: (clientId) => `${base}/systems/${encodeURIComponent(clientId)}`,
      transformResponse: (res: ApiEnvelope<RbacSystemDetail>) => res.data,
      providesTags: (_result, _error, clientId) => [{ type: "RbacSystem", id: clientId }],
    }),
    updateRbacSystem: builder.mutation<
      RbacSystem,
      { clientId: string; data: UpsertRbacSystemPayload }
    >({
      query: ({ clientId, data }) => ({
        url: `${base}/systems/${encodeURIComponent(clientId)}`,
        method: "PUT",
        body: data,
      }),
      transformResponse: (res: ApiEnvelope<RbacSystem>) => res.data,
      invalidatesTags: (_result, _error, { clientId }) => [
        { type: "RbacSystem", id: "LIST" },
        { type: "RbacSystem", id: clientId },
      ],
    }),
    listRbacPermissions: builder.query<RbacPermission[], void>({
      query: () => `${base}/permissions`,
      transformResponse: (res: ApiEnvelope<RbacPermission[]>) => res.data,
      providesTags: [{ type: "RbacPermission", id: "LIST" }],
    }),
    updateRbacPermissionMetadata: builder.mutation<
      RbacPermission,
      { permissionKey: string; data: RbacPermissionMetadataPayload }
    >({
      query: ({ permissionKey, data }) => ({
        url: `${base}/permissions/${encodeURIComponent(permissionKey)}`,
        method: "PUT",
        body: data,
      }),
      transformResponse: (res: ApiEnvelope<RbacPermission>) => res.data,
      invalidatesTags: [{ type: "RbacPermission", id: "LIST" }],
    }),
    listSystemRoles: builder.query<RbacSystemRole[], string>({
      query: (clientId) => `${base}/systems/${encodeURIComponent(clientId)}/roles`,
      transformResponse: (res: ApiEnvelope<RbacSystemRole[]>) => res.data,
      providesTags: (_result, _error, clientId) => [{ type: "RbacSystem", id: clientId }],
    }),
    createSystemRole: builder.mutation<
      RbacSystemRole,
      { clientId: string; data: RbacRolePayload }
    >({
      query: ({ clientId, data }) => ({
        url: `${base}/systems/${encodeURIComponent(clientId)}/roles`,
        method: "POST",
        body: data,
      }),
      transformResponse: (res: ApiEnvelope<RbacSystemRole>) => res.data,
      invalidatesTags: (_result, _error, { clientId }) => [{ type: "RbacSystem", id: clientId }],
    }),
    updateSystemRole: builder.mutation<
      RbacSystemRole,
      { roleId: string; clientId: string; data: RbacRolePayload }
    >({
      query: ({ roleId, data }) => ({
        url: `${base}/roles/${encodeURIComponent(roleId)}`,
        method: "PUT",
        body: data,
      }),
      transformResponse: (res: ApiEnvelope<RbacSystemRole>) => res.data,
      invalidatesTags: (_result, _error, { clientId }) => [{ type: "RbacSystem", id: clientId }],
    }),
    deleteSystemRole: builder.mutation<void, { roleId: string; clientId: string }>({
      query: ({ roleId }) => ({
        url: `${base}/roles/${encodeURIComponent(roleId)}`,
        method: "DELETE",
      }),
      invalidatesTags: (_result, _error, { clientId }) => [{ type: "RbacSystem", id: clientId }],
    }),
    getRoleUsage: builder.query<RbacRoleUsage, string>({
      query: (roleId) => `${base}/roles/${encodeURIComponent(roleId)}/usage`,
      transformResponse: (res: ApiEnvelope<RbacRoleUsage>) => res.data,
    }),
    copyRolePermissions: builder.mutation<
      RbacSystemRole,
      { roleId: string; clientId: string; data: RbacCopyPermissionsPayload }
    >({
      query: ({ roleId, data }) => ({
        url: `${base}/roles/${encodeURIComponent(roleId)}/copy-permissions`,
        method: "POST",
        body: data,
      }),
      transformResponse: (res: ApiEnvelope<RbacSystemRole>) => res.data,
      invalidatesTags: (_result, _error, { clientId }) => [{ type: "RbacSystem", id: clientId }],
    }),
    assignSystemRolePermission: builder.mutation<
      void,
      { roleId: string; clientId: string; data: RbacPermissionPayload }
    >({
      query: ({ roleId, data }) => ({
        url: `${base}/roles/${encodeURIComponent(roleId)}/permissions`,
        method: "POST",
        body: data,
      }),
      invalidatesTags: (_result, _error, { clientId }) => [{ type: "RbacSystem", id: clientId }],
    }),
    removeSystemRolePermission: builder.mutation<
      void,
      { roleId: string; clientId: string; permissionKey: string }
    >({
      query: ({ roleId, permissionKey }) => ({
        url: `${base}/roles/${encodeURIComponent(roleId)}/permissions/${encodeURIComponent(permissionKey)}`,
        method: "DELETE",
      }),
      invalidatesTags: (_result, _error, { clientId }) => [{ type: "RbacSystem", id: clientId }],
    }),
    listRealmRolePermissions: builder.query<RbacRealmRolePermissionGroup[], void>({
      query: () => `${base}/realm-roles`,
      transformResponse: (res: ApiEnvelope<RbacRealmRolePermissionGroup[]>) => res.data,
      providesTags: [{ type: "RbacRealmRole", id: "LIST" }],
    }),
    getRealmRoleUsage: builder.query<RbacRealmRoleUsage, string>({
      query: (realmRole) => `${base}/realm-roles/${encodeURIComponent(realmRole)}/usage`,
      transformResponse: (res: ApiEnvelope<RbacRealmRoleUsage>) => res.data,
    }),
    assignRealmRolePermission: builder.mutation<
      void,
      { realmRole: string; data: RbacPermissionPayload }
    >({
      query: ({ realmRole, data }) => ({
        url: `${base}/realm-roles/${encodeURIComponent(realmRole)}/permissions`,
        method: "POST",
        body: data,
      }),
      invalidatesTags: [{ type: "RbacRealmRole", id: "LIST" }],
    }),
    removeRealmRolePermission: builder.mutation<
      void,
      { realmRole: string; permissionKey: string }
    >({
      query: ({ realmRole, permissionKey }) => ({
        url: `${base}/realm-roles/${encodeURIComponent(realmRole)}/permissions/${encodeURIComponent(permissionKey)}`,
        method: "DELETE",
      }),
      invalidatesTags: [{ type: "RbacRealmRole", id: "LIST" }],
    }),
    addSystemAccessRole: builder.mutation<
      void,
      { clientId: string; data: RbacAccessRolePayload }
    >({
      query: ({ clientId, data }) => ({
        url: `${base}/systems/${encodeURIComponent(clientId)}/access-roles`,
        method: "POST",
        body: data,
      }),
      invalidatesTags: (_result, _error, { clientId }) => [{ type: "RbacSystem", id: clientId }],
    }),
    removeSystemAccessRole: builder.mutation<
      void,
      { clientId: string; roleName: string; force?: boolean }
    >({
      query: ({ clientId, roleName, force }) => ({
        url: `${base}/systems/${encodeURIComponent(clientId)}/access-roles/${encodeURIComponent(roleName)}${force ? "?force=true" : ""}`,
        method: "DELETE",
      }),
      invalidatesTags: (_result, _error, { clientId }) => [{ type: "RbacSystem", id: clientId }],
    }),
    getRbacDrift: builder.query<RbacDriftReport, void>({
      query: () => `${base}/drift`,
      transformResponse: (res: ApiEnvelope<RbacDriftReport>) => res.data,
      providesTags: [{ type: "RbacSystem", id: "DRIFT" }],
    }),
    previewRbacRealmExportDrift: builder.mutation<RbacDriftReport, RbacSyncPayload>({
      query: (data) => ({
        url: `${base}/drift/realm-export`,
        method: "POST",
        body: data,
      }),
      transformResponse: (res: ApiEnvelope<RbacDriftReport>) => res.data,
    }),
    previewRbacSync: builder.mutation<RbacSyncPreview, RbacSyncPayload>({
      query: (data) => ({
        url: `${base}/sync/preview`,
        method: "POST",
        body: data,
      }),
      transformResponse: (res: ApiEnvelope<RbacSyncPreview>) => res.data,
    }),
    applyRbacSync: builder.mutation<RbacSyncApplyResult, RbacSyncPayload>({
      query: (data) => ({
        url: `${base}/sync/apply`,
        method: "POST",
        body: data,
      }),
      transformResponse: (res: ApiEnvelope<RbacSyncApplyResult>) => res.data,
      invalidatesTags: [
        { type: "RbacSystem", id: "LIST" },
        { type: "RbacSystem", id: "DRIFT" },
        { type: "RbacRealmRole", id: "LIST" },
      ],
    }),
    getRbacEffectiveAccess: builder.query<
      RbacEffectiveAccess,
      { userId?: string; username?: string; email?: string }
    >({
      query: ({ userId = "lookup", username, email }) => {
        const params = new URLSearchParams();
        if (username) params.set("username", username);
        if (email) params.set("email", email);
        const query = params.toString();
        return `${base}/effective-access/users/${encodeURIComponent(userId)}${query ? `?${query}` : ""}`;
      },
      transformResponse: (res: ApiEnvelope<RbacEffectiveAccess>) => res.data,
    }),
    previewRbacChange: builder.mutation<RbacChangePreview, RbacChangePreviewPayload>({
      query: (data) => ({ url: `${base}/changes/preview`, method: "POST", body: data }),
      transformResponse: (res: ApiEnvelope<RbacChangePreview>) => res.data,
    }),
    exportRbacSeed: builder.query<unknown, void>({
      query: () => `${base}/export`,
      transformResponse: (res: ApiEnvelope<unknown>) => res.data,
    }),
    previewRbacImport: builder.mutation<RbacImportPreview, RbacImportPayload>({
      query: (data) => ({ url: `${base}/import/preview`, method: "POST", body: data }),
      transformResponse: (res: ApiEnvelope<RbacImportPreview>) => res.data,
    }),
    applyRbacImport: builder.mutation<RbacImportApplyResult, RbacImportPayload>({
      query: (data) => ({ url: `${base}/import/apply`, method: "POST", body: data }),
      transformResponse: (res: ApiEnvelope<RbacImportApplyResult>) => res.data,
      invalidatesTags: [{ type: "RbacSystem", id: "LIST" }],
    }),
    listRbacAuditEvents: builder.query<RbacAuditEvent[], RbacAuditFilter | void>({
      query: (filter) => {
        const params = new URLSearchParams();
        if (filter?.actor) params.set("actor", filter.actor);
        if (filter?.systemClientId) params.set("systemClientId", filter.systemClientId);
        if (filter?.roleName) params.set("roleName", filter.roleName);
        if (filter?.permissionKey) params.set("permissionKey", filter.permissionKey);
        if (filter?.action) params.set("action", filter.action);
        if (filter?.from) params.set("from", filter.from);
        if (filter?.to) params.set("to", filter.to);
        if (filter?.limit) params.set("limit", String(filter.limit));
        const query = params.toString();
        return `${base}/audit${query ? `?${query}` : ""}`;
      },
      transformResponse: (res: ApiEnvelope<RbacAuditEvent[]>) => res.data,
      providesTags: [{ type: "RbacAudit", id: "LIST" }],
    }),
    listRbacRoleTemplates: builder.query<RbacRoleTemplate[], void>({
      query: () => `${base}/role-templates`,
      transformResponse: (res: ApiEnvelope<RbacRoleTemplate[]>) => res.data,
    }),
    createRoleFromTemplate: builder.mutation<
      RbacSystemRole,
      { clientId: string; data: RbacRoleFromTemplatePayload }
    >({
      query: ({ clientId, data }) => ({
        url: `${base}/systems/${encodeURIComponent(clientId)}/roles/from-template`,
        method: "POST",
        body: data,
      }),
      transformResponse: (res: ApiEnvelope<RbacSystemRole>) => res.data,
      invalidatesTags: (_result, _error, { clientId }) => [{ type: "RbacSystem", id: clientId }],
    }),
    bulkAssignPermission: builder.mutation<void, RbacBulkPermissionPayload>({
      query: (data) => ({ url: `${base}/bulk/assign-permission`, method: "POST", body: data }),
      invalidatesTags: [{ type: "RbacSystem", id: "LIST" }],
    }),
    bulkRemovePermission: builder.mutation<void, RbacBulkPermissionPayload>({
      query: (data) => ({ url: `${base}/bulk/remove-permission`, method: "POST", body: data }),
      invalidatesTags: [{ type: "RbacSystem", id: "LIST" }],
    }),
    listAccessRequests: builder.query<RbacAccessRequest[], void>({
      query: () => `${base}/access-requests`,
      transformResponse: (res: ApiEnvelope<RbacAccessRequest[]>) => res.data,
      providesTags: [{ type: "RbacAccessRequest", id: "LIST" }],
    }),
    createAccessRequest: builder.mutation<RbacAccessRequest, RbacAccessRequestPayload>({
      query: (data) => ({ url: "/access-requests", method: "POST", body: data }),
      transformResponse: (res: ApiEnvelope<RbacAccessRequest>) => res.data,
      invalidatesTags: [{ type: "RbacAccessRequest", id: "LIST" }],
    }),
    decideAccessRequest: builder.mutation<
      RbacAccessRequest,
      { id: string; decision: "approved" | "rejected" | "cancelled"; data?: RbacDecisionPayload }
    >({
      query: ({ id, decision, data }) => ({
        url: `${base}/access-requests/${encodeURIComponent(id)}/${decision}`,
        method: "POST",
        body: data ?? {},
      }),
      transformResponse: (res: ApiEnvelope<RbacAccessRequest>) => res.data,
      invalidatesTags: [
        { type: "RbacAccessRequest", id: "LIST" },
        { type: "RbacAudit", id: "LIST" },
      ],
    }),
    listChangeRequests: builder.query<RbacChangeRequest[], void>({
      query: () => `${base}/change-requests`,
      transformResponse: (res: ApiEnvelope<RbacChangeRequest[]>) => res.data,
      providesTags: [{ type: "RbacChangeRequest", id: "LIST" }],
    }),
    createChangeRequest: builder.mutation<RbacChangeRequest, RbacChangeRequestPayload>({
      query: (data) => ({ url: `${base}/change-requests`, method: "POST", body: data }),
      transformResponse: (res: ApiEnvelope<RbacChangeRequest>) => res.data,
      invalidatesTags: [{ type: "RbacChangeRequest", id: "LIST" }],
    }),
    decideChangeRequest: builder.mutation<
      RbacChangeRequest,
      { id: string; decision: "approved" | "rejected" | "applied"; data?: RbacDecisionPayload }
    >({
      query: ({ id, decision, data }) => ({
        url: `${base}/change-requests/${encodeURIComponent(id)}/${decision}`,
        method: "POST",
        body: data ?? {},
      }),
      transformResponse: (res: ApiEnvelope<RbacChangeRequest>) => res.data,
      invalidatesTags: [
        { type: "RbacChangeRequest", id: "LIST" },
        { type: "RbacAudit", id: "LIST" },
      ],
    }),
    simulateRbacAccess: builder.mutation<RbacSimulationResult, RbacSimulationPayload>({
      query: (data) => ({ url: `${base}/simulate`, method: "POST", body: data }),
      transformResponse: (res: ApiEnvelope<RbacSimulationResult>) => res.data,
    }),
  }),
});

export const {
  useApplyRbacSyncMutation,
  useApplyRbacImportMutation,
  useAddSystemAccessRoleMutation,
  useAssignRealmRolePermissionMutation,
  useAssignSystemRolePermissionMutation,
  useBulkAssignPermissionMutation,
  useBulkRemovePermissionMutation,
  useCopyRolePermissionsMutation,
  useCreateAccessRequestMutation,
  useCreateChangeRequestMutation,
  useCreateRoleFromTemplateMutation,
  useCreateSystemRoleMutation,
  useDecideAccessRequestMutation,
  useDecideChangeRequestMutation,
  useDeleteSystemRoleMutation,
  useExportRbacSeedQuery,
  useGetRbacDriftQuery,
  useGetRbacEffectiveAccessQuery,
  useGetRealmRoleUsageQuery,
  useGetRoleUsageQuery,
  useListAccessRequestsQuery,
  useListChangeRequestsQuery,
  useListRbacAuditEventsQuery,
  useGetRbacSystemQuery,
  useListRbacPermissionsQuery,
  useListRbacRoleTemplatesQuery,
  useListRbacSystemsQuery,
  useListRealmRolePermissionsQuery,
  useListSystemRolesQuery,
  usePreviewRbacChangeMutation,
  usePreviewRbacImportMutation,
  usePreviewRbacRealmExportDriftMutation,
  usePreviewRbacSyncMutation,
  useRemoveRealmRolePermissionMutation,
  useRemoveSystemAccessRoleMutation,
  useRemoveSystemRolePermissionMutation,
  useSimulateRbacAccessMutation,
  useUpdateRbacPermissionMetadataMutation,
  useUpdateRbacSystemMutation,
  useUpdateSystemRoleMutation,
} = rbacApi;
