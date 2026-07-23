-- Consolidates the two parallel import storage models (template_uploads /
-- template_row_data for Excel, custom_files / custom_data_files for CSV)
-- into a single pair of tables. The original custom_files / custom_data_files
-- columns are preserved as-is (existing external pipelines read these by
-- name) — this migration only adds the columns needed to also support
-- template-driven (Excel) imports.
DROP TABLE IF EXISTS import.template_row_data;
DROP TABLE IF EXISTS import.template_uploads;

ALTER TABLE import.custom_files
    ADD COLUMN document_id   UUID,
    ADD COLUMN template_code TEXT,
    ADD COLUMN report_date   DATE,
    ADD CONSTRAINT uq_custom_file_document UNIQUE (document_id);

CREATE INDEX idx_custom_files_document_id   ON import.custom_files (document_id);
CREATE INDEX idx_custom_files_template_code ON import.custom_files (template_code);

ALTER TABLE import.custom_data_files
    ADD COLUMN document_id   UUID,
    ADD COLUMN template_code TEXT,
    ADD COLUMN sheet_code    TEXT,
    ADD COLUMN row_number    INT,
    ADD COLUMN report_date   DATE,
    ADD COLUMN raw_data      JSONB,
    ADD COLUMN row_hash      TEXT,
    ADD CONSTRAINT uq_custom_data_row_hash UNIQUE (row_hash),
    ADD CONSTRAINT fk_custom_data_file_key FOREIGN KEY (file_key) REFERENCES import.custom_files (file_key);

CREATE INDEX idx_custom_data_files_document_id   ON import.custom_data_files (document_id);
CREATE INDEX idx_custom_data_files_template_code ON import.custom_data_files (template_code);
CREATE INDEX idx_custom_data_files_sheet_code    ON import.custom_data_files (template_code, sheet_code);
CREATE INDEX idx_custom_data_files_gin           ON import.custom_data_files USING GIN (file_data);
