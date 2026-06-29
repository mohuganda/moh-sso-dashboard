import { API } from "@moh-sso/config";
import type { AuditActionsResponse, AuditFilters, AuditListResponse } from "../types";

import { baseApi } from "@moh-sso/api";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

export function buildAuditLogParams(filters: AuditFilters, extra?: Record<string, string>) {
  const params = new URLSearchParams();
  const append = (key: string, value: unknown) => {
    if (value === undefined || value === null || value === "") {
      return;
    }
    params.set(key, String(value));
  };

  append("from", filters.from);
  append("to", filters.to);
  append("action", filters.action);
  append("user_id", filters.user_id);
  append("client_id", filters.client_id);
  append("ip", filters.ip);
  append("success", filters.success);
  append("limit", filters.limit);
  append("cursor_id", filters.cursor_id);
  append("cursor_created_at", filters.cursor_created_at);

  Object.entries(extra ?? {}).forEach(([key, value]) => append(key, value));

  return params;
}

export function buildAuditExportUrl(filters: AuditFilters, format: "csv" | "json") {
  const params = buildAuditLogParams(filters, { format });
  params.delete("limit");
  params.delete("cursor_id");
  params.delete("cursor_created_at");

  return `${API.admin.audit.export()}?${params.toString()}`;
}

export const auditApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    listAuditLogs: builder.query<AuditListResponse, AuditFilters>({
      query: (params) => ({
        url: API.admin.audit.list(),
        params: Object.fromEntries(buildAuditLogParams(params)),
        credentials: "include",
      }),

      transformResponse: (response: ApiEnvelope<AuditListResponse>) => response.data,

      providesTags: ["Audit"],
    }),
    listAuditActions: builder.query<AuditActionsResponse, void>({
      query: () => ({
        url: API.admin.audit.actions(),
        credentials: "include",
      }),

      transformResponse: (response: ApiEnvelope<AuditActionsResponse>) => response.data,

      providesTags: ["Audit"],
    }),
  }),
});

export const { useListAuditActionsQuery, useListAuditLogsQuery } = auditApi;
