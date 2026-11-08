import { baseApi } from "@moh-sso/api";
import { API } from "@moh-sso/config";

import type {
  DQACompilePreview,
  DQAFlag,
  DQAPhysicalColumn,
  DQARule,
  DQARuleInput,
  DQARuleMappingCandidate,
  DQARuleMappingResult,
  DQARunResult,
  DQARunScope,
  DQARunSummary,
  DQAScheduledRun,
  DQATableMapping,
} from "../dqa.types";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

export const dqaApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    listDQATables: builder.query<DQATableMapping[], void>({
      query: () => ({ url: API.dataValidation.dqa.tables.list(), method: "GET", credentials: "include" }),
      transformResponse: (res: ApiEnvelope<DQATableMapping[]>) => res.data,
      providesTags: ["DQATables"],
    }),

    upsertDQATable: builder.mutation<DQATableMapping, DQATableMapping>({
      query: (body) => ({ url: API.dataValidation.dqa.tables.upsert(), method: "POST", body, credentials: "include" }),
      transformResponse: (res: ApiEnvelope<DQATableMapping>) => res.data,
      invalidatesTags: ["DQATables"],
    }),

    listDQAPhysicalColumns: builder.query<DQAPhysicalColumn[], string>({
      query: (table) => ({
        url: API.dataValidation.dqa.columns(table),
        method: "GET",
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<DQAPhysicalColumn[]>) => res.data,
    }),

    deleteDQATable: builder.mutation<void, string>({
      query: (tableId) => ({ url: API.dataValidation.dqa.tables.remove(tableId), method: "DELETE", credentials: "include" }),
      invalidatesTags: ["DQATables"],
    }),

    listDQARules: builder.query<DQARule[], { tableId?: string } | void>({
      query: (args) => ({ url: API.dataValidation.dqa.rules.list(args?.tableId), method: "GET", credentials: "include" }),
      transformResponse: (res: ApiEnvelope<DQARule[]>) => res.data,
      providesTags: ["DQARules"],
    }),

    upsertDQARule: builder.mutation<DQARule, DQARuleInput>({
      query: (body) => ({ url: API.dataValidation.dqa.rules.upsert(), method: "POST", body, credentials: "include" }),
      transformResponse: (res: ApiEnvelope<DQARule>) => res.data,
      invalidatesTags: ["DQARules"],
    }),

    compileDQARule: builder.mutation<DQACompilePreview, DQARuleInput>({
      query: (body) => ({ url: API.dataValidation.dqa.rules.compile(), method: "POST", body, credentials: "include" }),
      transformResponse: (res: ApiEnvelope<DQACompilePreview>) => res.data,
    }),

    scheduleDQARun: builder.mutation<
      DQAScheduledRun,
      { table_id: string; scope: DQARunScope; scheduled_at: string }
    >({
      query: (body) => ({
        url: API.dataValidation.dqa.schedules.create(),
        method: "POST",
        body,
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<DQAScheduledRun>) => res.data,
      invalidatesTags: ["DQASchedules"],
    }),

    listDQAScheduledRuns: builder.query<DQAScheduledRun[], { status?: string } | void>({
      query: (args) => ({
        url: API.dataValidation.dqa.schedules.list(args?.status),
        method: "GET",
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<DQAScheduledRun[]>) => res.data,
      providesTags: ["DQASchedules"],
    }),

    cancelDQAScheduledRun: builder.mutation<DQAScheduledRun, number>({
      query: (scheduleId) => ({
        url: API.dataValidation.dqa.schedules.cancel(scheduleId),
        method: "DELETE",
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<DQAScheduledRun>) => res.data,
      invalidatesTags: ["DQASchedules"],
    }),

    previewRuleMapping: builder.mutation<DQARuleMappingCandidate[], { table_id: string }>({
      query: (body) => ({
        url: API.dataValidation.dqa.rules.mapping.preview(),
        method: "POST",
        body,
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<DQARuleMappingCandidate[]>) => res.data,
    }),

    applyRuleMapping: builder.mutation<DQARuleMappingResult, { table_id: string; codes: string[] }>({
      query: (body) => ({
        url: API.dataValidation.dqa.rules.mapping.apply(),
        method: "POST",
        body,
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<DQARuleMappingResult>) => res.data,
      invalidatesTags: ["DQARules"],
    }),

    deleteDQARule: builder.mutation<void, { tableId: string; code: string }>({
      query: ({ tableId, code }) => ({
        url: API.dataValidation.dqa.rules.remove(tableId, code),
        method: "DELETE",
        credentials: "include",
      }),
      invalidatesTags: ["DQARules"],
    }),

    runDQATable: builder.mutation<DQARunResult, { tableId: string; scope?: DQARunScope }>({
      query: ({ tableId, scope }) => ({
        url: API.dataValidation.dqa.run(),
        method: "POST",
        body: { table_id: tableId, scope },
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<DQARunResult>) => res.data,
      invalidatesTags: ["DQARuns"],
    }),

    listDQARuns: builder.query<DQARunSummary[], { tableId?: string } | void>({
      query: (args) => ({ url: API.dataValidation.dqa.runs.list(args?.tableId), method: "GET", credentials: "include" }),
      transformResponse: (res: ApiEnvelope<DQARunSummary[]>) => res.data,
      providesTags: ["DQARuns"],
    }),

    listDQAFlags: builder.query<DQAFlag[], { runId: number; severity?: string }>({
      query: ({ runId, severity }) => ({
        url: API.dataValidation.dqa.runs.flags(runId, severity),
        method: "GET",
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<DQAFlag[]>) => res.data,
      providesTags: (_result, _error, args) => [{ type: "DQAFlags", id: args.runId }],
    }),
  }),
});

export const {
  useListDQATablesQuery,
  useUpsertDQATableMutation,
  useDeleteDQATableMutation,
  useLazyListDQAPhysicalColumnsQuery,
  useListDQARulesQuery,
  useUpsertDQARuleMutation,
  useCompileDQARuleMutation,
  usePreviewRuleMappingMutation,
  useApplyRuleMappingMutation,
  useDeleteDQARuleMutation,
  useRunDQATableMutation,
  useScheduleDQARunMutation,
  useListDQAScheduledRunsQuery,
  useCancelDQAScheduledRunMutation,
  useListDQARunsQuery,
  useListDQAFlagsQuery,
} = dqaApi;
