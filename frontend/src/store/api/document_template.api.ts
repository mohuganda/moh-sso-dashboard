import type {
  DocumentTemplate,
  DocumentTemplateColumn,
  DocumentTemplateSheet,
} from "../types/document_template.types";
import { baseApi } from "./baseApi";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

export const documentsApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getTemplates: builder.query<DocumentTemplate[], void>({
      query: () => "/document-templates",
      transformResponse: (res: ApiEnvelope<DocumentTemplate[]>) => res.data,
      providesTags: ["DocumentTemplates"],
    }),

    getTemplateById: builder.query<DocumentTemplate, string>({
      query: (id) => `/document-templates/${id}`,
      transformResponse: (res: ApiEnvelope<DocumentTemplate>) => res.data,
      providesTags: (_, __, id) => [{ type: "DocumentTemplates", id }],
    }),

    getTemplateSheets: builder.query<DocumentTemplateSheet[], string>({
      query: (templateId) => `/document-templates/${templateId}/sheets`,
      transformResponse: (res: ApiEnvelope<DocumentTemplateSheet[]>) => res.data,
      providesTags: ["DocumentSheets"],
    }),

    getSheetById: builder.query<DocumentTemplateSheet, string>({
      query: (id) => `/document-sheets/${id}`,
      transformResponse: (res: ApiEnvelope<DocumentTemplateSheet>) => res.data,
    }),

    getSheetColumns: builder.query<DocumentTemplateColumn[], string>({
      query: (sheetId) => `/document-sheets/${sheetId}/columns`,
      transformResponse: (res: ApiEnvelope<DocumentTemplateColumn[]>) => res.data,
      providesTags: ["DocumentColumns"],
    }),

    createTemplate: builder.mutation<DocumentTemplate, Partial<DocumentTemplate>>({
      query: (body) => ({
        url: "/document-templates",
        method: "POST",
        body,
      }),
      invalidatesTags: ["DocumentTemplates"],
    }),

    createSheet: builder.mutation<DocumentTemplateSheet, Partial<DocumentTemplateSheet>>({
      query: (body) => ({
        url: "/document-sheets",
        method: "POST",
        body,
      }),
      invalidatesTags: ["DocumentSheets"],
    }),

    createColumn: builder.mutation<DocumentTemplateColumn, Partial<DocumentTemplateColumn>>({
      query: (body) => ({
        url: "/document-columns",
        method: "POST",
        body,
      }),
      invalidatesTags: ["DocumentColumns"],
    }),
  }),
});

export const {
  useGetTemplatesQuery,
  useGetTemplateByIdQuery,
  useGetTemplateSheetsQuery,
  useGetSheetByIdQuery,
  useGetSheetColumnsQuery,
  useCreateTemplateMutation,
  useCreateSheetMutation,
  useCreateColumnMutation,
} = documentsApi;
