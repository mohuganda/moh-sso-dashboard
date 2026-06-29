import type {
  EmailOutboxItem,
  EmailOutboxItemResponse,
  EmailListParams,
  EmailListByStatusParams,
  SendEmailRequest,
} from "@moh-sso/types";
import { baseApi } from "@moh-sso/api";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

function normalizeEmailOutboxItem(item: EmailOutboxItemResponse): EmailOutboxItem {
  return {
    id: item.ID,
    tenant_id: item.TenantID,
    message_id: item.MessageID,
    message: item.Message,
    status: item.Status,
    attempts: item.Attempts,
    max_attempts: item.MaxAttempts,
    last_error: item.LastError,
    scheduled_at: item.ScheduledAt,
    locked_at: item.LockedAt,
    sent_at: item.SentAt,
    created_at: item.CreatedAt,
    updated_at: item.UpdatedAt,
  };
}

function normalizeEmailOutboxItems(items: EmailOutboxItemResponse[]): EmailOutboxItem[] {
  return items.map(normalizeEmailOutboxItem);
}

export const emailApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    sendEmail: builder.mutation<ApiEnvelope<{ message: string }>, SendEmailRequest>({
      query: (body) => ({
        url: "/emails/send",
        method: "POST",
        body,
      }),
      invalidatesTags: ["Emails"],
    }),

    queueEmail: builder.mutation<ApiEnvelope<{ message: string }>, SendEmailRequest>({
      query: (body) => ({
        url: "/emails/queue",
        method: "POST",
        body,
      }),
      invalidatesTags: ["Emails"],
    }),

    listEmails: builder.query<EmailOutboxItem[], EmailListParams | void>({
      query: (params) => ({
        url: "/emails",
        method: "GET",
        params: {
          limit: params?.limit ?? 20,
          offset: params?.offset ?? 0,
        },
      }),
      transformResponse: (response: ApiEnvelope<EmailOutboxItemResponse[]>) =>
        normalizeEmailOutboxItems(response.data ?? []),
      providesTags: ["Emails"],
    }),

    getEmailById: builder.query<EmailOutboxItem, string>({
      query: (id) => ({
        url: `/emails/${id}`,
        method: "GET",
      }),
      transformResponse: (response: ApiEnvelope<EmailOutboxItemResponse>) =>
        normalizeEmailOutboxItem(response.data),
      providesTags: (_result, _error, id) => [{ type: "Emails", id }],
    }),

    listEmailsByStatus: builder.query<EmailOutboxItem[], EmailListByStatusParams>({
      query: ({ status, limit = 20, offset = 0 }) => ({
        url: `/emails/status/${status}`,
        method: "GET",
        params: {
          limit,
          offset,
        },
      }),
      transformResponse: (response: ApiEnvelope<EmailOutboxItemResponse[]>) =>
        normalizeEmailOutboxItems(response.data ?? []),
      providesTags: ["Emails"],
    }),

    retryEmail: builder.mutation<ApiEnvelope<{ message: string }>, string>({
      query: (id) => ({
        url: `/emails/${id}/retry`,
        method: "POST",
      }),
      invalidatesTags: (_result, _error, id) => ["Emails", { type: "Emails", id }],
    }),

    deleteEmail: builder.mutation<ApiEnvelope<{ message: string }>, string>({
      query: (id) => ({
        url: `/emails/${id}`,
        method: "DELETE",
      }),
      invalidatesTags: ["Emails"],
    }),
  }),
});

export const {
  useSendEmailMutation,
  useQueueEmailMutation,
  useListEmailsQuery,
  useGetEmailByIdQuery,
  useListEmailsByStatusQuery,
  useRetryEmailMutation,
  useDeleteEmailMutation,
} = emailApi;
