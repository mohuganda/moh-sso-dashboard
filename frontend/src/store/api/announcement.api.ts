import type {
  Announcement,
  ListAnnouncementsParams,
  CreateAnnouncementRequest,
} from "../types/announcements.types";
import { baseApi } from "./baseApi";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

export const announcementApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    listAnnouncements: builder.query<Announcement[], ListAnnouncementsParams | void>({
      query: (params) => ({
        url: "/announcements",
        params: {
          limit: params?.limit ?? 20,
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

    createAnnouncement: builder.mutation<Announcement, CreateAnnouncementRequest>({
      query: (body) => ({
        url: "/announcements",
        method: "POST",
        body,
      }),
      transformResponse: (response: ApiEnvelope<Announcement>) => response.data,
      invalidatesTags: [{ type: "Announcements", id: "LIST" }],
    }),

    deleteAnnouncement: builder.mutation<void, string>({
      query: (id) => ({
        url: `/announcements/${id}`,
        method: "DELETE",
      }),
      invalidatesTags: (_result, _error, id) => [
        { type: "Announcements", id },
        { type: "Announcements", id: "LIST" },
      ],
    }),
  }),
  overrideExisting: false,
});

export const {
  useListAnnouncementsQuery,
  useCreateAnnouncementMutation,
  useDeleteAnnouncementMutation,
} = announcementApi;
