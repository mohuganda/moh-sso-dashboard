import type {
  CreateDocumentTemplateRequest,
  CreateDocumentTemplateSheetRequest,
  CreateDocumentTemplateColumnRequest,
  DocumentTemplate,
  DocumentTemplateColumn,
  DocumentTemplateSheet,
  TemplateStructure,
  CreateTemplateStructureRequest,
} from "@moh-sso/types";
import { baseApi } from "@moh-sso/api";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

export const documentTemplateApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getTemplates: builder.query<DocumentTemplate[], void>({
      query: () => "/document-templates",
      transformResponse: (res: ApiEnvelope<DocumentTemplate[]>) => res.data,
      providesTags: ["DocumentTemplates"],
    }),

    listActiveTemplates: builder.query<DocumentTemplate[], void>({
      query: () => "/document-templates?active=true",
      transformResponse: (res: ApiEnvelope<DocumentTemplate[]>) => res.data,
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
  }),
});

export const {
  useGetTemplatesQuery,
  useListActiveTemplatesQuery,
  useGetTemplateByIdQuery,
  useGetTemplateStructureQuery,
  useGetTemplateSheetsQuery,
  useGetSheetColumnsQuery,
  useCreateTemplateMutation,
  useCreateTemplateStructureMutation,
  useCreateSheetMutation,
  useCreateColumnMutation,
} = documentTemplateApi;
