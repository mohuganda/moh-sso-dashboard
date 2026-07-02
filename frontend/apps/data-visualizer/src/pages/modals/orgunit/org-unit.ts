import { baseApi } from "@moh-sso/api";
import { API } from "@moh-sso/config";

export const hierarchyApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    /* --------------------------------
     * List Visualizer Themes
     * GET /visualizer/themes
     * -------------------------------- */
    getHierarchy: builder.query<any, void>({
      query: () => ({
        url: API.visualizer.hierarchy(),
        method: "GET",
        credentials: "include",
      }),
      transformResponse: (response: any) => response?.data,
    }),
  }),
});

export const { useGetHierarchyQuery } = hierarchyApi;
