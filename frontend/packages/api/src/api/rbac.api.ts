import type {
  RbacAccessRolePayload,
  RbacPermission,
  RbacPermissionPayload,
  RbacRealmRolePermissionGroup,
  RbacRolePayload,
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
  }),
});

export const {
  useAddSystemAccessRoleMutation,
  useAssignRealmRolePermissionMutation,
  useAssignSystemRolePermissionMutation,
  useCreateSystemRoleMutation,
  useDeleteSystemRoleMutation,
  useGetRbacSystemQuery,
  useListRbacPermissionsQuery,
  useListRbacSystemsQuery,
  useListRealmRolePermissionsQuery,
  useListSystemRolesQuery,
  useRemoveRealmRolePermissionMutation,
  useRemoveSystemAccessRoleMutation,
  useRemoveSystemRolePermissionMutation,
  useUpdateRbacSystemMutation,
  useUpdateSystemRoleMutation,
} = rbacApi;
