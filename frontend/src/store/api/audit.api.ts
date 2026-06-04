import { API } from "@/lib/constants/api.constants";
import type { AuditFilters, AuditListResponse } from "@/store/types/audit.types";

import { baseApi } from "./baseApi";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

export const auditApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    listAuditLogs: builder.query<AuditListResponse, AuditFilters>({
      query: (params) => ({
        url: API.admin.audit.list(),
        params,
        credentials: "include",
      }),

      transformResponse: (response: ApiEnvelope<AuditListResponse>) => response.data,

      providesTags: ["Audit"],
    }),
  }),
});

export const { useListAuditLogsQuery } = auditApi;
