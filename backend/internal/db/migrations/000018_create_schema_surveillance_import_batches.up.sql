CREATE TABLE surveillance_import_batches (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  source_name TEXT NOT NULL,
  file_name TEXT,
  dataset_type TEXT NOT NULL, -- facility_metrics, district_status, alerts, national_status, region_status
  imported_by TEXT,
  imported_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  status TEXT NOT NULL DEFAULT 'PENDING',
  notes TEXT
);


ALTER TABLE surveillance_import_batches
ADD COLUMN total_rows INT NOT NULL DEFAULT 0,
ADD COLUMN success_rows INT NOT NULL DEFAULT 0,
ADD COLUMN failed_rows INT NOT NULL DEFAULT 0,
ADD COLUMN completed_at TIMESTAMPTZ;

ALTER TABLE surveillance_import_batches
ADD CONSTRAINT surveillance_import_batches_status_chk
CHECK (status IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED'));