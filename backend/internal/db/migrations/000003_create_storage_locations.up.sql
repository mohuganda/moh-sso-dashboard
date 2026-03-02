CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS storage_locations (
    id UUID PRIMARY KEY,
    code TEXT UNIQUE NOT NULL,              -- e.g. "imports"
    name TEXT NOT NULL,                     -- display name
    provider TEXT NOT NULL,                 -- "s3" | "local"
    base_uri TEXT NOT NULL,                 -- e.g. s3://bucket/path
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO storage_locations (
    id,
    code,
    name,
    provider,
    base_uri,
    is_active
) VALUES (
    uuid_generate_v4(),
    'imports-local',
    'Local Import Storage',
    'local',
    'file:///var/app/storage/imports',
    TRUE
)
ON CONFLICT (code) DO NOTHING;


INSERT INTO storage_locations (
    id,
    code,
    name,
    provider,
    base_uri,
    is_active
) VALUES (
    uuid_generate_v4(),
    'imports-s3',
    'S3 Import Bucket',
    's3',
    's3://moh-imports-bucket/imports/',
    TRUE
)
ON CONFLICT (code) DO NOTHING;

INSERT INTO storage_locations (
    id,
    code,
    name,
    provider,
    base_uri,
    is_active
) VALUES (
    uuid_generate_v4(),
    'imports-minio',
    'MinIO Import Storage',
    'minio',
    's3://minio-bucket/imports/',
    TRUE
)
ON CONFLICT (code) DO NOTHING;

INSERT INTO storage_locations (
    id,
    code,
    name,
    provider,
    base_uri,
    is_active
) VALUES (
    uuid_generate_v4(),
    'imports-nfs',
    'NFS Import Storage',
    'nfs',
    'file:///mnt/nfs/imports',
    TRUE
)
ON CONFLICT (code) DO NOTHING;