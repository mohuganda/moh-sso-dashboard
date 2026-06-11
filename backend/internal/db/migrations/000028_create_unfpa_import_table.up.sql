-- Conditional: only runs when the import schema exists (remote-postgres / file_upload)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = 'import') THEN
    RETURN;
  END IF;

  -- ── UNFPA Pipeline ────────────────────────────────────────────────────────────
  CREATE TABLE IF NOT EXISTS import.unfpa_pipeline (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id         UUID        NOT NULL,
    row_number          INT         NOT NULL,
    report_date         DATE,
    requisition_no      TEXT,
    dept                TEXT,
    product_id          TEXT,
    quantum_item_number TEXT,
    uom                 TEXT,
    moh_units           NUMERIC,
    dkt_units           NUMERIC,
    msi_units           NUMERIC,
    psi_units           NUMERIC,
    ippf_units          NUMERIC,
    total_units         NUMERIC,
    total_cost          NUMERIC,
    unit_price          NUMERIC,
    vendor              TEXT,
    req_line_no         TEXT,
    po_number           TEXT,
    po_due_date         TEXT,
    order_life_cycle    TEXT,
    fund_status         TEXT,
    status              TEXT,
    eta                 TEXT,
    tranche             TEXT,
    funding_year        INT,
    period              TEXT,
    raw_payload         JSONB       NOT NULL DEFAULT '{}',
    row_hash            TEXT,
    is_valid            BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMP   NOT NULL DEFAULT NOW(),
    last_updated        TIMESTAMP   NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_unfpa_pipeline_hash UNIQUE (row_hash)
  );
  CREATE INDEX IF NOT EXISTS idx_unfpa_pipeline_document_id ON import.unfpa_pipeline (document_id);
  CREATE INDEX IF NOT EXISTS idx_unfpa_pipeline_report_date ON import.unfpa_pipeline (report_date);

END;
$$;
