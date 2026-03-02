CREATE TABLE IF NOT EXISTS storage_locations (
    id UUID PRIMARY KEY,
    code TEXT UNIQUE NOT NULL,              -- e.g. "imports"
    name TEXT NOT NULL,                     -- display name
    provider TEXT NOT NULL,                 -- "s3" | "local"
    base_uri TEXT NOT NULL,                 -- e.g. s3://bucket/path
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);