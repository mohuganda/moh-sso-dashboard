import type {
  EpiWeek,
  Disease,
  Subcounty,
  FacilityWeeklyMetric,
  DistrictWeeklyStatus,
  RegionWeeklyStatus,
  NationalWeeklyStatus,
} from "../types/surveillance.types";
import { baseApi } from "./baseApi";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

export const surveillanceApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    listEpiWeeks: builder.query<EpiWeek[], number | void>({
      query: (year) => ({
        url: "/surveillance/weeks",
        params: year !== undefined ? { year } : undefined,
      }),
      transformResponse: (res: ApiEnvelope<EpiWeek[]>) => res.data,
      providesTags: (result) =>
        result
          ? [
              ...result.map((week) => ({
                type: "Surveillance" as const,
                id: week.id,
              })),
              { type: "Surveillance" as const, id: "LIST" },
            ]
          : [{ type: "Surveillance" as const, id: "LIST" }],
    }),

    listDiseases: builder.query<Disease[], void>({
      query: () => ({
        url: "/surveillance/diseases",
      }),
      transformResponse: (res: ApiEnvelope<Disease[]>) => res.data,
      providesTags: (result) =>
        result
          ? [
              ...result.map((disease) => ({
                type: "Surveillance" as const,
                id: disease.id,
              })),
              { type: "Surveillance" as const, id: "LIST" },
            ]
          : [{ type: "Surveillance" as const, id: "LIST" }],
    }),

    listSubcountiesByDistrict: builder.query<Subcounty[], string>({
      query: (districtID) => ({
        url: `/surveillance/districts/${districtID}/subcounties`,
      }),
      transformResponse: (res: ApiEnvelope<Subcounty[]>) => res.data,
      providesTags: (result) =>
        result
          ? [
              ...result.map((subcounty) => ({
                type: "Surveillance" as const,
                id: subcounty.id,
              })),
              { type: "Surveillance" as const, id: "LIST" },
            ]
          : [{ type: "Surveillance" as const, id: "LIST" }],
    }),

    getSubcountyById: builder.query<Subcounty, string>({
      query: (id) => ({
        url: `/surveillance/subcounties/${id}`,
      }),
      transformResponse: (res: ApiEnvelope<Subcounty>) => res.data,
      providesTags: (_result, _error, id) => [{ type: "Surveillance" as const, id }],
    }),

    listFacilityWeeklyMetricsByWeek: builder.query<FacilityWeeklyMetric[], string>({
      query: (epiWeekID) => ({
        url: `/surveillance/facility-weekly-metrics/week/${epiWeekID}`,
      }),
      transformResponse: (res: ApiEnvelope<FacilityWeeklyMetric[]>) => res.data,
      providesTags: (result) =>
        result
          ? [
              ...result.map((item) => ({
                type: "Surveillance" as const,
                id: item.id,
              })),
              { type: "Surveillance" as const, id: "LIST" },
            ]
          : [{ type: "Surveillance" as const, id: "LIST" }],
    }),

    listFacilityWeeklyMetricsByFacility: builder.query<FacilityWeeklyMetric[], string>({
      query: (facilityID) => ({
        url: `/surveillance/facility-weekly-metrics/facility/${facilityID}`,
      }),
      transformResponse: (res: ApiEnvelope<FacilityWeeklyMetric[]>) => res.data,
      providesTags: (result) =>
        result
          ? [
              ...result.map((item) => ({
                type: "Surveillance" as const,
                id: item.id,
              })),
              { type: "Surveillance" as const, id: "LIST" },
            ]
          : [{ type: "Surveillance" as const, id: "LIST" }],
    }),

    listDistrictWeeklyStatusesByWeek: builder.query<DistrictWeeklyStatus[], string>({
      query: (epiWeekID) => ({
        url: `/surveillance/district-weekly-statuses/week/${epiWeekID}`,
      }),
      transformResponse: (res: ApiEnvelope<DistrictWeeklyStatus[]>) => res.data,
      providesTags: (result) =>
        result
          ? [
              ...result.map((item) => ({
                type: "Surveillance" as const,
                id: item.id,
              })),
              { type: "Surveillance" as const, id: "LIST" },
            ]
          : [{ type: "Surveillance" as const, id: "LIST" }],
    }),

    listDistrictWeeklyStatuses: builder.query<
      DistrictWeeklyStatus[],
      { districtID?: string; epiWeekID?: string } | void
    >({
      query: (params) => ({
        url: "/surveillance/district-weekly-statuses",
        params: params ?? undefined,
      }),
      transformResponse: (res: ApiEnvelope<DistrictWeeklyStatus[]>) => res.data,
      providesTags: (result) =>
        result
          ? [
              ...result.map((item) => ({
                type: "Surveillance" as const,
                id: item.id,
              })),
              { type: "Surveillance" as const, id: "LIST" },
            ]
          : [{ type: "Surveillance" as const, id: "LIST" }],
    }),

    listRegionWeeklyStatusesByWeek: builder.query<RegionWeeklyStatus[], string>({
      query: (epiWeekID) => ({
        url: `/surveillance/region-weekly-statuses/week/${epiWeekID}`,
      }),
      transformResponse: (res: ApiEnvelope<RegionWeeklyStatus[]>) => res.data,
      providesTags: (result) =>
        result
          ? [
              ...result.map((item) => ({
                type: "Surveillance" as const,
                id: item.id,
              })),
              { type: "Surveillance" as const, id: "LIST" },
            ]
          : [{ type: "Surveillance" as const, id: "LIST" }],
    }),

    listNationalWeeklyStatusesByWeek: builder.query<NationalWeeklyStatus[], string>({
      query: (epiWeekID) => ({
        url: `/surveillance/national-weekly-statuses/week/${epiWeekID}`,
      }),
      transformResponse: (res: ApiEnvelope<NationalWeeklyStatus[]>) => res.data,
      providesTags: (result) =>
        result
          ? [
              ...result.map((item) => ({
                type: "Surveillance" as const,
                id: item.id,
              })),
              { type: "Surveillance" as const, id: "LIST" },
            ]
          : [{ type: "Surveillance" as const, id: "LIST" }],
    }),
  }),
});

export const {
  useListEpiWeeksQuery,
  useListDiseasesQuery,
  useListSubcountiesByDistrictQuery,
  useGetSubcountyByIdQuery,
  useListFacilityWeeklyMetricsByWeekQuery,
  useListFacilityWeeklyMetricsByFacilityQuery,
  useListDistrictWeeklyStatusesByWeekQuery,
  useListDistrictWeeklyStatusesQuery,
  useListRegionWeeklyStatusesByWeekQuery,
  useListNationalWeeklyStatusesByWeekQuery,
} = surveillanceApi;
