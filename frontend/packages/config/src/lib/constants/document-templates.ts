import type { DocumentProcessType } from "@moh-sso/types";

export type StandardTemplateKey =
  | ""
  | "district_status"
  | "facility_metrics"
  | "alerts"
  | "lab_results";

export type StandardTemplate = {
  key: Exclude<StandardTemplateKey, "">;
  label: string;
  requiredHeaders: string[];
  optionalHeaders?: string[];
  allowedExtensions: string[];
  suggestedProcessType?: DocumentProcessType;
};

export const STANDARD_TEMPLATES: StandardTemplate[] = [
  {
    key: "district_status",
    label: "District Status Template",
    requiredHeaders: ["district", "epi_week", "epi_year", "disease", "status"],
    optionalHeaders: ["region", "source_name"],
    allowedExtensions: [".csv", ".xlsx", ".xls"],
    suggestedProcessType: "SURVEILLANCE_CSV_IMPORT",
  },
  {
    key: "facility_metrics",
    label: "Facility Metrics Template",
    requiredHeaders: ["facility", "epi_week", "epi_year", "indicator", "value"],
    optionalHeaders: ["district", "region", "source_name"],
    allowedExtensions: [".csv", ".xlsx", ".xls"],
    suggestedProcessType: "FACILITY_IMPORT",
  },

  {
    key: "lab_results",
    label: "Lab Results Template",
    requiredHeaders: ["sample_id", "district", "disease", "result", "date_reported"],
    optionalHeaders: ["facility", "patient_id"],
    allowedExtensions: [".csv", ".xlsx", ".xls"],
    suggestedProcessType: "LAB_RESULT_IMPORT",
  },
];
