import { baseApi } from "@moh-sso/api";

import type {
  HealthContextAlias,
  HealthContextAssignment,
  HealthContextAssignmentPayload,
  HealthContextDriftReport,
  HealthContextNode,
  HealthContextNodePayload,
  HealthContextSyncPreview,
  HealthContextSyncRequest,
  HealthContextSyncResult,
} from "../types";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

type AssignmentTarget = {
  targetId: string;
  assignments: HealthContextAssignmentPayload[];
};

const contextsBase = "/admin/health-contexts";

export const healthContextAdminApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    listHealthContexts: builder.query<
      HealthContextNode[],
      { includeDisabled?: boolean; type?: string } | void
    >({
      query: (filter) => {
        const params = new URLSearchParams();
        if (filter?.includeDisabled) params.set("includeDisabled", "true");
        if (filter?.type) params.set("type", filter.type);
        const query = params.toString();
        return `${contextsBase}${query ? `?${query}` : ""}`;
      },
      transformResponse: (response: ApiEnvelope<HealthContextNode[]>) => response.data,
      providesTags: [{ type: "HealthContext", id: "LIST" }],
    }),
    createHealthContext: builder.mutation<HealthContextNode, HealthContextNodePayload>({
      query: (body) => ({ url: contextsBase, method: "POST", body }),
      transformResponse: (response: ApiEnvelope<HealthContextNode>) => response.data,
      invalidatesTags: [{ type: "HealthContext", id: "LIST" }],
    }),
    updateHealthContext: builder.mutation<
      HealthContextNode,
      { contextId: string; body: HealthContextNodePayload & { version: number } }
    >({
      query: ({ contextId, body }) => ({
        url: `${contextsBase}/${encodeURIComponent(contextId)}`,
        method: "PUT",
        body,
      }),
      transformResponse: (response: ApiEnvelope<HealthContextNode>) => response.data,
      invalidatesTags: (_result, _error, { contextId }) => [
        { type: "HealthContext", id: "LIST" },
        { type: "HealthContext", id: contextId },
      ],
    }),
    deleteHealthContext: builder.mutation<void, string>({
      query: (contextId) => ({
        url: `${contextsBase}/${encodeURIComponent(contextId)}`,
        method: "DELETE",
      }),
      invalidatesTags: [{ type: "HealthContext", id: "LIST" }],
    }),
    listHealthContextAliases: builder.query<HealthContextAlias[], string>({
      query: (contextId) => `${contextsBase}/${encodeURIComponent(contextId)}/aliases`,
      transformResponse: (response: ApiEnvelope<HealthContextAlias[]>) => response.data,
      providesTags: (_result, _error, contextId) => [
        { type: "HealthContext", id: `${contextId}:aliases` },
      ],
    }),
    upsertHealthContextAlias: builder.mutation<
      HealthContextAlias,
      { contextId: string; namespace: string; externalId: string }
    >({
      query: ({ contextId, ...body }) => ({
        url: `${contextsBase}/${encodeURIComponent(contextId)}/aliases`,
        method: "PUT",
        body,
      }),
      transformResponse: (response: ApiEnvelope<HealthContextAlias>) => response.data,
      invalidatesTags: (_result, _error, { contextId }) => [
        { type: "HealthContext", id: `${contextId}:aliases` },
      ],
    }),
    deleteHealthContextAlias: builder.mutation<
      void,
      { contextId: string; aliasId: string }
    >({
      query: ({ contextId, aliasId }) => ({
        url: `${contextsBase}/${encodeURIComponent(contextId)}/aliases/${encodeURIComponent(aliasId)}`,
        method: "DELETE",
      }),
      invalidatesTags: (_result, _error, { contextId }) => [
        { type: "HealthContext", id: `${contextId}:aliases` },
      ],
    }),
    listUserHealthContexts: builder.query<HealthContextAssignment[], string>({
      query: (userId) => `/admin/users/${encodeURIComponent(userId)}/health-contexts`,
      transformResponse: (response: ApiEnvelope<HealthContextAssignment[]>) => response.data,
      providesTags: (_result, _error, userId) => [
        { type: "HealthContext", id: `user:${userId}` },
      ],
    }),
    replaceUserHealthContexts: builder.mutation<void, AssignmentTarget>({
      query: ({ targetId, assignments }) => ({
        url: `/admin/users/${encodeURIComponent(targetId)}/health-contexts`,
        method: "PUT",
        body: { assignments },
      }),
      invalidatesTags: (_result, _error, { targetId }) => [
        { type: "HealthContext", id: `user:${targetId}` },
      ],
    }),
    listGroupHealthContexts: builder.query<HealthContextAssignment[], string>({
      query: (groupId) => `/admin/rbac/groups/${encodeURIComponent(groupId)}/health-contexts`,
      transformResponse: (response: ApiEnvelope<HealthContextAssignment[]>) => response.data,
      providesTags: (_result, _error, groupId) => [
        { type: "HealthContext", id: `group:${groupId}` },
      ],
    }),
    replaceGroupHealthContexts: builder.mutation<void, AssignmentTarget>({
      query: ({ targetId, assignments }) => ({
        url: `/admin/rbac/groups/${encodeURIComponent(targetId)}/health-contexts`,
        method: "PUT",
        body: { assignments },
      }),
      invalidatesTags: (_result, _error, { targetId }) => [
        { type: "HealthContext", id: `group:${targetId}` },
      ],
    }),
    getHealthContextDrift: builder.query<HealthContextDriftReport, void>({
      query: () => `${contextsBase}/drift`,
      transformResponse: (response: ApiEnvelope<HealthContextDriftReport>) => response.data,
      providesTags: [{ type: "HealthContext", id: "DRIFT" }],
    }),
    previewHealthContextSync: builder.mutation<
      HealthContextSyncPreview,
      HealthContextSyncRequest
    >({
      query: (body) => ({
        url: `${contextsBase}/sync/preview`,
        method: "POST",
        body,
      }),
      transformResponse: (response: ApiEnvelope<HealthContextSyncPreview>) => response.data,
    }),
    applyHealthContextSync: builder.mutation<
      HealthContextSyncResult,
      HealthContextSyncRequest
    >({
      query: (body) => ({
        url: `${contextsBase}/sync/apply`,
        method: "POST",
        body,
      }),
      transformResponse: (response: ApiEnvelope<HealthContextSyncResult>) => response.data,
      invalidatesTags: [
        { type: "HealthContext", id: "DRIFT" },
        { type: "HealthContext", id: "LIST" },
      ],
    }),
  }),
});

export const {
  useApplyHealthContextSyncMutation,
  useCreateHealthContextMutation,
  useDeleteHealthContextAliasMutation,
  useDeleteHealthContextMutation,
  useListGroupHealthContextsQuery,
  useGetHealthContextDriftQuery,
  useListHealthContextAliasesQuery,
  useListHealthContextsQuery,
  useListUserHealthContextsQuery,
  useReplaceGroupHealthContextsMutation,
  useReplaceUserHealthContextsMutation,
  usePreviewHealthContextSyncMutation,
  useUpdateHealthContextMutation,
  useUpsertHealthContextAliasMutation,
} = healthContextAdminApi;
