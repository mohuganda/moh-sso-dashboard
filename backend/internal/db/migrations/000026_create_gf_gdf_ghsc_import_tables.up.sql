-- Conditional: only runs when the import schema exists (remote-postgres / file_upload)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = 'import') THEN
    RETURN;
  END IF;

  -- ── Global Fund Pipeline TrackAndTrace ──────────────────────────────────────
  CREATE TABLE IF NOT EXISTS import.gf_pipeline (
    id                              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id                     UUID        NOT NULL,
    row_number                      INT         NOT NULL,
    report_date                     DATE,
    psa_name                        TEXT,
    country                         TEXT,
    grant_name                      TEXT,
    grant_budget_id                 TEXT,
    req_number                      TEXT,
    epo_number                      TEXT,
    po_number                       TEXT,
    shipment_number                 TEXT,
    status                          TEXT,
    vendor_group_name               TEXT,
    item_name_tgf                   TEXT,
    sq_so_item_quantity             NUMERIC,
    inco_term_client                TEXT,
    shipment_mode                   TEXT,
    country_ship_to_city            TEXT,
    po_item_quantity                NUMERIC,
    shipment_item_qty_ordered       NUMERIC,
    confirmed_received_quantity     NUMERIC,
    number_of_pallets               NUMERIC,
    shipment_gross_weight           NUMERIC,
    shipment_volume                 NUMERIC,
    number_of_containers_total      NUMERIC,
    estimated_vendor_inco_date      DATE,
    estimated_delivery_date         DATE,
    delivery_date_actual            DATE,
    so_item_linenumber              TEXT,
    raw_payload                     JSONB       NOT NULL DEFAULT '{}',
    row_hash                        TEXT,
    is_valid                        BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at                      TIMESTAMP   NOT NULL DEFAULT NOW(),
    last_updated                    TIMESTAMP   NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_gf_pipeline_hash UNIQUE (row_hash)
  );
  CREATE INDEX IF NOT EXISTS idx_gf_pipeline_document_id ON import.gf_pipeline (document_id);
  CREATE INDEX IF NOT EXISTS idx_gf_pipeline_report_date ON import.gf_pipeline (report_date);

  -- ── GDF TB Orders ────────────────────────────────────────────────────────────
  CREATE TABLE IF NOT EXISTS import.gdf_tb_orders (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id         UUID        NOT NULL,
    row_number          INT         NOT NULL,
    report_date         DATE,
    year_created        INT,
    country             TEXT,
    line                TEXT,
    serial_number       TEXT,
    total_cost          NUMERIC,
    order_status        TEXT,
    shipment_code       TEXT,
    product_code        TEXT,
    inn_code            TEXT,
    supplier            TEXT,
    quantity_shipped    NUMERIC,
    units_shipped       NUMERIC,
    price               NUMERIC,
    estimated_delivery  DATE,
    actual_delivery     DATE,
    inco_term           TEXT,
    shipping_mode       TEXT,
    raw_payload         JSONB       NOT NULL DEFAULT '{}',
    row_hash            TEXT,
    is_valid            BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMP   NOT NULL DEFAULT NOW(),
    last_updated        TIMESTAMP   NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_gdf_tb_orders_hash UNIQUE (row_hash)
  );
  CREATE INDEX IF NOT EXISTS idx_gdf_tb_orders_document_id ON import.gdf_tb_orders (document_id);
  CREATE INDEX IF NOT EXISTS idx_gdf_tb_orders_report_date ON import.gdf_tb_orders (report_date);

  -- ── GHSC-PSM Lab ─────────────────────────────────────────────────────────────
  CREATE TABLE IF NOT EXISTS import.ghsc_psm_lab (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id             UUID        NOT NULL,
    row_number              INT         NOT NULL,
    report_date             DATE,
    line_number             TEXT,
    commodity_category      TEXT,
    item_description        TEXT,
    uom                     TEXT,
    stock_on_hand           NUMERIC,
    amc                     NUMERIC,
    mos                     NUMERIC,
    quantity_on_order       NUMERIC,
    requested_delivery_date DATE,
    estimated_delivery_date DATE,
    status                  TEXT,
    comment                 TEXT,
    raw_payload             JSONB       NOT NULL DEFAULT '{}',
    row_hash                TEXT,
    is_valid                BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMP   NOT NULL DEFAULT NOW(),
    last_updated            TIMESTAMP   NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_ghsc_psm_lab_hash UNIQUE (row_hash)
  );
  CREATE INDEX IF NOT EXISTS idx_ghsc_psm_lab_document_id ON import.ghsc_psm_lab (document_id);

  -- ── GHSC-PSM Pharma ───────────────────────────────────────────────────────────
  CREATE TABLE IF NOT EXISTS import.ghsc_psm_pharma (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id             UUID        NOT NULL,
    row_number              INT         NOT NULL,
    report_date             DATE,
    line_number             TEXT,
    commodity_category      TEXT,
    item_description        TEXT,
    uom                     TEXT,
    stock_on_hand           NUMERIC,
    amc                     NUMERIC,
    mos                     NUMERIC,
    quantity_on_order       NUMERIC,
    estimated_delivery_date DATE,
    status                  TEXT,
    raw_payload             JSONB       NOT NULL DEFAULT '{}',
    row_hash                TEXT,
    is_valid                BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMP   NOT NULL DEFAULT NOW(),
    last_updated            TIMESTAMP   NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_ghsc_psm_pharma_hash UNIQUE (row_hash)
  );
  CREATE INDEX IF NOT EXISTS idx_ghsc_psm_pharma_document_id ON import.ghsc_psm_pharma (document_id);

  -- ── GHSC-PSM Malaria ──────────────────────────────────────────────────────────
  CREATE TABLE IF NOT EXISTS import.ghsc_psm_malaria (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id             UUID        NOT NULL,
    row_number              INT         NOT NULL,
    report_date             DATE,
    line_number             TEXT,
    commodity_category      TEXT,
    item_description        TEXT,
    uom                     TEXT,
    stock_on_hand           NUMERIC,
    amc                     NUMERIC,
    mos                     NUMERIC,
    quantity_on_order       NUMERIC,
    estimated_delivery_date DATE,
    status                  TEXT,
    raw_payload             JSONB       NOT NULL DEFAULT '{}',
    row_hash                TEXT,
    is_valid                BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMP   NOT NULL DEFAULT NOW(),
    last_updated            TIMESTAMP   NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_ghsc_psm_malaria_hash UNIQUE (row_hash)
  );
  CREATE INDEX IF NOT EXISTS idx_ghsc_psm_malaria_document_id ON import.ghsc_psm_malaria (document_id);

END;
$$;
