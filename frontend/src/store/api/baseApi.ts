import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";

import type { RootState } from "..";
import { API } from "../../lib/constants/api.constants";
import { selectAccessToken } from "../auth/auth.selectors";

export const baseApi = createApi({
  reducerPath: "api",
  baseQuery: fetchBaseQuery({
    baseUrl: API.base,
    credentials: "include",
    prepareHeaders: (headers, { getState }) => {
      const token = selectAccessToken(getState() as RootState);

      if (token) {
        headers.set("Authorization", `Bearer ${token}`);
      }

      return headers;
    },
  }),
  tagTypes: [
    "Auth",
    "User",
    "Client",
    "Audit",
    "Metrics",
    "Notification",
    "ClientRole",
    "UserClientRole",
  ],
  endpoints: () => ({}),
});
