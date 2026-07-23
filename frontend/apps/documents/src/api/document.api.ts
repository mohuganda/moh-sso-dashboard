import type {
  CreateDocumentPayload,
  DataPreviewResponse,
  DocumentTemplate,
  DocumentProcess,
  DocumentResponse,
  StorageLocation,
  UpdateDocumentPayload,
} from "../types";
import { baseApi } from "@moh-sso/api";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

export const documentsApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    // -----------------------------
    // LIST DOCUMENTS
    // -----------------------------
    listDocuments: builder.query<DocumentResponse[], void>({
      query: () => ({
        url: `/documents`,
        method: "GET",
      }),
      transformResponse: (response: ApiEnvelope<DocumentResponse[]>) => response.data ?? [],
      providesTags: (result) =>
        result
          ? [
              { type: "Documents", id: "LIST" },
              ...result.map((doc) => ({
                type: "Document" as const,
                id: doc.id,
              })),
            ]
          : [{ type: "Documents", id: "LIST" }],
    }),

    // -----------------------------
    // GET SINGLE DOCUMENT
    // -----------------------------
    getDocument: builder.query<DocumentResponse, string>({
      query: (id) => ({
        url: `/documents/${id}`,
        method: "GET",
      }),
      transformResponse: (response: ApiEnvelope<DocumentResponse>) => response.data,
      providesTags: (_res, _err, id) => [{ type: "Document", id }],
    }),

    // -----------------------------
    // GET DOCUMENT PROCESSES
    // -----------------------------
    getDocumentProcesses: builder.query<DocumentProcess[], string>({
      query: (documentId) => ({
        url: `/documents/${documentId}/processes`,
        method: "GET",
      }),
      transformResponse: (response: ApiEnvelope<DocumentProcess[]>) => response.data,
      providesTags: (_res, _err, documentId) => [{ type: "DocumentProcesses", id: documentId }],
    }),

    // -----------------------------
    // CREATE DOCUMENT (UPLOAD)
    // -----------------------------
    createDocument: builder.mutation<DocumentResponse, CreateDocumentPayload>({
      query: ({ file, storageLocation, processType, isTemplate, metadata }) => {
        const formData = new FormData();

        formData.append("file", file);
        formData.append("storage_location", storageLocation);
        formData.append("is_template", String(Boolean(isTemplate)));

        if (processType) {
          formData.append("process_type", processType);
        }

        if (metadata) {
          formData.append("metadata", JSON.stringify(metadata));
        }

        return {
          url: `/documents`,
          method: "POST",
          body: formData,
        };
      },

      transformResponse: (response: ApiEnvelope<DocumentResponse>) => response.data,

      invalidatesTags: [{ type: "Documents", id: "LIST" }],
    }),

    // -----------------------------
    // UPDATE DOCUMENT
    // -----------------------------
    updateDocument: builder.mutation<DocumentResponse, UpdateDocumentPayload>({
      query: ({ id, payload }) => ({
        url: `/documents/${id}`,
        method: "PUT",
        body: payload,
      }),
      transformResponse: (response: ApiEnvelope<DocumentResponse>) => response.data,
      invalidatesTags: (_res, _err, arg) => [
        { type: "Document", id: arg.id },
        { type: "Documents", id: "LIST" },
      ],
    }),

    // -----------------------------
    // DELETE DOCUMENT
    // -----------------------------
    deleteDocument: builder.mutation<{ message: string }, string>({
      query: (id) => ({
        url: `/documents/${id}`,
        method: "DELETE",
      }),
      invalidatesTags: (_res, _err, id) => [
        { type: "Document", id },
        { type: "Documents", id: "LIST" },
      ],
    }),

    // -----------------------------
    // VIEW DOCUMENT INLINE
    // -----------------------------
    viewDocument: builder.query<Blob, string>({
      query: (id) => ({
        url: `/documents/files/${id}/view`,
        method: "GET",
        responseHandler: (response) => response.blob(),
      }),
    }),

    // -----------------------------
    // DOWNLOAD DOCUMENT
    // -----------------------------
    downloadDocument: builder.query<Blob, string>({
      query: (id) => ({
        url: `/documents/files/${id}/download`,
        method: "GET",
        responseHandler: (response) => response.blob(),
      }),
    }),

    // -----------------------------
    // LIST STORAGE LOCATIONS
    // -----------------------------
    listStorageLocations: builder.query<StorageLocation[], void>({
      query: () => ({
        url: `/storage-locations`,
        method: "GET",
      }),
      transformResponse: (response: ApiEnvelope<StorageLocation[]>) => response.data,
      providesTags: [{ type: "StorageLocations", id: "LIST" }],
    }),

    // -----------------------------
    // GET SINGLE STORAGE LOCATION
    // -----------------------------
    getStorageLocation: builder.query<StorageLocation, string>({
      query: (id) => ({
        url: `/storage-locations/${id}`,
        method: "GET",
      }),
      transformResponse: (response: ApiEnvelope<StorageLocation>) => response.data,
      providesTags: (_res, _err, id) => [{ type: "StorageLocation", id }],
    }),

    // -----------------------------
    // REPROCESS DOCUMENT
    // -----------------------------
    reprocessDocument: builder.mutation<{ message: string }, string>({
      query: (id) => ({
        url: `/documents/${id}/reprocess`,
        method: "POST",
      }),
      invalidatesTags: (_res, _err, id) => [
        { type: "Document", id },
        { type: "DocumentProcesses", id },
        { type: "Documents", id: "LIST" },
      ],
    }),

    listDocumentTemplates: builder.query<DocumentTemplate[], void>({
      query: () => "/document-templates",
      transformResponse: (res: ApiEnvelope<DocumentTemplate[]>) => res.data,
      providesTags: ["DocumentTemplates"],
    }),

    // -----------------------------
    // DOCUMENT STATS
    // -----------------------------
    getDocumentStats: builder.query<{
      total: number;
      total_size: number;
      pending: number;
      processing: number;
      completed: number;
      failed: number;
    }, void>({
      query: () => `/documents/stats`,
      transformResponse: (res: ApiEnvelope<{
        total: number; total_size: number;
        pending: number; processing: number; completed: number; failed: number;
      }>) => res.data,
      providesTags: [{ type: "Documents", id: "LIST" }],
    }),

    // -----------------------------
    // DATA PREVIEW
    // -----------------------------
    getDocumentDataPreview: builder.query<DataPreviewResponse, string>({
      query: (id) => `/documents/${id}/data-preview`,
      transformResponse: (res: ApiEnvelope<DataPreviewResponse>) => res.data,
      providesTags: (_res, _err, id) => [{ type: "Document", id }],
    }),

    exportDataPreview: builder.mutation<
      Blob,
      {
        id: string;
        sheetCode: string;
        columns: string[];
        search?: string[];
        filters?: Record<string, string[]>;
      }
    >({
      query: ({ id, sheetCode, columns, search, filters }) => ({
        url: `/documents/${id}/data-preview/export`,
        method: "POST",
        body: {
          sheet_code: sheetCode,
          columns,
          search,
          filters,
        },
        responseHandler: (response) => response.blob(),
      }),
    }),

    // -----------------------------
    // PARSE STRUCTURE (server-side detection from stored document)
    // -----------------------------
    parseDocumentStructure: builder.query<{
      sheets: {
        name: string;
        header_row: number;
        start_row: number;
        columns: { column_key: string; column_name: string }[];
      }[];
    }, string>({
      query: (id) => `/documents/${id}/parse-structure`,
      transformResponse: (res: ApiEnvelope<{
        sheets: {
          name: string;
          header_row: number;
          start_row: number;
          columns: { column_key: string; column_name: string }[];
        }[];
      }>) => res.data,
    }),

    // -----------------------------
    // SCAN STRUCTURE (server-side detection from uploaded file, no storage)
    // -----------------------------
    scanDocumentStructure: builder.mutation<{
      sheets: {
        name: string;
        header_row: number;
        start_row: number;
        columns: { column_key: string; column_name: string }[];
      }[];
    }, File>({
      query: (file) => {
        const formData = new FormData();
        formData.append("file", file);
        return {
          url: `/documents/scan-structure`,
          method: "POST",
          body: formData,
        };
      },
      transformResponse: (res: ApiEnvelope<{
        sheets: {
          name: string;
          header_row: number;
          start_row: number;
          columns: { column_key: string; column_name: string }[];
        }[];
      }>) => res.data,
    }),
  }),
});

export const {
  useListDocumentsQuery,
  useGetDocumentQuery,
  useGetDocumentProcessesQuery,
  useCreateDocumentMutation,
  useUpdateDocumentMutation,
  useDeleteDocumentMutation,
  useViewDocumentQuery,
  useLazyViewDocumentQuery,
  useDownloadDocumentQuery,
  useLazyDownloadDocumentQuery,
  useReprocessDocumentMutation,
  useListStorageLocationsQuery,
  useGetStorageLocationQuery,
  useListDocumentTemplatesQuery,
  useGetDocumentDataPreviewQuery,
  useLazyGetDocumentDataPreviewQuery,
  useExportDataPreviewMutation,
  useGetDocumentStatsQuery,
  useLazyParseDocumentStructureQuery,
  useScanDocumentStructureMutation,
} = documentsApi;
