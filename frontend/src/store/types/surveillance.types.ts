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
  facility_name?: string;
  region_id?: string;
  region_name?: string;
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

export type RiskLevel = "MAROON" | "RED" | "YELLOW" | "GREEN";

export interface WeeklyStatus {
  id: string;
  region_id: string | null;
  district_id: string | null;
  sub_county_id: string | null;
  disease_id: string | null;
  indicator_id: string | null;
  epi_week_id: string;
  status: RiskLevel;
  source_name: string | null;
  imported_at: string;
  created_at: string;
}

export interface WeeklyStatusDetailed {
  id: string;
  region_id: string | null;
  district_id: string | null;
  sub_county_id: string | null;
  disease_id: string | null;
  indicator_id: string | null;
  epi_week_id: string;
  status: RiskLevel;
  source_name: string | null;
  imported_at: string | null;
  created_at: string;

  region_name: string | null;
  district_name: string | null;
  sub_county_name: string | null;
  disease_name: string | null;
  indicator_name: string | null;
}

export interface ListWeeklyStatusesParams {
  epiWeekID?: string;
  regionID?: string;
  districtID?: string;
  subCountyID?: string;
  diseaseID?: string;
  indicatorID?: string;
  status?: RiskLevel;
}

export type ListAlertsParams = {
  epiWeekID?: string;
  districtID?: string;
  regionID?: string;
  diseaseID?: string;
  status?: string;
};

export interface Subcounty {
  id: string;
  name: string;
  district_id: string;
}

export interface Alert {
  id: string;
  external_id?: string | null;
  disease_id: string;
  district_id?: string | null;
  epi_week_id?: string | null;
  occurred_on?: string | null;
  created_on?: string | null;
  narrative: string;
  submitted_by?: string | null;
  status: string;
  source_name?: string | null;
  created_at: string;
  updated_at: string;

  disease_name: string;
  district_name?: string | null;
  epi_year?: number | null;
  epi_week?: number | null;
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
