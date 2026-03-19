CREATE TABLE surveillance_import_raw_rows (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  batch_id UUID NOT NULL REFERENCES surveillance_import_batches(id) ON DELETE CASCADE,
  row_number INT NOT NULL,
  payload JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_surveillance_import_raw_rows_batch ON surveillance_import_raw_rows(batch_id);
CREATE INDEX idx_surveillance_import_raw_rows_payload ON surveillance_import_raw_rows USING GIN(payload);