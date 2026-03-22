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
        params: year ? { year } : undefined,
      }),
      transformResponse: (res: ApiEnvelope<EpiWeek[]>) => res.data,
      providesTags: (result) =>
        result
          ? [
              ...result.map((client) => ({
                type: "Surveillance" as const,
                id: client.id,
              })),
              { type: "Surveillance", id: "LIST" },
            ]
          : [{ type: "Surveillance", id: "LIST" }],
    }),

    listDiseases: builder.query<Disease[], void>({
      query: () => "/surveillance/diseases",
      providesTags: ["Surveillance"],
    }),

    listSubcountiesByDistrict: builder.query<Subcounty[], string>({
      query: (districtID) => `/surveillance/districts/${districtID}/subcounties`,
      transformResponse: (res: ApiEnvelope<Subcounty[]>) => res.data,
      providesTags: ["Surveillance"],
    }),

    getSubcountyById: builder.query<Subcounty, string>({
      query: (id) => `/surveillance/subcounties/${id}`,
      transformResponse: (res: ApiEnvelope<Subcounty[]>) => res.data,
      providesTags: ["Surveillance"],
    }),

    listFacilityWeeklyMetricsByWeek: builder.query<FacilityWeeklyMetric[], string>({
      query: (epiWeekID) => `/surveillance/facility-weekly-metrics/week/${epiWeekID}`,
      transformResponse: (res: ApiEnvelope<FacilityWeeklyMetric[]>) => res.data,
      providesTags: ["Surveillance"],
    }),

    listFacilityWeeklyMetricsByFacility: builder.query<FacilityWeeklyMetric[], string>({
      query: (facilityID) => `/surveillance/facility-weekly-metrics/facility/${facilityID}`,
      transformResponse: (res: ApiEnvelope<FacilityWeeklyMetric[]>) => res.data,
      providesTags: ["Surveillance"],
    }),

    listDistrictWeeklyStatusesByWeek: builder.query<DistrictWeeklyStatus[], string>({
      query: (epiWeekID) => `/surveillance/district-weekly-statuses/week/${epiWeekID}`,
      transformResponse: (res: ApiEnvelope<DistrictWeeklyStatus[]>) => res.data,
      providesTags: ["Surveillance"],
    }),

    listDistrictWeeklyStatuses: builder.query<
      DistrictWeeklyStatus[],
      { districtID?: string; epiWeekID?: string } | void
    >({
      query: (params) => ({
        url: "/surveillance/district-weekly-statuses",
        params,
      }),
      transformResponse: (res: ApiEnvelope<DistrictWeeklyStatus[]>) => res.data,
      providesTags: ["Surveillance"],
    }),

    listRegionWeeklyStatusesByWeek: builder.query<RegionWeeklyStatus[], string>({
      query: (epiWeekID) => `/surveillance/region-weekly-statuses/week/${epiWeekID}`,
      transformResponse: (res: ApiEnvelope<RegionWeeklyStatus[]>) => res.data,
      providesTags: ["Surveillance"],
    }),

    listNationalWeeklyStatusesByWeek: builder.query<NationalWeeklyStatus[], string>({
      query: (epiWeekID) => `/surveillance/national-weekly-statuses/week/${epiWeekID}`,
      transformResponse: (res: ApiEnvelope<NationalWeeklyStatus[]>) => res.data,
      providesTags: ["Surveillance"],
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
