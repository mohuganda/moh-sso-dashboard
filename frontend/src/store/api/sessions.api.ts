import type { UserSession } from "../types/sessions.types";

import { baseApi } from "./baseApi";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

export const sessionsApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getSessions: builder.query<UserSession[], void>({
      query: () => ({
        url: "/sessions",
      }),
      transformResponse: (res: ApiEnvelope<UserSession[]>) => res.data,
      providesTags: (result) =>
        result
          ? [
              ...result.map((client) => ({
                type: "UserSession" as const,
                id: client.id,
              })),
              { type: "UserSessions", id: "LIST" },
            ]
          : [{ type: "UserSessions", id: "LIST" }],
    }),

    logoutSession: builder.mutation<void, string>({
      query: (sessionId) => ({
        url: `/sessions/${sessionId}`,
        method: "DELETE",
      }),
      transformResponse: () => undefined,
      invalidatesTags: (_r, _e, id) => [
        { type: "UserSession", id },
        { type: "UserSessions", id: "LIST" },
      ],
    }),
  }),
});

export const { useGetSessionsQuery, useLogoutSessionMutation } = sessionsApi;
