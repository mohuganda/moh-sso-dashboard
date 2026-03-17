import type {
  Announcement,
  AnnouncementStats,
  CreateAnnouncementRequest,
  ListAnnouncementsParams,
  UpdateAnnouncementRequest,
  ScheduleAnnouncementRequest,
  SetAnnouncementPinnedRequest,
  SetAnnouncementPriorityRequest,
} from "../types/announcements.types";
import { baseApi } from "./baseApi";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

export const announcementApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    listAnnouncementsAdmin: builder.query<Announcement[], ListAnnouncementsParams | void>({
      query: (params) => ({
        url: "/admin/announcements",
        params: {
          limit: params?.limit ?? 20,
          offset: params?.offset ?? 0,
        },
      }),
      transformResponse: (response: ApiEnvelope<Announcement[]>) => response.data,
      providesTags: (result) =>
        result
          ? [
              ...result.map((item) => ({
                type: "Announcements" as const,
                id: item.id,
              })),
              { type: "Announcements" as const, id: "LIST" },
            ]
          : [{ type: "Announcements" as const, id: "LIST" }],
    }),

    getAnnouncementById: builder.query<Announcement, string>({
      query: (id) => ({
        url: `/admin/announcements/${id}`,
      }),
      transformResponse: (response: ApiEnvelope<Announcement>) => response.data,
      providesTags: (_result, _error, id) => [{ type: "Announcements" as const, id }],
    }),

    getAnnouncementStats: builder.query<AnnouncementStats, void>({
      query: () => ({
        url: "/admin/announcements/stats",
      }),
      transformResponse: (response: ApiEnvelope<AnnouncementStats>) => response.data,
      providesTags: [{ type: "Announcements", id: "STATS" }],
    }),

    createAnnouncement: builder.mutation<Announcement, CreateAnnouncementRequest>({
      query: (body) => ({
        url: "/admin/announcements",
        method: "POST",
        body,
      }),
      transformResponse: (response: ApiEnvelope<Announcement>) => response.data,
      invalidatesTags: [
        { type: "Announcements", id: "LIST" },
        { type: "Announcements", id: "STATS" },
        { type: "Announcements", id: "MY_LIST" },
      ],
    }),

    updateAnnouncement: builder.mutation<
      Announcement,
      { id: string; body: UpdateAnnouncementRequest }
    >({
      query: ({ id, body }) => ({
        url: `/admin/announcements/${id}`,
        method: "PUT",
        body,
      }),
      transformResponse: (response: ApiEnvelope<Announcement>) => response.data,
      invalidatesTags: (_result, _error, { id }) => [
        { type: "Announcements", id },
        { type: "Announcements", id: "LIST" },
        { type: "Announcements", id: "STATS" },
        { type: "Announcements", id: "MY_LIST" },
      ],
    }),

    deleteAnnouncement: builder.mutation<{ message: string }, string>({
      query: (id) => ({
        url: `/admin/announcements/${id}`,
        method: "DELETE",
      }),
      transformResponse: (response: ApiEnvelope<{ message: string }>) => response.data,
      invalidatesTags: (_result, _error, id) => [
        { type: "Announcements", id },
        { type: "Announcements", id: "LIST" },
        { type: "Announcements", id: "STATS" },
        { type: "Announcements", id: "MY_LIST" },
      ],
    }),

    restoreAnnouncement: builder.mutation<Announcement, string>({
      query: (id) => ({
        url: `/admin/announcements/${id}/restore`,
        method: "POST",
      }),
      transformResponse: (response: ApiEnvelope<Announcement>) => response.data,
      invalidatesTags: (_result, _error, id) => [
        { type: "Announcements", id },
        { type: "Announcements", id: "LIST" },
        { type: "Announcements", id: "STATS" },
      ],
    }),

    publishAnnouncement: builder.mutation<Announcement, string>({
      query: (id) => ({
        url: `/admin/announcements/${id}/publish`,
        method: "POST",
      }),
      transformResponse: (response: ApiEnvelope<Announcement>) => response.data,
      invalidatesTags: (_result, _error, id) => [
        { type: "Announcements", id },
        { type: "Announcements", id: "LIST" },
        { type: "Announcements", id: "STATS" },
        { type: "Announcements", id: "MY_LIST" },
      ],
    }),

    scheduleAnnouncement: builder.mutation<
      Announcement,
      { id: string; body: ScheduleAnnouncementRequest }
    >({
      query: ({ id, body }) => ({
        url: `/admin/announcements/${id}/schedule`,
        method: "POST",
        body,
      }),
      transformResponse: (response: ApiEnvelope<Announcement>) => response.data,
      invalidatesTags: (_result, _error, { id }) => [
        { type: "Announcements", id },
        { type: "Announcements", id: "LIST" },
        { type: "Announcements", id: "STATS" },
      ],
    }),

    archiveAnnouncement: builder.mutation<Announcement, string>({
      query: (id) => ({
        url: `/admin/announcements/${id}/archive`,
        method: "POST",
      }),
      transformResponse: (response: ApiEnvelope<Announcement>) => response.data,
      invalidatesTags: (_result, _error, id) => [
        { type: "Announcements", id },
        { type: "Announcements", id: "LIST" },
        { type: "Announcements", id: "STATS" },
        { type: "Announcements", id: "MY_LIST" },
      ],
    }),

    setAnnouncementPinned: builder.mutation<
      Announcement,
      { id: string; body: SetAnnouncementPinnedRequest }
    >({
      query: ({ id, body }) => ({
        url: `/admin/announcements/${id}/pin`,
        method: "PATCH",
        body,
      }),
      transformResponse: (response: ApiEnvelope<Announcement>) => response.data,
      invalidatesTags: (_result, _error, { id }) => [
        { type: "Announcements", id },
        { type: "Announcements", id: "LIST" },
        { type: "Announcements", id: "MY_LIST" },
      ],
    }),

    setAnnouncementPriority: builder.mutation<
      Announcement,
      { id: string; body: SetAnnouncementPriorityRequest }
    >({
      query: ({ id, body }) => ({
        url: `/admin/announcements/${id}/priority`,
        method: "PATCH",
        body,
      }),
      transformResponse: (response: ApiEnvelope<Announcement>) => response.data,
      invalidatesTags: (_result, _error, { id }) => [
        { type: "Announcements", id },
        { type: "Announcements", id: "LIST" },
        { type: "Announcements", id: "MY_LIST" },
      ],
    }),

    listMyAnnouncements: builder.query<Announcement[], ListAnnouncementsParams | void>({
      query: (params) => ({
        url: "/announcements/me",
        params: {
          limit: params?.limit ?? 20,
          offset: params?.offset ?? 0,
          role: params?.role,
          client_id: params?.clientId,
        },
      }),
      transformResponse: (response: ApiEnvelope<Announcement[]>) => response.data,
      providesTags: (result) =>
        result
          ? [
              ...result.map((item) => ({
                type: "Announcements" as const,
                id: item.id,
              })),
              { type: "Announcements" as const, id: "MY_LIST" },
            ]
          : [{ type: "Announcements" as const, id: "MY_LIST" }],
    }),
  }),
  overrideExisting: false,
});

export const {
  useListAnnouncementsAdminQuery,
  useGetAnnouncementByIdQuery,
  useGetAnnouncementStatsQuery,
  useCreateAnnouncementMutation,
  useUpdateAnnouncementMutation,
  useDeleteAnnouncementMutation,
  useRestoreAnnouncementMutation,
  usePublishAnnouncementMutation,
  useScheduleAnnouncementMutation,
  useArchiveAnnouncementMutation,
  useSetAnnouncementPinnedMutation,
  useSetAnnouncementPriorityMutation,
  useListMyAnnouncementsQuery,
} = announcementApi;
