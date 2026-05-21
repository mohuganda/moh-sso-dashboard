CREATE TYPE document_status AS ENUM (
    'UPLOADED',
    'PROCESSING',
    'READY',
    'ARCHIVED',
    'DELETED',
    'FAILED'
);

CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY,

    original_filename TEXT NOT NULL,
    normalized_filename TEXT,

    content_type TEXT,
    extension TEXT,

    size_bytes BIGINT NOT NULL,

    checksum_sha256 TEXT,

    storage_location_id UUID NOT NULL
        REFERENCES storage_locations(id),

    object_key TEXT NOT NULL,

    uploaded_by UUID NOT NULL,

    status document_status NOT NULL DEFAULT 'UPLOADED',

    version INTEGER NOT NULL DEFAULT 1,

    parent_document_id UUID
        REFERENCES documents(id),

    metadata JSONB NOT NULL DEFAULT '{}',

    tags JSONB NOT NULL DEFAULT '[]',

    is_template BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    deleted_at TIMESTAMP
);

CREATE INDEX idx_documents_storage_location ON documents(storage_location_id);
CREATE INDEX idx_documents_uploaded_by ON documents(uploaded_by);
CREATE INDEX idx_documents_status ON documents(status);