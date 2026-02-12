import { API } from "../../lib/constants/api.constants";
import type { GetNotificationsParams, Notification } from "../types/notifications.types";

import { baseApi } from "./baseApi";

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
} = notificationsApi;
