import type {
  CreateDocumentTemplateRequest,
  CreateDocumentTemplateSheetRequest,
  CreateDocumentTemplateColumnRequest,
  DocumentTemplate,
  DocumentTemplateColumn,
  DocumentTemplateSheet,
  TemplateStructure,
  CreateTemplateStructureRequest,
  UpdateTemplatePayload,
  UpdateColumnPayload,
} from "../types/document_template.types";
import { baseApi } from "./baseApi";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

export const documentTemplateApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getTemplates: builder.query<DocumentTemplate[], void>({
      query: () => "/document-templates",
      transformResponse: (res: ApiEnvelope<DocumentTemplate[]>) => res.data ?? [],
      providesTags: ["DocumentTemplates"],
    }),

    listActiveTemplates: builder.query<DocumentTemplate[], void>({
      query: () => "/document-templates?active=true",
      transformResponse: (res: ApiEnvelope<DocumentTemplate[]>) => res.data ?? [],
      providesTags: ["DocumentTemplates"],
    }),

    getTemplateById: builder.query<DocumentTemplate, string>({
      query: (id) => `/document-templates/${id}`,
      transformResponse: (res: ApiEnvelope<DocumentTemplate>) => res.data,
      providesTags: (_, __, id) => [{ type: "DocumentTemplates", id }],
    }),

    getTemplateStructure: builder.query<TemplateStructure, string>({
      query: (code) => `/document-templates/code/${code}/structure`,
      transformResponse: (res: ApiEnvelope<TemplateStructure>) => res.data,
      providesTags: ["DocumentTemplates", "DocumentSheets", "DocumentColumns"],
    }),

    getTemplateHasData: builder.query<{ has_data: boolean }, string>({
      query: (code) => `/document-templates/code/${code}/has-data`,
      transformResponse: (res: ApiEnvelope<{ has_data: boolean }>) => res.data,
    }),

    getTemplateSheets: builder.query<DocumentTemplateSheet[], string>({
      query: (templateId) => `/document-templates/${templateId}/sheets`,
      transformResponse: (res: ApiEnvelope<DocumentTemplateSheet[]>) => res.data,
      providesTags: ["DocumentSheets"],
    }),

    getSheetColumns: builder.query<
      DocumentTemplateColumn[],
      { templateId: string; sheetId: string }
    >({
      query: ({ templateId, sheetId }) =>
        `/document-templates/${templateId}/sheets/${sheetId}/columns`,
      transformResponse: (res: ApiEnvelope<DocumentTemplateColumn[]>) => res.data,
      providesTags: ["DocumentColumns"],
    }),

    createTemplate: builder.mutation<DocumentTemplate, CreateDocumentTemplateRequest>({
      query: (body) => ({
        url: "/document-templates",
        method: "POST",
        body,
      }),
      transformResponse: (res: ApiEnvelope<DocumentTemplate>) => res.data,
      invalidatesTags: ["DocumentTemplates"],
    }),

    createTemplateStructure: builder.mutation<TemplateStructure, CreateTemplateStructureRequest>({
      query: (body) => ({
        url: "/document-templates/structure",
        method: "POST",
        body,
      }),
      transformResponse: (res: ApiEnvelope<TemplateStructure>) => res.data,
      invalidatesTags: ["DocumentTemplates", "DocumentSheets", "DocumentColumns"],
    }),

    createSheet: builder.mutation<DocumentTemplateSheet, CreateDocumentTemplateSheetRequest>({
      query: (body) => ({
        url: "/document-sheets",
        method: "POST",
        body,
      }),
      transformResponse: (res: ApiEnvelope<DocumentTemplateSheet>) => res.data,
      invalidatesTags: ["DocumentSheets"],
    }),

    createColumn: builder.mutation<DocumentTemplateColumn, CreateDocumentTemplateColumnRequest>({
      query: (body) => ({
        url: "/document-columns",
        method: "POST",
        body,
      }),
      transformResponse: (res: ApiEnvelope<DocumentTemplateColumn>) => res.data,
      invalidatesTags: ["DocumentColumns"],
    }),

    updateColumn: builder.mutation<DocumentTemplateColumn, UpdateColumnPayload>({
      query: ({ templateId, sheetId, columnId, ...body }) => ({
        url: `/document-templates/${templateId}/sheets/${sheetId}/columns/${columnId}`,
        method: "PUT",
        body,
      }),
      transformResponse: (res: ApiEnvelope<DocumentTemplateColumn>) => res.data,
      invalidatesTags: ["DocumentTemplates", "DocumentColumns"],
    }),

    updateTemplate: builder.mutation<DocumentTemplate, UpdateTemplatePayload>({
      query: ({ id, ...body }) => ({
        url: `/document-templates/${id}`,
        method: "PUT",
        body,
      }),
      transformResponse: (res: ApiEnvelope<DocumentTemplate>) => res.data,
      invalidatesTags: (_r, _e, { id }) => [{ type: "DocumentTemplates", id }, "DocumentTemplates"],
    }),

    publishTemplate: builder.mutation<void, string>({
      query: (id) => ({ url: `/document-templates/${id}/publish`, method: "POST" }),
      invalidatesTags: ["DocumentTemplates"],
    }),

    archiveTemplate: builder.mutation<void, string>({
      query: (id) => ({ url: `/document-templates/${id}/archive`, method: "POST" }),
      invalidatesTags: ["DocumentTemplates"],
    }),

    deleteTemplate: builder.mutation<void, string>({
      query: (id) => ({ url: `/document-templates/${id}`, method: "DELETE" }),
      invalidatesTags: ["DocumentTemplates"],
    }),
  }),
});

export const {
  useGetTemplatesQuery,
  useListActiveTemplatesQuery,
  useGetTemplateByIdQuery,
  useGetTemplateStructureQuery,
  useGetTemplateHasDataQuery,
  useGetTemplateSheetsQuery,
  useGetSheetColumnsQuery,
  useCreateTemplateMutation,
  useUpdateTemplateMutation,
  useUpdateColumnMutation,
  useCreateTemplateStructureMutation,
  useCreateSheetMutation,
  useCreateColumnMutation,
  usePublishTemplateMutation,
  useArchiveTemplateMutation,
  useDeleteTemplateMutation,
} = documentTemplateApi;
