-- =========================================================
-- document_template_columns
-- =========================================================

CREATE TYPE document_template_column_type AS ENUM (
    'STRING',
    'TEXT',
    'INTEGER',
    'DECIMAL',
    'BOOLEAN',
    'DATE',
    'DATETIME',
    'TIME',
    'ENUM',
    'UUID',
    'EMAIL',
    'PHONE',
    'JSON'
);

CREATE TABLE IF NOT EXISTS document_template_columns (
    id UUID PRIMARY KEY,

    sheet_id UUID NOT NULL
        REFERENCES document_template_sheets(id)
        ON DELETE CASCADE,

    -- Internal identifier
    column_key TEXT NOT NULL,

    -- Expected header name in uploaded file
    column_name TEXT NOT NULL,

    -- Human readable label
    display_name TEXT,

    -- Data typing
    data_type document_template_column_type NOT NULL,

    -- Validation rules
    required BOOLEAN NOT NULL DEFAULT FALSE,
    is_unique BOOLEAN NOT NULL DEFAULT FALSE,

    -- Parsing configuration
    column_order INTEGER,

    -- Default fallback value
    default_value TEXT,

    -- Flexible validation/configuration
    configuration JSONB NOT NULL DEFAULT '{}',

    -- Allowed enum values
    allowed_values JSONB NOT NULL DEFAULT '[]',

    -- Alternative accepted headers
    aliases JSONB NOT NULL DEFAULT '[]',

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    archived_at TIMESTAMP,

    CONSTRAINT uq_document_template_column_key
        UNIQUE(sheet_id, column_key),

    CONSTRAINT uq_document_template_column_name
        UNIQUE(sheet_id, column_name)
);

-- =========================================================
-- Indexes
-- =========================================================

CREATE INDEX idx_document_template_columns_sheet_id
    ON document_template_columns(sheet_id);

CREATE INDEX idx_document_template_columns_required
    ON document_template_columns(required);

CREATE INDEX idx_document_template_columns_data_type
    ON document_template_columns(data_type);

CREATE INDEX idx_document_template_columns_column_order
    ON document_template_columns(column_order);

CREATE INDEX idx_document_template_columns_configuration
    ON document_template_columns
    USING GIN(configuration);

CREATE INDEX idx_document_template_columns_allowed_values
    ON document_template_columns
    USING GIN(allowed_values);

CREATE INDEX idx_document_template_columns_aliases
    ON document_template_columns
    USING GIN(aliases);