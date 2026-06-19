import { API } from "@moh-sso/config";
import type {
  Issue,
  IssuePayload,
  IssueResponse,
  IssueTransaction,
  IssueTransactionPayload,
  IssueTransactionsResponse,
  SingleIssueResponse,
} from "@moh-sso/types";

import { baseApi } from "./baseApi";

export const issuesApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getIssues: builder.query<IssueResponse, void>({
      query: () => ({
        url: API.issue.list(),
        method: "GET",
        credentials: "include",
      }),
      providesTags: ["Issues"],
    }),

    createIssue: builder.mutation<Issue, IssuePayload>({
      query: (body) => ({
        url: API.issue.list(),
        method: "POST",
        body,
        credentials: "include",
      }),
      transformResponse: (res: SingleIssueResponse) => res.data,
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
      transformResponse: (res: SingleIssueResponse) => res.data,
      invalidatesTags: ["Issues"],
    }),

    getTransactions: builder.query<IssueTransactionsResponse, string | number>({
      query: (id) => ({
        url: `${API.issue.list()}/${id}/transactions`,
        method: "GET",
        credentials: "include",
      }),
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
      transformResponse: (res: { data: IssueTransaction }) => res.data,
      invalidatesTags: ["Transactions", "Issues"],
    }),
  }),
});

export const {
  useGetIssuesQuery,
  useCreateIssueMutation,
  useUpdateIssueMutation,
  useGetTransactionsQuery,
  useCreateTransactionMutation,
} = issuesApi;
