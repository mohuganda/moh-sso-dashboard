CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY,
    original_filename TEXT NOT NULL,
    content_type TEXT,
    size_bytes BIGINT NOT NULL,
    checksum_sha256 TEXT,
    storage_location_id UUID NOT NULL REFERENCES storage_locations(id),
    object_key TEXT NOT NULL,               -- relative path inside base_uri
    uploaded_by UUID NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_documents_storage_location ON documents(storage_location_id);

CREATE INDEX idx_documents_uploaded_by ON documents(uploaded_by);