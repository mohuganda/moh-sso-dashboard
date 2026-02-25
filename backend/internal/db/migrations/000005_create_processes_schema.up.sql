DO $$ BEGIN
    CREATE TYPE process_status AS ENUM (
        'PENDING',
        'PROCESSING',
        'COMPLETED',
        'FAILED',
        'CANCELLED'
    );
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS processes (
    id UUID PRIMARY KEY,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    process_type TEXT NOT NULL,             -- e.g. DOCUMENT_IMPORT
    status process_status NOT NULL DEFAULT 'PENDING',
    progress INT NOT NULL DEFAULT 0 CHECK (progress >= 0 AND progress <= 100),
    message TEXT,                           -- last progress message
    error TEXT,
    attempts INT NOT NULL DEFAULT 0,
    created_by UUID NOT NULL,
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_process_status ON processes(status);

CREATE INDEX idx_process_document ON processes(document_id);

CREATE INDEX idx_process_created_at ON processes(created_at);