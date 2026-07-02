import { API } from "@moh-sso/config";
import type {
  Issue,
  IssuePayload,
  IssueTransaction,
  IssueTransactionPayload,
} from "../types";

import { baseApi } from "@moh-sso/api";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

export const issuesApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getIssues: builder.query<Issue[], void>({
      query: () => ({
        url: API.issue.list(),
        method: "GET",
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<Issue[]>) => res.data,
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
  }),
});

export const {
  useGetIssuesQuery,
  useCreateIssueMutation,
  useUpdateIssueMutation,
  useGetTransactionsQuery,
  useCreateTransactionMutation,
} = issuesApi;
