import { baseApi } from "@moh-sso/api";
import { API } from "@moh-sso/config";

export type Dataset = {
  dataset_id: string;
  display_name: string;
};

export type ThemeElement = {
  theme_category_id: number;
  theme_id: string;
  theme_name: string;
  data_element_key: number;
  data_element_id: string;
  data_element_short_name: string;
};

export const datasetApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    /* --------------------------------
     * List Visualizer Datasets
     * GET /visualizer/Datasets
     * -------------------------------- */
    getDataSets: builder.query<Dataset[], void>({
      query: () => ({
        url: API.visualizer.datasets(),
        method: "GET",
        credentials: "include",
      }),
    }),
    getDataSetElements: builder.query<ThemeElement[], string | number>({
      query: (id) => ({
        url: `${API.visualizer.dataElements()}?data_set_id=${id}`,
        method: "GET",
        credentials: "include",
      }),
      providesTags: ["Datasets"],
    }),
  }),
});

export const { useGetDataSetsQuery, useLazyGetDataSetElementsQuery } = datasetApi;
