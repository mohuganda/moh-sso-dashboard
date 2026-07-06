CREATE SCHEMA IF NOT EXISTS import;

-- Custom file uploads (legacy pattern, kept for existing integrations)
CREATE TABLE IF NOT EXISTS import.custom_files (
    file_key             SERIAL PRIMARY KEY,
    file_name            VARCHAR(255) NOT NULL UNIQUE,
    effective_start_date TIMESTAMP    NOT NULL,
    effective_end_date   TIMESTAMP    NOT NULL DEFAULT '2199-12-31 23:59:59.999',
    row_version          INT          NOT NULL DEFAULT 1,
    is_current           CHAR(1)      NOT NULL DEFAULT 'Y',
    create_date          TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    create_user          VARCHAR(50)           DEFAULT 'system',
    last_update_date     TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    last_update_user     VARCHAR(50)           DEFAULT 'system',
    file_path            TEXT
);

CREATE INDEX IF NOT EXISTS idx_custom_files_file_key        ON import.custom_files (file_key);
CREATE INDEX IF NOT EXISTS idx_custom_files_effective_dates ON import.custom_files (effective_start_date, effective_end_date);
CREATE INDEX IF NOT EXISTS idx_custom_files_is_current      ON import.custom_files (is_current);
CREATE INDEX IF NOT EXISTS idx_custom_files_row_version     ON import.custom_files (row_version);

-- Custom file row data (legacy pattern)
CREATE TABLE IF NOT EXISTS import.custom_data_files (
    data_file_key        SERIAL PRIMARY KEY,
    file_data            JSONB   NOT NULL,
    file_key             INT     NOT NULL,
    effective_start_date TIMESTAMP NOT NULL,
    effective_end_date   TIMESTAMP NOT NULL DEFAULT '2199-12-31 23:59:59.999',
    row_version          INT       NOT NULL DEFAULT 1,
    is_current           CHAR(1)   NOT NULL DEFAULT 'Y',
    create_date          TIMESTAMP          DEFAULT CURRENT_TIMESTAMP,
    create_user          VARCHAR(50)        DEFAULT 'system',
    last_update_date     TIMESTAMP          DEFAULT CURRENT_TIMESTAMP,
    last_update_user     VARCHAR(50)        DEFAULT 'system'
);

CREATE INDEX IF NOT EXISTS idx_custom_data_files_data_file_key   ON import.custom_data_files (data_file_key);
CREATE INDEX IF NOT EXISTS idx_custom_data_files_file_key        ON import.custom_data_files (file_key);
CREATE INDEX IF NOT EXISTS idx_custom_data_files_effective_dates ON import.custom_data_files (effective_start_date, effective_end_date);
CREATE INDEX IF NOT EXISTS idx_custom_data_files_is_current      ON import.custom_data_files (is_current);
CREATE INDEX IF NOT EXISTS idx_custom_data_files_row_version     ON import.custom_data_files (row_version);

-- Generic template upload records
CREATE TABLE IF NOT EXISTS import.template_uploads (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id   UUID NOT NULL,
    template_code TEXT NOT NULL,
    report_date   DATE,
    is_valid      BOOLEAN   NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    last_updated  TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_template_upload_document UNIQUE (document_id)
);

CREATE INDEX IF NOT EXISTS idx_template_uploads_document_id   ON import.template_uploads (document_id);
CREATE INDEX IF NOT EXISTS idx_template_uploads_template_code ON import.template_uploads (template_code);
CREATE INDEX IF NOT EXISTS idx_template_uploads_report_date   ON import.template_uploads (report_date);

-- Generic template row data
CREATE TABLE IF NOT EXISTS import.template_row_data (
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

CREATE INDEX IF NOT EXISTS idx_template_row_data_upload_id     ON import.template_row_data (upload_id);
CREATE INDEX IF NOT EXISTS idx_template_row_data_document_id   ON import.template_row_data (document_id);
CREATE INDEX IF NOT EXISTS idx_template_row_data_template_code ON import.template_row_data (template_code);
CREATE INDEX IF NOT EXISTS idx_template_row_data_sheet_code    ON import.template_row_data (template_code, sheet_code);
CREATE INDEX IF NOT EXISTS idx_template_row_data_report_date   ON import.template_row_data (report_date);
CREATE INDEX IF NOT EXISTS idx_template_row_data_is_valid      ON import.template_row_data (is_valid);
CREATE INDEX IF NOT EXISTS idx_template_row_data_gin           ON import.template_row_data USING GIN (data);
