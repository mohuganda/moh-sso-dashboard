import { API } from "@moh-sso/config";
import type {
  GetNotificationsParams,
  Notification,
  NotificationDelivery,
  NotificationDeliveryList,
  NotificationDeliveryListParams,
  NotificationDeliveryMetrics,
  NotificationPreferences,
  TestSMSRequest,
  TestSMSResponse,
  UpdateNotificationPreferencesRequest,
} from "@moh-sso/types";

import { baseApi } from "@moh-sso/api";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

export const notificationsApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    /* --------------------------------
     * List notifications
     * -------------------------------- */
    getNotifications: builder.query<Notification[], GetNotificationsParams>({
      query: ({ unread, limit = 20, offset = 0 }) => {
        const params = new URLSearchParams();

        if (unread !== undefined) {
          params.set("unread", String(unread));
        }

        params.set("limit", String(limit));
        params.set("offset", String(offset));

        return {
          url: `${API.admin.notifications.list()}?${params.toString()}`,
          credentials: "include",
        };
      },

      transformResponse: (res: ApiEnvelope<Notification[]>) => res.data,

      providesTags: (result) =>
        result
          ? [
              ...result.map(({ id }) => ({
                type: "Notification" as const,
                id,
              })),
              { type: "Notification", id: "LIST" },
            ]
          : [{ type: "Notification", id: "LIST" }],
    }),

    /* --------------------------------
     * Mark as read
     * -------------------------------- */
    markNotificationAsRead: builder.mutation<void, string>({
      query: (id) => ({
        url: API.admin.notifications.markAsRead(id),
        method: "PATCH",
        credentials: "include",
      }),

      transformResponse: () => undefined,

      invalidatesTags: (_r, _e, id) => [
        { type: "Notification", id },
        { type: "Notification", id: "LIST" },
        { type: "Notification", id: "COUNT" },
      ],
    }),

    /* --------------------------------
     * Delete notification
     * -------------------------------- */
    deleteNotification: builder.mutation<void, string>({
      query: (id) => ({
        url: API.admin.notifications.delete(id),
        credentials: "include",
        method: "DELETE",
      }),

      transformResponse: () => undefined,

      invalidatesTags: ["Notification"],
    }),

    /* --------------------------------
     * Delete old notifications
     * -------------------------------- */
    deleteOldNotifications: builder.mutation<void, void>({
      query: () => ({
        url: API.admin.notifications.deleteOld(),
        credentials: "include",
        method: "DELETE",
      }),

      transformResponse: () => undefined,

      invalidatesTags: ["Notification"],
    }),

    /* --------------------------------
     * Count notifications
     * -------------------------------- */
    countNotifications: builder.query<number, void>({
      query: () => ({
        url: API.admin.notifications.count(),
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<{ count: number }>) => res.data.count,

      providesTags: ["Notification"],
    }),

    /* --------------------------------
     * Create notification (internal)
     * -------------------------------- */
    notify: builder.mutation<void, void>({
      query: () => ({
        url: API.admin.notifications.notify(),
        credentials: "include",
        method: "POST",
      }),

      transformResponse: () => undefined,

      invalidatesTags: ["Notification"],
    }),

    /* --------------------------------
     * Get single notification
     * -------------------------------- */
    getNotification: builder.query<Notification, string>({
      query: (id) => ({
        url: API.admin.notifications.byId(id),
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<Notification>) => res.data,

      providesTags: (_r, _e, id) => [{ type: "Notification", id }],
    }),

    /* --------------------------------
     * Count unread notifications
     * -------------------------------- */
    getUnreadNotificationsCount: builder.query<number, void>({
      query: () => ({
        url: API.admin.notifications.countUnread(),
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<{ count: number }>) => res.data.count,

      providesTags: [{ type: "Notification", id: "COUNT" }],
    }),

    getNotificationDeliveries: builder.query<NotificationDelivery[], string>({
      query: (notificationId) => ({
        url: API.admin.notifications.deliveries(notificationId),
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<NotificationDelivery[]>) => res.data,

      providesTags: (_result, _error, notificationId) => [
        { type: "Notification", id: `DELIVERIES-${notificationId}` },
      ],
    }),

    getAllNotificationDeliveries: builder.query<
      NotificationDeliveryList,
      NotificationDeliveryListParams | void
    >({
      query: (filters) => {
        const params = new URLSearchParams();
        if (filters?.channel) {
          params.set("channel", filters.channel);
        }
        if (filters?.status) {
          params.set("status", filters.status);
        }
        if (filters?.search) {
          params.set("search", filters.search);
        }
        params.set("limit", String(filters?.limit ?? 20));
        params.set("offset", String(filters?.offset ?? 0));

        return {
          url: `${API.admin.notifications.deliveryList()}?${params.toString()}`,
          credentials: "include",
        };
      },

      transformResponse: (res: ApiEnvelope<NotificationDeliveryList>) => res.data,

      providesTags: [{ type: "Notification", id: "DELIVERY-LIST" }],
    }),

    getNotificationDeliveryMetrics: builder.query<NotificationDeliveryMetrics, void>({
      query: () => ({
        url: API.admin.notifications.deliveryMetrics(),
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<NotificationDeliveryMetrics>) => res.data,

      providesTags: [{ type: "Notification", id: "DELIVERY-METRICS" }],
    }),

    getNotificationDelivery: builder.query<NotificationDelivery, string>({
      query: (deliveryId) => ({
        url: API.admin.notifications.deliveryById(deliveryId),
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<NotificationDelivery>) => res.data,

      providesTags: (_result, _error, deliveryId) => [
        { type: "Notification", id: `DELIVERY-${deliveryId}` },
      ],
    }),

    retryNotificationDelivery: builder.mutation<void, { deliveryId: string; notificationId: string }>({
      query: ({ deliveryId }) => ({
        url: API.admin.notifications.retryDelivery(deliveryId),
        method: "POST",
        credentials: "include",
      }),

      transformResponse: () => undefined,

      invalidatesTags: (_result, _error, { notificationId }) => [
        { type: "Notification", id: `DELIVERIES-${notificationId}` },
        { type: "Notification", id: "DELIVERY-LIST" },
        { type: "Notification", id: "DELIVERY-METRICS" },
        { type: "Notification", id: "LIST" },
      ],
    }),

    cancelNotificationDelivery: builder.mutation<
      void,
      { deliveryId: string; notificationId?: string }
    >({
      query: ({ deliveryId }) => ({
        url: API.admin.notifications.cancelDelivery(deliveryId),
        method: "POST",
        credentials: "include",
      }),

      transformResponse: () => undefined,

      invalidatesTags: (_result, _error, { deliveryId, notificationId }) => [
        { type: "Notification", id: `DELIVERY-${deliveryId}` },
        { type: "Notification", id: "DELIVERY-LIST" },
        { type: "Notification", id: "DELIVERY-METRICS" },
        ...(notificationId
          ? [{ type: "Notification" as const, id: `DELIVERIES-${notificationId}` }]
          : []),
      ],
    }),

    sendTestSMS: builder.mutation<TestSMSResponse, TestSMSRequest>({
      query: (body) => ({
        url: API.admin.notifications.testSms(),
        method: "POST",
        body,
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<TestSMSResponse>) => res.data,

      invalidatesTags: [
        { type: "Notification", id: "DELIVERY-LIST" },
        { type: "Notification", id: "DELIVERY-METRICS" },
        { type: "Notification", id: "LIST" },
      ],
    }),

    getNotificationPreferences: builder.query<NotificationPreferences, void>({
      query: () => ({
        url: API.notifications.preferences(),
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<NotificationPreferences>) => res.data,

      providesTags: [{ type: "Notification", id: "PREFERENCES" }],
    }),

    updateNotificationPreferences: builder.mutation<
      NotificationPreferences,
      UpdateNotificationPreferencesRequest
    >({
      query: (body) => ({
        url: API.notifications.preferences(),
        method: "PUT",
        body,
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<NotificationPreferences>) => res.data,

      invalidatesTags: [{ type: "Notification", id: "PREFERENCES" }],
    }),
  }),
});

export const {
  useGetNotificationsQuery,
  useMarkNotificationAsReadMutation,
  useDeleteNotificationMutation,
  useDeleteOldNotificationsMutation,
  useCountNotificationsQuery,
  useNotifyMutation,
  useGetNotificationQuery,
  useGetUnreadNotificationsCountQuery,
  useGetNotificationDeliveriesQuery,
  useGetAllNotificationDeliveriesQuery,
  useGetNotificationDeliveryQuery,
  useGetNotificationDeliveryMetricsQuery,
  useRetryNotificationDeliveryMutation,
  useCancelNotificationDeliveryMutation,
  useSendTestSMSMutation,
  useGetNotificationPreferencesQuery,
  useUpdateNotificationPreferencesMutation,
} = notificationsApi;
