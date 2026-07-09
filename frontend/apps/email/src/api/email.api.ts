import type {
  EmailOutboxItem,
  EmailOutboxItemResponse,
  EmailListParams,
  EmailListByStatusParams,
  EmailRecipientPreviewRequest,
  EmailRecipientPreviewResponse,
  SendEmailRequest,
} from "../types";
import { baseApi } from "@moh-sso/api";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

function normalizeEmailOutboxItem(item: EmailOutboxItemResponse): EmailOutboxItem {
  return {
    id: item.id,
    tenant_id: item.tenant_id,
    message_id: item.message_id,
    message: item.message,
    status: item.status,
    attempts: item.attempts,
    max_attempts: item.max_attempts,
    last_error: item.last_error,
    scheduled_at: item.scheduled_at,
    locked_at: item.locked_at,
    sent_at: item.sent_at,
    created_at: item.created_at,
    updated_at: item.updated_at,
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

    previewEmailRecipients: builder.mutation<
      ApiEnvelope<EmailRecipientPreviewResponse>,
      EmailRecipientPreviewRequest
    >({
      query: (body) => ({
        url: "/emails/recipient-preview",
        method: "POST",
        body,
      }),
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
  usePreviewEmailRecipientsMutation,
  useListEmailsQuery,
  useGetEmailByIdQuery,
  useListEmailsByStatusQuery,
  useRetryEmailMutation,
  useDeleteEmailMutation,
} = emailApi;
