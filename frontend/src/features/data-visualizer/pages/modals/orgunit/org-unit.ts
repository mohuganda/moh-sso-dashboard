import { baseApi } from "@/store/api/baseApi.ts";
import { API } from "@/lib/constants/api.constants.ts";

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
    }),
  }),
});

export const { useGetHierarchyQuery } = hierarchyApi;
