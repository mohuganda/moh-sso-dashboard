-- =========================================================
-- document_template_sheets
-- =========================================================

CREATE TABLE IF NOT EXISTS document_template_sheets (
    id UUID PRIMARY KEY,

    template_id UUID NOT NULL
        REFERENCES document_templates(id)
        ON DELETE CASCADE,

    -- Internal sheet identifier
    code TEXT NOT NULL,

    -- Expected workbook sheet name
    name TEXT NOT NULL,

    -- Human readable label
    display_name TEXT,

    -- Sheet requirements
    required BOOLEAN NOT NULL DEFAULT TRUE,

    -- Workbook positioning
    sheet_order INTEGER,

    -- Parsing configuration
    header_row INTEGER NOT NULL DEFAULT 1,
    start_row INTEGER NOT NULL DEFAULT 2,

    -- Validation behavior
    allow_extra_columns BOOLEAN NOT NULL DEFAULT FALSE,
    allow_duplicate_headers BOOLEAN NOT NULL DEFAULT FALSE,

    -- Flexible sheet configuration
    configuration JSONB NOT NULL DEFAULT '{}',

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    archived_at TIMESTAMP,

    CONSTRAINT uq_document_template_sheet_code
        UNIQUE(template_id, code),

    CONSTRAINT uq_document_template_sheet_name
        UNIQUE(template_id, name)
);

-- =========================================================
-- Indexes
-- =========================================================

CREATE INDEX idx_document_template_sheets_template_id
    ON document_template_sheets(template_id);

CREATE INDEX idx_document_template_sheets_required
    ON document_template_sheets(required);

CREATE INDEX idx_document_template_sheets_sheet_order
    ON document_template_sheets(sheet_order);

CREATE INDEX idx_document_template_sheets_configuration
    ON document_template_sheets
    USING GIN(configuration);