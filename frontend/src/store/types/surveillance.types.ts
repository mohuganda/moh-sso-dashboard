export interface EpiWeek {
  id: string;
  epi_year: number;
  epi_week: number;
  start_date?: string;
  end_date?: string;
}

export interface Disease {
  id: string;
  name: string;
  code?: string;
}

export interface District {
  id: string;
  name: string;
  region_id: string;
  region_name?: string;
  code: string;
}

export interface Region {
  id: string;
  name: string;
  code: string;
}

export interface FacilityWeeklyMetric {
  id: string;
  facility_id: string;
  facility_name: string;
  district_id?: string;
  district_name?: string;
  subcounty_id?: string;
  subcounty_name?: string;
  disease_id?: string;
  disease_name?: string;
  indicator_id?: string;
  indicator_name?: string;
  value: number;
  epi_week_id: string;
  year?: number;
  week?: number;
}

export interface DistrictWeeklyStatus {
  id: string;
  district_id: string;
  district_name?: string;
  epi_week_id: string;
  maroon?: string[];
  red?: string[];
  yellow?: string[];
  green?: string[];
}

export interface RegionWeeklyStatus {
  id: string;
  region_id: string;
  region_name?: string;
  epi_week_id: string;
  maroon?: number;
  red?: number;
  yellow?: number;
  green?: number;
}

export interface NationalWeeklyStatus {
  id: string;
  epi_week_id: string;
  maroon?: number;
  red?: number;
  yellow?: number;
  green?: number;
}

export interface Subcounty {
  id: string;
  name: string;
  district_id: string;
}

export type SurveillanceImportBatch = {
  id: string;
  source_name: string;
  file_name?: string | null;
  dataset_type: string;
  imported_by?: string | null;
  imported_at: string;
  status: string;
  notes?: string | null;
};

export type SurveillanceImportRawRow = {
  id: string;
  batch_id: string;
  row_number: number;
  payload: Record<string, unknown>;
  created_at: string;
};

export type UpdateImportBatchStatusPayload = {
  batchID: string;
  status: string;
  notes?: string | null;
};

export type CreateImportBatchPayload = {
  source_name: string;
  file_name?: string | null;
  dataset_type: string;
  imported_by?: string | null;
  status?: string;
  notes?: string | null;
};

export type UploadSurveillanceCsvRequest = {
  file: File;
  storage_location?: string;
};
