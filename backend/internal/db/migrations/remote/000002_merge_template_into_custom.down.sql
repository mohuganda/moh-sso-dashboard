ALTER TABLE import.custom_data_files
    DROP CONSTRAINT IF EXISTS fk_custom_data_file_key,
    DROP CONSTRAINT IF EXISTS uq_custom_data_row_hash,
    DROP COLUMN IF EXISTS document_id,
    DROP COLUMN IF EXISTS template_code,
    DROP COLUMN IF EXISTS sheet_code,
    DROP COLUMN IF EXISTS row_number,
    DROP COLUMN IF EXISTS report_date,
    DROP COLUMN IF EXISTS raw_data,
    DROP COLUMN IF EXISTS row_hash;

ALTER TABLE import.custom_files
    DROP CONSTRAINT IF EXISTS uq_custom_file_document,
    DROP COLUMN IF EXISTS document_id,
    DROP COLUMN IF EXISTS template_code,
    DROP COLUMN IF EXISTS report_date;

CREATE TABLE import.template_uploads (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id   UUID NOT NULL,
    template_code TEXT NOT NULL,
    report_date   DATE,
    is_valid      BOOLEAN   NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    last_updated  TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_template_upload_document UNIQUE (document_id)
);

CREATE INDEX idx_template_uploads_document_id   ON import.template_uploads (document_id);
CREATE INDEX idx_template_uploads_template_code ON import.template_uploads (template_code);
CREATE INDEX idx_template_uploads_report_date   ON import.template_uploads (report_date);

CREATE TABLE import.template_row_data (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    upload_id     UUID    NOT NULL REFERENCES import.template_uploads (id),
    document_id   UUID    NOT NULL,
    template_code TEXT    NOT NULL,
    sheet_code    TEXT    NOT NULL,
    row_number    INT     NOT NULL,
    report_date   DATE,
    data          JSONB   NOT NULL DEFAULT '{}',
    raw_data      JSONB   NOT NULL DEFAULT '{}',
    row_hash      TEXT    NOT NULL,
    is_valid      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    last_updated  TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_template_row_hash UNIQUE (row_hash)
);

CREATE INDEX idx_template_row_data_upload_id     ON import.template_row_data (upload_id);
CREATE INDEX idx_template_row_data_document_id   ON import.template_row_data (document_id);
CREATE INDEX idx_template_row_data_template_code ON import.template_row_data (template_code);
CREATE INDEX idx_template_row_data_sheet_code    ON import.template_row_data (template_code, sheet_code);
CREATE INDEX idx_template_row_data_report_date   ON import.template_row_data (report_date);
CREATE INDEX idx_template_row_data_is_valid      ON import.template_row_data (is_valid);
CREATE INDEX idx_template_row_data_gin           ON import.template_row_data USING GIN (data);
