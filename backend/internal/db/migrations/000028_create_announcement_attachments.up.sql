CREATE TABLE IF NOT EXISTS announcement_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    file_name TEXT NOT NULL,
    original_file_name TEXT NOT NULL,
    content_type TEXT,
    file_size BIGINT NOT NULL,
    storage_provider TEXT NOT NULL,
    storage_key TEXT NOT NULL,
    checksum TEXT,
    uploaded_by UUID REFERENCES users(id) ON DELETE SET NULL,
    include_in_email BOOLEAN NOT NULL DEFAULT TRUE,
    inline BOOLEAN NOT NULL DEFAULT FALSE,
    content_id TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    deleted_by UUID REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_announcement_attachments_announcement_id
    ON announcement_attachments(announcement_id);

CREATE INDEX IF NOT EXISTS idx_announcement_attachments_announcement_deleted
    ON announcement_attachments(announcement_id, deleted_at);

CREATE UNIQUE INDEX IF NOT EXISTS idx_announcement_attachments_storage_key_unique
    ON announcement_attachments(storage_key);
