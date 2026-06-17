import { baseApi } from "@moh-sso/api";
import { API } from "@moh-sso/config";

export type Theme = {
  theme_id: string;
  theme_name: string;
};

export type ThemeElement = {
  theme_category_id: number;
  theme_id: string;
  theme_name: string;
  data_element_key: number;
  data_element_id: string;
  data_element_short_name: string;
};

export const themesApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    /* --------------------------------
     * List Visualizer Themes
     * GET /visualizer/themes
     * -------------------------------- */
    getThemes: builder.query<Theme[], void>({
      query: () => ({
        url: API.visualizer.themes(),
        method: "GET",
        credentials: "include",
      }),
    }),
    getThemeElements: builder.query<ThemeElement[], string>({
      query: (id) => ({
        url: API.visualizer.theme(),
        method: "POST",
        body: { theme_id: id },
        credentials: "include",
      }),
      providesTags: (result) =>
        result
          ? [
              ...result.map(({ theme_id }) => ({ type: "ThemeElement" as const, theme_id })),
              { type: "ThemeElement", theme_id: "LIST" }, // The "General" tag
            ]
          : [{ type: "ThemeElement", theme_id: "LIST" }],
      keepUnusedDataFor: 0,
    }),
  }),
});

export const { useGetThemesQuery, useLazyGetThemeElementsQuery } = themesApi;
