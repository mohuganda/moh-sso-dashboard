import { baseApi } from "./baseApi";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

type GeoJsonObject = GeoJSON.FeatureCollection | GeoJSON.Feature;

export const geojsonApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getGeoJson: builder.query<GeoJsonObject, "water" | "districts" | "subcounties">({
      query: (name) => ({
        url: `/geojson/${name}`,
        method: "GET",
      }),
      transformResponse: (response: ApiEnvelope<GeoJsonObject> | GeoJsonObject) =>
        "data" in response ? response.data : response,
      providesTags: (_result, _error, name) => [{ type: "GeoJson", id: name }],
    }),
  }),
  overrideExisting: false,
});

export const { useGetGeoJsonQuery, useLazyGetGeoJsonQuery } = geojsonApi;
