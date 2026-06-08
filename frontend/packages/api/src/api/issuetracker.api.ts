import { baseApi } from "@moh-sso/api";
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

export const issuesApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    /* --------------------------------
     * GET /issues
     * -------------------------------- */
    getIssues: builder.query<IssueResponse, void>({
      query: () => ({
        url: API.issue.list(),
        method: "GET",
        credentials: "include",
      }),
      providesTags: ["Issues"],
    }),

    /* --------------------------------
     * POST /issues
     * -------------------------------- */
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

    /* --------------------------------
     * PUT /issues/:id
     * -------------------------------- */
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

    /* --------------------------------
     * GET /issues/:id/transactions
     * -------------------------------- */
    getTransactions: builder.query<IssueTransactionsResponse, string | number>({
      query: (id) => ({
        url: `${API.issue.list()}/${id}/transactions`,
        method: "GET",
        credentials: "include",
      }),
      providesTags: ["Transactions", "Issues"],
    }),

    /* --------------------------------
     * POST /issues/:id/resolveIssue
     * -------------------------------- */
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
