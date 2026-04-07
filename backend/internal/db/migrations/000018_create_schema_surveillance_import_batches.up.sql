CREATE TABLE surveillance_import_batches (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  source_name TEXT NOT NULL,
  file_name TEXT,
  dataset_type TEXT NOT NULL,
  imported_by TEXT,
  imported_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  status TEXT NOT NULL DEFAULT 'PENDING',
  notes TEXT,
  total_rows INT NOT NULL DEFAULT 0,
  success_rows INT NOT NULL DEFAULT 0,
  failed_rows INT NOT NULL DEFAULT 0,
  completed_at TIMESTAMPTZ,
  document_id UUID REFERENCES documents(id) ON DELETE SET NULL
);

ALTER TABLE surveillance_import_batches
ADD CONSTRAINT surveillance_import_batches_status_chk
CHECK (status IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED'));