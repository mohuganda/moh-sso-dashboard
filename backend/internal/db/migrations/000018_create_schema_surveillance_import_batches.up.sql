CREATE TABLE surveillance_import_batches (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  source_name TEXT NOT NULL,
  file_name TEXT,
  dataset_type TEXT NOT NULL, -- facility_metrics, district_status, alerts, national_status, region_status
  imported_by TEXT,
  imported_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  status TEXT NOT NULL DEFAULT 'COMPLETED',
  notes TEXT
);