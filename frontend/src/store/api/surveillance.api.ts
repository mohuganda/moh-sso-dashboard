import type {
  EpiWeek,
  Disease,
  Subcounty,
  FacilityWeeklyMetric,
  District,
  Region,
  CreateImportBatchPayload,
  SurveillanceImportBatch,
  SurveillanceImportRawRow,
  UpdateImportBatchStatusPayload,
  UploadSurveillanceCsvRequest,
  Alert,
  WeeklyStatus,
  ListWeeklyStatusesParams,
  WeeklyStatusDetailed,
  ListAlertsParams,
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
              { type: "Surveillance" as const, id: "SURVEILLANCE_WEEKS_LIST" },
            ]
          : [{ type: "Surveillance" as const, id: "SURVEILLANCE_WEEKS_LIST" }],
    }),

    listAlerts: builder.query<Alert[], ListAlertsParams | void>({
      query: (params: ListAlertsParams) => ({
        url: "/surveillance/alerts",
        method: "GET",
        params,
      }),
      transformResponse: (res: ApiEnvelope<Alert[]>) => res.data,
      providesTags: (result) =>
        result
          ? [
              ...result.map((alert) => ({
                type: "Surveillance" as const,
                id: alert.id,
              })),
              { type: "Surveillance" as const, id: "SURVEILLANCE_ALERTS_LIST" },
            ]
          : [{ type: "Surveillance" as const, id: "SURVEILLANCE_ALERTS_LIST" }],
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
              { type: "Surveillance" as const, id: "SURVEILLANCE_DISEASES_LIST" },
            ]
          : [{ type: "Surveillance" as const, id: "SURVEILLANCE_DISEASES_LIST" }],
    }),

    listRegions: builder.query<Region[], void>({
      query: () => ({
        url: "/surveillance/regions",
      }),
      transformResponse: (res: ApiEnvelope<Region[]>) => res.data,
      providesTags: (result) =>
        result
          ? [
              ...result.map((region) => ({
                type: "Surveillance" as const,
                id: region.id,
              })),
              { type: "Surveillance" as const, id: "SURVEILLANCE_REGIONS_LIST" },
            ]
          : [{ type: "Surveillance" as const, id: "SURVEILLANCE_REGIONS_LIST" }],
    }),

    listDistricts: builder.query<District[], void>({
      query: () => ({
        url: "/surveillance/districts",
      }),
      transformResponse: (res: ApiEnvelope<District[]>) => res.data,
      providesTags: (result) =>
        result
          ? [
              ...result.map((district) => ({
                type: "Surveillance" as const,
                id: district.id,
              })),
              { type: "Surveillance" as const, id: "SURVEILLANCE_DISTRICTS_LIST" },
            ]
          : [{ type: "Surveillance" as const, id: "SURVEILLANCE_DISTRICTS_LIST" }],
    }),

    listDistrictsByRegion: builder.query<District[], string>({
      query: (regionID) => ({
        url: `/surveillance/regions/${regionID}/districts`,
      }),
      transformResponse: (res: ApiEnvelope<District[]>) => res.data,
      providesTags: (result, _error, regionID) =>
        result
          ? [
              ...result.map((district) => ({
                type: "Surveillance" as const,
                id: district.id,
              })),
              { type: "Surveillance" as const, id: `SURVEILLANCE_REGION_DISTRICTS_${regionID}` },
            ]
          : [{ type: "Surveillance" as const, id: `SURVEILLANCE_REGION_DISTRICTS_${regionID}` }],
    }),

    listSubcountiesByDistrict: builder.query<Subcounty[], string>({
      query: (districtID) => ({
        url: `/surveillance/districts/${districtID}/subcounties`,
      }),
      transformResponse: (res: ApiEnvelope<Subcounty[]>) => res.data,
      providesTags: (result, _error, districtID) =>
        result
          ? [
              ...result.map((subcounty) => ({
                type: "Surveillance" as const,
                id: subcounty.id,
              })),
              {
                type: "Surveillance" as const,
                id: `SURVEILLANCE_DISTRICT_SUBCOUNTIES_${districtID}`,
              },
            ]
          : [
              {
                type: "Surveillance" as const,
                id: `SURVEILLANCE_DISTRICT_SUBCOUNTIES_${districtID}`,
              },
            ],
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
      providesTags: (result, _error, epiWeekID) =>
        result
          ? [
              ...result.map((item) => ({
                type: "Surveillance" as const,
                id: item.id,
              })),
              {
                type: "Surveillance" as const,
                id: `SURVEILLANCE_FACILITY_METRICS_WEEK_${epiWeekID}`,
              },
            ]
          : [
              {
                type: "Surveillance" as const,
                id: `SURVEILLANCE_FACILITY_METRICS_WEEK_${epiWeekID}`,
              },
            ],
    }),

    listFacilityWeeklyMetricsByFacility: builder.query<FacilityWeeklyMetric[], string>({
      query: (facilityID) => ({
        url: `/surveillance/facility-weekly-metrics/facility/${facilityID}`,
      }),
      transformResponse: (res: ApiEnvelope<FacilityWeeklyMetric[]>) => res.data,
      providesTags: (result, _error, facilityID) =>
        result
          ? [
              ...result.map((item) => ({
                type: "Surveillance" as const,
                id: item.id,
              })),
              {
                type: "Surveillance" as const,
                id: `SURVEILLANCE_FACILITY_METRICS_FACILITY_${facilityID}`,
              },
            ]
          : [
              {
                type: "Surveillance" as const,
                id: `SURVEILLANCE_FACILITY_METRICS_FACILITY_${facilityID}`,
              },
            ],
    }),

    listWeeklyStatuses: builder.query<WeeklyStatus[], ListWeeklyStatusesParams | void>({
      query: (params: ListWeeklyStatusesParams) => ({
        url: "/surveillance/weekly-statuses/list",
        method: "GET",
        params,
      }),
      transformResponse: (res: ApiEnvelope<WeeklyStatus[]>) => res.data,
      providesTags: (result) =>
        result
          ? [
              ...result.map((item) => ({
                type: "Surveillance" as const,
                id: item.id,
              })),
              {
                type: "Surveillance" as const,
                id: "WEEKLY_STATUSES",
              },
            ]
          : [
              {
                type: "Surveillance" as const,
                id: "WEEKLY_STATUSES",
              },
            ],
    }),

    listWeeklyStatusesDetailed: builder.query<
      WeeklyStatusDetailed[],
      ListWeeklyStatusesParams | void
    >({
      query: (params: ListWeeklyStatusesParams) => ({
        url: "/surveillance/weekly-statuses/detailed",
        method: "GET",
        params,
      }),
      transformResponse: (res: ApiEnvelope<WeeklyStatusDetailed[]>) => res.data,
      providesTags: (result) =>
        result
          ? [
              ...result.map((item) => ({
                type: "Surveillance" as const,
                id: item.id,
              })),
              {
                type: "Surveillance" as const,
                id: "WEEKLY_STATUSES_DETAILED",
              },
            ]
          : [
              {
                type: "Surveillance" as const,
                id: "WEEKLY_STATUSES_DETAILED",
              },
            ],
    }),
    listImportBatches: builder.query<SurveillanceImportBatch[], void>({
      query: () => ({
        url: `/surveillance/imports`,
      }),
      transformResponse: (res: ApiEnvelope<SurveillanceImportBatch[]>) => res.data,
      providesTags: (result) =>
        result
          ? [
              ...result.map((item) => ({
                type: "Surveillance" as const,
                id: item.id,
              })),
              {
                type: "Surveillance" as const,
                id: "SURVEILLANCE_IMPORT_BATCHES",
              },
            ]
          : [
              {
                type: "Surveillance" as const,
                id: "SURVEILLANCE_IMPORT_BATCHES",
              },
            ],
    }),

    getImportBatchById: builder.query<SurveillanceImportBatch, string>({
      query: (batchID) => ({
        url: `/surveillance/imports/${batchID}`,
      }),
      transformResponse: (res: ApiEnvelope<SurveillanceImportBatch>) => res.data,
      providesTags: (_result, _error, batchID) => [{ type: "Surveillance" as const, id: batchID }],
    }),

    listImportRawRowsByBatch: builder.query<SurveillanceImportRawRow[], string>({
      query: (batchID) => ({
        url: `/surveillance/imports/${batchID}/raw-rows`,
      }),
      transformResponse: (res: ApiEnvelope<SurveillanceImportRawRow[]>) => res.data,
      providesTags: (result, _error, batchID) =>
        result
          ? [
              ...result.map((item) => ({
                type: "Surveillance" as const,
                id: item.id,
              })),
              {
                type: "Surveillance" as const,
                id: `SURVEILLANCE_IMPORT_RAW_ROWS_${batchID}`,
              },
            ]
          : [
              {
                type: "Surveillance" as const,
                id: `SURVEILLANCE_IMPORT_RAW_ROWS_${batchID}`,
              },
            ],
    }),

    createImportBatch: builder.mutation<SurveillanceImportBatch, CreateImportBatchPayload>({
      query: (body) => ({
        url: `/surveillance/imports`,
        method: "POST",
        body,
      }),
      transformResponse: (res: ApiEnvelope<SurveillanceImportBatch>) => res.data,
      invalidatesTags: [{ type: "Surveillance" as const, id: "SURVEILLANCE_IMPORT_BATCHES" }],
    }),

    updateImportBatchStatus: builder.mutation<
      SurveillanceImportBatch,
      UpdateImportBatchStatusPayload
    >({
      query: ({ batchID, ...body }) => ({
        url: `/surveillance/imports/${batchID}/status`,
        method: "PATCH",
        body,
      }),
      transformResponse: (res: ApiEnvelope<SurveillanceImportBatch>) => res.data,
      invalidatesTags: (_result, _error, { batchID }) => [
        { type: "Surveillance" as const, id: batchID },
        { type: "Surveillance" as const, id: "SURVEILLANCE_IMPORT_BATCHES" },
      ],
    }),

    // upload
    uploadSurveillanceCsv: builder.mutation<{ message?: string }, UploadSurveillanceCsvRequest>({
      query: ({ file, storage_location }) => {
        const formData = new FormData();
        formData.append("file", file);
        if (storage_location) {
          formData.append("storage_location", storage_location);
        }

        return {
          url: "/surveillance/import",
          method: "POST",
          body: formData,
        };
      },
      transformResponse: (res: ApiEnvelope<{ message?: string }>) => res.data,
      invalidatesTags: [
        { type: "Surveillance", id: "SURVEILLANCE_EPI_WEEKS" },
        { type: "Surveillance", id: "SURVEILLANCE_DISEASES" },
        { type: "Surveillance", id: "SURVEILLANCE_INDICATORS" },
        { type: "Surveillance", id: "SURVEILLANCE_REGIONS" },
        { type: "Surveillance", id: "SURVEILLANCE_DISTRICTS" },
      ],
    }),
  }),
});

export const {
  useListEpiWeeksQuery,
  useListAlertsQuery,
  useListDiseasesQuery,
  useListRegionsQuery,
  useListDistrictsQuery,
  useListDistrictsByRegionQuery,
  useListSubcountiesByDistrictQuery,
  useGetSubcountyByIdQuery,
  useListFacilityWeeklyMetricsByWeekQuery,
  useListFacilityWeeklyMetricsByFacilityQuery,
  useListWeeklyStatusesDetailedQuery,
  useListWeeklyStatusesQuery,
  useListImportBatchesQuery,
  useGetImportBatchByIdQuery,
  useListImportRawRowsByBatchQuery,
  useCreateImportBatchMutation,
  useUpdateImportBatchStatusMutation,
  useUploadSurveillanceCsvMutation,
} = surveillanceApi;
