export type DocumentTemplate = {
  id: string;
  document_id?: string | null;
  code: string;
  name: string;
  description?: string | null;
  file_type: string;
  version: number;
  is_active: boolean;
  configuration: Record<string, unknown> | null;
  created_by?: string;
  created_at: string;
  updated_at: string;
  archived_at?: string | null;
};

export type CreateDocumentTemplateRequest = {
  document_id?: string | null;
  code: string;
  name: string;
  description?: string | null;
  file_type: string;
  version?: number;
  is_active?: boolean;
  configuration?: Record<string, unknown> | null;
};

export type DocumentTemplateSheet = {
  id: string;
  template_id: string;
  code: string;
  name: string;
  display_name?: string | null;
  required: boolean;
  sheet_order?: number | null;
  header_row: number;
  start_row: number;
  allow_extra_columns: boolean;
  allow_duplicate_headers: boolean;
  configuration: Record<string, unknown> | null;
  created_at: string;
  updated_at: string;
  archived_at?: string | null;
};

export type CreateDocumentTemplateSheetRequest = {
  template_id: string;
  code: string;
  name: string;
  display_name?: string | null;
  required?: boolean;
  sheet_order?: number | null;
  header_row?: number;
  start_row?: number;
  allow_extra_columns?: boolean;
  allow_duplicate_headers?: boolean;
  configuration?: Record<string, unknown> | null;
};

export type DocumentTemplateColumn = {
  id: string;
  sheet_id: string;
  column_key: string;
  column_name: string;
  display_name?: string | null;
  data_type: DocumentTemplateColumnType;
  required: boolean;
  is_unique: boolean;
  column_order?: number | null;
  default_value?: string | null;
  configuration: Record<string, unknown> | null;
  allowed_values: unknown[] | null;
  aliases: string[] | null;
  created_at: string;
  updated_at: string;
  archived_at?: string | null;
};

export type DocumentTemplateColumnType =
  | "text"
  | "number"
  | "integer"
  | "decimal"
  | "boolean"
  | "date"
  | "datetime"
  | "uuid"
  | "json";

export type CreateDocumentTemplateColumnRequest = {
  sheet_id: string;
  column_key: string;
  column_name: string;
  display_name?: string | null;
  data_type: DocumentTemplateColumnType;
  required?: boolean;
  is_unique?: boolean;
  column_order?: number | null;
  default_value?: string | null;
  configuration?: Record<string, unknown> | null;
  allowed_values?: unknown[] | null;
  aliases?: string[] | null;
};

export type CreateTemplateStructureRequest = {
  template: CreateTemplateRequest;
  sheets: CreateTemplateSheetWithColumnsRequest[];
};

export type CreateTemplateSheetWithColumnsRequest = {
  sheet: CreateSheetRequest;
  columns: CreateColumnRequest[];
};

export type CreateTemplateRequest = {
  document_id?: string | null;
  code: string;
  name: string;
  description: string;
  file_type: string;
  created_by?: string;
  configuration: Record<string, unknown>;
};

export type CreateSheetRequest = {
  template_id?: string;
  code: string;
  name: string;
  display_name: string;
  required: boolean;
  sheet_order?: number | null;
  header_row: number;
  start_row: number;
  allow_extra_columns: boolean;
  allow_duplicate_headers: boolean;
  configuration: Record<string, unknown>;
};

export type CreateColumnRequest = {
  sheet_id?: string;
  column_key: string;
  column_name: string;
  display_name: string;
  data_type: string;
  required: boolean;
  is_unique: boolean;
  column_order?: number | null;
  default_value?: string | null;
  allowed_values: string[];
  aliases: string[];
  configuration: Record<string, unknown>;
};

export type TemplateStructure = {
  template: DocumentTemplate;
  sheets: TemplateSheetStructure[];
};

export type TemplateSheetStructure = {
  id: string;
  template_id: string;
  code: string;
  name: string;
  display_name: string;
  required: boolean;
  sheet_order?: number | null;
  header_row: number;
  start_row: number;
  allow_extra_columns: boolean;
  allow_duplicate_headers: boolean;
  configuration: Record<string, unknown>;
  columns: TemplateColumnStructure[];
};

export type TemplateColumnStructure = {
  id: string;
  sheet_id: string;
  column_key: string;
  column_name: string;
  display_name: string;
  data_type: string;
  required: boolean;
  is_unique: boolean;
  column_order?: number | null;
  default_value?: string | null;
  allowed_values: string[];
  aliases: string[];
  configuration: Record<string, unknown>;
};
