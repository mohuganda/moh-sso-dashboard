import {baseApi} from "../../../../store/api/baseApi.ts";
import {API} from "../../../../lib/constants/api.constants.ts";

export const issuesApi = baseApi.injectEndpoints({
    endpoints: (builder) => ({
        /* --------------------------------
         * GET /issues
         * -------------------------------- */
        getIssues: builder.query<any, void>({
            query: () => ({
                url: API.issue.list(),
                method: "GET",
                credentials: "include",
            }),
            providesTags: ['Issues'],
        }),
        createIssue: builder.mutation<any, void>({
            query: (body) => ({
                url: API.issue.list(),
                method: "POST",
                body,
                credentials: "include",
            }),
            transformResponse: (res: any) => res.data,
            invalidatesTags: ["Issues"],
        }),
        updateIssue: builder.mutation<any, { id: string | number; body: any }>({
            query: ({ id, body }) => ({
                url: `${API.issue.list()}/${id}`,
                method: "PUT",
                body,
                credentials: "include",
            }),
            invalidatesTags: ["Issues"],
        }),
        getTransactions: builder.query<any, string>({
            query: (id) => ({
                url: `${API.issue.list()}/${id}/transactions`,
                method: "GET",
                credentials: "include",
            }),
            providesTags: ['Transactions', "Issues"],
        }),
        createTransaction: builder.mutation<any, { id: string | number; body: any }>({
            query: ({ id, body }) => ({
                url: `${API.issue.list()}/${id}/resolveIssue`,
                method: "POST",
                body,
                credentials: "include",
            }),
            transformResponse: (res: any) => res.data,
            invalidatesTags: ["Transactions", "Issues"],
        }),
    }),
});

export const {
    useGetIssuesQuery,
    useCreateIssueMutation,
    useUpdateIssueMutation,
    useGetTransactionsQuery,
    useCreateTransactionMutation
} = issuesApi;

