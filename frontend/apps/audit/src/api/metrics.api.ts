import { API } from "@moh-sso/config";
import type {
  AuditFailedLoginsByDayResponse,
  AuditMetricsFilters,
  AuditOverview,
  AuditTopFailureIpsFilters,
  AuditTopFailureIpsResponse,
} from "@moh-sso/types";

import { baseApi } from "@moh-sso/api";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

export const metricsApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    auditOverview: builder.query<AuditOverview, AuditMetricsFilters>({
      query: ({ from, to }) => ({
        url: API.admin.audit.metrics.overview(),
        params: { from, to },
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<AuditOverview>) => res.data,

      providesTags: ["Audit"],
    }),

    auditFailedLoginsByDay: builder.query<
      AuditFailedLoginsByDayResponse,
      AuditMetricsFilters
    >({
      query: ({ from, to }) => ({
        url: API.admin.audit.metrics.failedLoginsByDay(),
        params: { from, to },
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<AuditFailedLoginsByDayResponse>) => res.data,

      providesTags: ["Audit"],
    }),

    auditTopFailureIps: builder.query<
      AuditTopFailureIpsResponse,
      AuditTopFailureIpsFilters
    >({
      query: ({ from, to, limit = 10 }) => ({
        url: API.admin.audit.metrics.topFailureIps(),
        params: { from, to, limit },
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<AuditTopFailureIpsResponse>) => res.data,

      providesTags: ["Audit"],
    }),
  }),
});

export const {
  useAuditFailedLoginsByDayQuery,
  useAuditOverviewQuery,
  useAuditTopFailureIpsQuery,
} = metricsApi;
