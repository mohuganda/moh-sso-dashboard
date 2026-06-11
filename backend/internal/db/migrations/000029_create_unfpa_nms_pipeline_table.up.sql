-- Conditional: only runs when the import schema exists (remote-postgres / file_upload)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = 'import') THEN
    RETURN;
  END IF;

  -- ── UNFPA NMS Pipeline ────────────────────────────────────────────────────────
  CREATE TABLE IF NOT EXISTS import.unfpa_nms_pipeline (
    id                   UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id          UUID        NOT NULL,
    row_number           INT         NOT NULL,
    report_date          DATE,
    item                 TEXT,
    item_id              TEXT,
    item_name            TEXT,
    mot                  TEXT,
    eta                  TEXT,
    quantity             NUMERIC,
    value                NUMERIC,
    supplier             TEXT,
    po_number            TEXT,
    in_production_until  TEXT,
    eta_as_per_offer     TEXT,
    status               TEXT,
    raw_payload          JSONB       NOT NULL DEFAULT '{}',
    row_hash             TEXT,
    is_valid             BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at           TIMESTAMP   NOT NULL DEFAULT NOW(),
    last_updated         TIMESTAMP   NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_unfpa_nms_pipeline_hash UNIQUE (row_hash)
  );
  CREATE INDEX IF NOT EXISTS idx_unfpa_nms_pipeline_document_id ON import.unfpa_nms_pipeline (document_id);
  CREATE INDEX IF NOT EXISTS idx_unfpa_nms_pipeline_report_date ON import.unfpa_nms_pipeline (report_date);

END;
$$;
