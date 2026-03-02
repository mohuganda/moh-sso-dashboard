export type ProcessStatus = "PENDING" | "PROCESSING" | "COMPLETED" | "FAILED";

/**
 * EXACT backend match
 */
export interface DocumentResponse {
  id: string;
  original_filename: string;
  content_type: string;
  size_bytes: number;
  checksum_sha256?: string | null;
  storage_location: string;
  object_key: string;
  uploaded_by: string;
  created_at: string;
}

export interface DocumentProcess {
  id: string;
  document_id: string;
  status: ProcessStatus;
  progress: number;
  message?: string;
  created_at: string;
}
