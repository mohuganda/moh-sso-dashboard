-- 20260309_create_announcements.sql

CREATE TABLE announcements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    tag TEXT NOT NULL,
    priority INT DEFAULT 0,
    link_url TEXT,
    created_by UUID,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_announcements_created_at
ON announcements (created_at DESC);