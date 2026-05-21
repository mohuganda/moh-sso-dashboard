-- =========================================================
-- document_templates
-- =========================================================

CREATE TABLE IF NOT EXISTS document_templates (
    id UUID PRIMARY KEY,

    -- Optional reference to the actual template file
    document_id UUID
        REFERENCES documents(id)
        ON DELETE SET NULL,

    -- Stable unique identifier
    code TEXT NOT NULL UNIQUE,

    -- Human readable name
    name TEXT NOT NULL,

    description TEXT,

    -- csv, xls, xlsx, json, xml
    file_type TEXT NOT NULL,

    -- Template versioning
    version INTEGER NOT NULL DEFAULT 1,

    -- Lifecycle
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    -- Dynamic template configuration
    configuration JSONB NOT NULL DEFAULT '{}',

    -- Auditing
    created_by UUID NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    archived_at TIMESTAMP
);

-- =========================================================
-- Indexes
-- =========================================================

CREATE INDEX idx_document_templates_document_id
    ON document_templates(document_id);

CREATE INDEX idx_document_templates_created_by
    ON document_templates(created_by);

CREATE INDEX idx_document_templates_is_active
    ON document_templates(is_active);

CREATE INDEX idx_document_templates_file_type
    ON document_templates(file_type);

CREATE INDEX idx_document_templates_configuration
    ON document_templates
    USING GIN(configuration);

-- Optional:
-- prevent duplicate active versions for same code
CREATE UNIQUE INDEX uq_document_templates_code_version
    ON document_templates(code, version);