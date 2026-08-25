import { API } from "@moh-sso/config";
import type {
  Issue,
  IssuePayload,
  IssueTransaction,
  IssueTransactionPayload,
  IssueProgramSummary,
  GetIssuesParams,
  GetIssuesResponse,
  KeycloakGroup,
  KeycloakGroupMember,
  AssignIssuesPayload,
} from "../types";

import { baseApi } from "@moh-sso/api";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

export const issuesApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getIssues: builder.query<GetIssuesResponse, GetIssuesParams | void>({
      query: (params) => {
        let url = API.issue.list();
        const queryParams = new URLSearchParams();
        if (params) {
          if (params.limit !== undefined) queryParams.append("limit", String(params.limit));
          if (params.offset !== undefined) queryParams.append("offset", String(params.offset));
          if (params.program) queryParams.append("program", params.program);
        }
        const queryString = queryParams.toString();
        if (queryString) {
          url += `?${queryString}`;
        }
        return {
          url,
          method: "GET",
          credentials: "include",
        };
      },
      transformResponse: (res: ApiEnvelope<Issue[]>, meta: any) => {
        const totalHeader = meta?.response?.headers?.get("X-Total-Count");
        const totalCount = totalHeader ? parseInt(totalHeader, 10) : (res.data?.length ?? 0);
        return {
          items: res.data ?? [],
          totalCount,
        };
      },
      providesTags: ["Issues"],
    }),

    getIssuesSummaryByProgram: builder.query<IssueProgramSummary[], void>({
      query: () => ({
        url: `${API.issue.list()}/summary-by-program`,
        method: "GET",
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<IssueProgramSummary[]>) => res.data,
      providesTags: ["Issues"],
    }),

    createIssue: builder.mutation<Issue, IssuePayload>({
      query: (body) => ({
        url: API.issue.list(),
        method: "POST",
        body,
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<Issue>) => res.data,
      invalidatesTags: ["Issues"],
    }),

    updateIssue: builder.mutation<
      Issue,
      {
        id: string | number;
        body: Partial<IssuePayload>;
      }
    >({
      query: ({ id, body }) => ({
        url: `${API.issue.list()}/${id}`,
        method: "PUT",
        body,
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<Issue>) => res.data,
      invalidatesTags: ["Issues"],
    }),

    getTransactions: builder.query<IssueTransaction[], string | number>({
      query: (id) => ({
        url: `${API.issue.list()}/${id}/transactions`,
        method: "GET",
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<IssueTransaction[]>) => res.data,
      providesTags: ["Transactions", "Issues"],
    }),

    createTransaction: builder.mutation<
      IssueTransaction,
      {
        id: string | number;
        body: IssueTransactionPayload;
      }
    >({
      query: ({ id, body }) => ({
        url: `${API.issue.list()}/${id}/resolveIssue`,
        method: "POST",
        body,
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<IssueTransaction>) => res.data,
      invalidatesTags: ["Transactions", "Issues"],
    }),

    getKeycloakGroups: builder.query<KeycloakGroup[], void>({
      query: () => ({
        url: `${API.issue.list()}/keycloak-groups`,
        method: "GET",
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<KeycloakGroup[]>) => res.data ?? [],
    }),

    getKeycloakGroupMembers: builder.query<KeycloakGroupMember[], string>({
      query: (groupId) => ({
        url: `${API.issue.list()}/keycloak-groups/${groupId}/members`,
        method: "GET",
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<KeycloakGroupMember[]>) => res.data ?? [],
    }),

    getIssueByCode: builder.query<Issue, string>({
      query: (issueCode) => ({
        url: `${API.issue.list()}/${issueCode}`,
        method: "GET",
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<Issue>) => res.data,
      providesTags: ["Issues"],
    }),

    assignIssues: builder.mutation<IssueTransaction[], AssignIssuesPayload>({
      query: (body) => ({
        url: `${API.issue.list()}/assign`,
        method: "POST",
        body,
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<IssueTransaction[]>) => res.data,
      invalidatesTags: ["Issues", "Transactions"],
    }),
  }),
});

export const {
  useGetIssuesQuery,
  useLazyGetIssuesQuery,
  useGetIssueByCodeQuery,
  useLazyGetIssueByCodeQuery,
  useGetIssuesSummaryByProgramQuery,
  useCreateIssueMutation,
  useUpdateIssueMutation,
  useGetTransactionsQuery,
  useCreateTransactionMutation,
  useGetKeycloakGroupsQuery,
  useGetKeycloakGroupMembersQuery,
  useAssignIssuesMutation,
} = issuesApi;
