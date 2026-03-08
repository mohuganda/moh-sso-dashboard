import { API } from "../../lib/constants/api.constants";
import {
  loginSuccess,
  logout as logoutAction,
  authLoaded,
  setAccessToken,
} from "../auth/auth.slice";
import type { AuthUser } from "../auth/auth.types";

import { baseApi } from "./baseApi";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

type RefreshResponse = ApiEnvelope<{
  access_token: string;
}>;

export const authApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    /* -----------------------------
     * Get current user
     * ----------------------------- */
    me: builder.query<AuthUser, void>({
      query: () => ({
        url: API.auth.me(),
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<{ user: AuthUser }>) => res.data.user,
    }),

    /* -----------------------------
     * Update profile
     * ----------------------------- */
    updateProfile: builder.mutation<AuthUser, { firstName?: string; lastName?: string }>({
      query: (body) => ({
        url: API.auth.me(),
        method: "PUT",
        body,
        credentials: "include",
      }),
    }),

    /* -----------------------------
     * Refresh session
     * ----------------------------- */
    refresh: builder.mutation<RefreshResponse, void>({
      query: () => ({
        url: API.auth.refresh(),
        method: "POST",
        credentials: "include",
      }),

      async onQueryStarted(_, { dispatch, queryFulfilled }) {
        try {
          const { data } = await queryFulfilled;

          dispatch(setAccessToken(data.data.access_token));

          const user = await dispatch(
            authApi.endpoints.me.initiate(undefined, {
              forceRefetch: true,
            }),
          ).unwrap();

          dispatch(
            loginSuccess({
              accessToken: data.data.access_token,
              user,
            }),
          );
        } catch {
          dispatch(logoutAction());
        } finally {
          dispatch(authLoaded());
        }
      },
    }),
  }),
});

export const { useMeQuery, useUpdateProfileMutation, useRefreshMutation } = authApi;
