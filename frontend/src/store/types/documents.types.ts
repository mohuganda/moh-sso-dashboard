export type ProcessStatus = "PENDING" | "PROCESSING" | "COMPLETED" | "FAILED";

/**
 * EXACT backend match
 */

export type CreateDocumentPayload = {
  file: File;
  storageLocation: string;
  isTemplate?: boolean;
  processType?: DocumentProcessType;
  metadata?: Record<string, unknown>;
};

export type UpdateDocumentPayload = {
  id: string;
  payload: {
    original_filename: string;
    content_type: string;
  };
};
export type DocumentResponse = {
  id: string;
  original_filename: string;
  content_type: string;
  size_bytes: number;
  checksum_sha256?: string | null;
  storage_location: string;
  object_key: string;
  uploaded_by: string;
  status: string;
  metadata?: Record<string, string> | null;
  object_url?: string;
  view_url?: string;
  download_url?: string;
  created_at: string;
  updated_at: string;
};

export type DataPreviewSheet = {
  name: string;
  row_count: number;
  columns: string[];
  rows: Record<string, unknown>[];
};

export type DataPreviewResponse = {
  template_code: string;
  report_date?: string;
  sheets: DataPreviewSheet[];
};

export interface DocumentProcess {
  id: string;
  document_id: string;
  status: ProcessStatus;
  progress: number;
  message?: string;
  created_at: string;
  started_at: string;
  finished_at: string;
  updated_at: string;
}
export type DocumentProcessType =
  | "SURVEILLANCE_CSV_IMPORT"
  | "CSV_IMPORT"
  | "EXCEL_IMPORT"
  | "FHIR_IMPORT"
  | "USER_BULK_IMPORT"
  | "SURVEILLANCE_EXCEL_IMPORT"
  | "FACILITY_IMPORT"
  | "LAB_RESULT_IMPORT"
  | "INVENTORY_IMPORT"
  | "HISTORICAL_BACKFILL_IMPORT"
  | "DRY_RUN_IMPORT"
  | "ETL_IMPORT"
  | "ARCHIVE_IMPORT"
  | "IMAGE_PROCESSING"
  | "DOCUMENT_EXTRACTION"
  | "NO_OP_UPLOAD";

export const DOCUMENT_PROCESS_TYPE_OPTIONS: Array<{
  value: DocumentProcessType;
  label: string;
}> = [
  { value: "SURVEILLANCE_CSV_IMPORT", label: "Surveillance CSV Import" },
  { value: "CSV_IMPORT", label: "CSV Import" },
  { value: "EXCEL_IMPORT", label: "Excel Import" },
  { value: "FHIR_IMPORT", label: "FHIR Import" },
  { value: "USER_BULK_IMPORT", label: "User Bulk Import" },
  { value: "SURVEILLANCE_EXCEL_IMPORT", label: "Surveillance Excel Import" },
  { value: "FACILITY_IMPORT", label: "Facility Import" },
  { value: "LAB_RESULT_IMPORT", label: "Lab Result Import" },
  { value: "INVENTORY_IMPORT", label: "Inventory Import" },
  { value: "HISTORICAL_BACKFILL_IMPORT", label: "Historical Backfill Import" },
  { value: "DRY_RUN_IMPORT", label: "Dry Run Import" },
  { value: "ETL_IMPORT", label: "ETL Import" },
  { value: "ARCHIVE_IMPORT", label: "Archive Import" },
  { value: "IMAGE_PROCESSING", label: "Image Processing" },
  { value: "DOCUMENT_EXTRACTION", label: "Document Extraction" },
  { value: "NO_OP_UPLOAD", label: "No-op Upload" },
];
