import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { selectAccessToken } from "../auth/auth.selectors";
import type { RootState } from "..";
import { API } from "../../lib/constants/api.constants";
import { logout } from "../auth/auth.slice";

/**
 * Base query with JWT + future-safe 401 handling
 */
const rawBaseQuery = fetchBaseQuery({
  baseUrl: API.base,
  credentials: "include",
  prepareHeaders: (headers, { getState }) => {
    const token = selectAccessToken(getState() as RootState);

    if (token) {
      headers.set("Authorization", `Bearer ${token}`);
    }

    return headers;
  },
});

/**
 * Optional: 401 auto-logout handler
 */
const baseQueryWithReauth: typeof rawBaseQuery = async (args, api, extraOptions) => {
  const result = await rawBaseQuery(args, api, extraOptions);

  if (result.error?.status === 401) {
    // Optional: auto logout if token expired
    api.dispatch(logout());
  }

  return result;
};

export const baseApi = createApi({
  reducerPath: "api",
  baseQuery: baseQueryWithReauth,

  /**
   * 🔥 Add Document tag types here
   */
  tagTypes: [
    "Auth",
    "User",
    "Client",
    "Audit",
    "Metrics",
    "Notification",
    "ClientRole",
    "UserClientRole",
    "Document",
    "Documents",
    "DocumentProcesses",
    "StorageLocations",
    "StorageLocation",
    "UserSession",
    "UserSessions",
    "Announcements",
    "ThemeElement",
    "Surveillance",
    "GeoJson",
  ],

  endpoints: () => ({}),
});
