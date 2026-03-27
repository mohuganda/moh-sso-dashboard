import type {
  DocumentProcess,
  DocumentProcessType,
  DocumentResponse,
} from "../types/documents.types";
import type { StorageLocation } from "../types/storage.types";
import { baseApi } from "./baseApi";

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
      transformResponse: (response: ApiEnvelope<DocumentResponse[]>) => response.data,
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
    createDocument: builder.mutation<
      DocumentResponse,
      { file: File; storageLocation: string; processType: DocumentProcessType }
    >({
      query: ({ file, storageLocation, processType }) => {
        const formData = new FormData();
        formData.append("file", file);
        formData.append("storage_location", storageLocation);
        formData.append("process_type", processType);

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
    updateDocument: builder.mutation<
      DocumentResponse,
      { id: string; payload: Partial<DocumentResponse> }
    >({
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
    deleteDocument: builder.mutation<{ success: boolean }, string>({
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
    // DOWNLOAD DOCUMENT
    // -----------------------------
    downloadDocument: builder.query<Blob, string>({
      query: (id) => ({
        url: `/documents/${id}/download`,
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

    reprocessDocument: builder.mutation<void, string>({
      query: (id) => ({
        url: `/documents/${id}/reprocess`,
        method: "POST",
      }),
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
  useDownloadDocumentQuery,
  useReprocessDocumentMutation,
  useLazyDownloadDocumentQuery,
  useListStorageLocationsQuery,
  useGetStorageLocationQuery,
} = documentsApi;
