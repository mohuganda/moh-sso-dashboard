import { baseApi } from "@moh-sso/api";
import { API } from "@moh-sso/config";

export type DataItem = {
  rows: row[];
};

export type row = {
  category_combo?: string;
  data_element_id: string;
  dataelement: string;
  district: string;
  facility: string;
  level: string;
  org_unit_id: string;
  period: string;
  region: string;
  sub_county: string;
  value: number;
};

export const dataValuesApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getDataValues: builder.query<any, void>({
      query: (bodyParams) => ({
        url: API.visualizer.dataValues(),
        method: "POST",
        body: bodyParams,
        credentials: "include",
      }),
      keepUnusedDataFor: 0,
      transformResponse: (response: any) => response?.data,
    }),
  }),
});

export const { useLazyGetDataValuesQuery } = dataValuesApi;
