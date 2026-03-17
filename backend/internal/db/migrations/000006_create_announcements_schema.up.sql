-- 20260309_create_announcements.sql

-- Optional: for gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- -------------------------------------
-- Enums
-- -------------------------------------
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_type WHERE typname = 'announcement_status'
    ) THEN
        CREATE TYPE announcement_status AS ENUM (
            'DRAFT',
            'SCHEDULED',
            'PUBLISHED',
            'ARCHIVED'
        );
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_type WHERE typname = 'announcement_level'
    ) THEN
        CREATE TYPE announcement_level AS ENUM (
            'INFO',
            'SUCCESS',
            'WARNING',
            'CRITICAL'
        );
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_type WHERE typname = 'announcement_audience_type'
    ) THEN
        CREATE TYPE announcement_audience_type AS ENUM (
            'ALL_USERS',
            'ADMINS_ONLY',
            'SPECIFIC_CLIENTS',
            'SPECIFIC_ROLES',
            'SPECIFIC_USERS'
        );
    END IF;
END $$;

-- -------------------------------------
-- Main announcements table
-- -------------------------------------
CREATE TABLE IF NOT EXISTS announcements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Core content
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    summary VARCHAR(500),
    level announcement_level NOT NULL DEFAULT 'INFO',
    tag VARCHAR(100),
    link_url TEXT,

    -- Display / ordering
    priority INT NOT NULL DEFAULT 0,
    is_pinned BOOLEAN NOT NULL DEFAULT FALSE,

    -- Lifecycle
    status announcement_status NOT NULL DEFAULT 'DRAFT',
    publish_at TIMESTAMP NULL,
    expires_at TIMESTAMP NULL,

    -- Audience
    audience_type announcement_audience_type NOT NULL DEFAULT 'ALL_USERS',

    -- Ownership / audit
    created_by UUID NOT NULL,
    updated_by UUID,
    published_by UUID,
    archived_by UUID,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    published_at TIMESTAMP NULL,
    archived_at TIMESTAMP NULL,

    -- Soft delete
    deleted_at TIMESTAMP NULL,
    deleted_by UUID NULL,

    -- Versioning / concurrency
    version INT NOT NULL DEFAULT 1,

    -- Constraints
    CONSTRAINT chk_announcements_title_not_blank
        CHECK (btrim(title) <> ''),

    CONSTRAINT chk_announcements_message_not_blank
        CHECK (btrim(message) <> ''),

    CONSTRAINT chk_announcements_priority_non_negative
        CHECK (priority >= 0),

    CONSTRAINT chk_announcements_publish_expiry_order
        CHECK (
            expires_at IS NULL
            OR publish_at IS NULL
            OR expires_at > publish_at
        )
);

-- -------------------------------------
-- Audience mapping tables
-- -------------------------------------

-- For client-scoped announcements
CREATE TABLE IF NOT EXISTS announcement_clients (
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    client_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (announcement_id, client_id)
);

-- For role-scoped announcements
CREATE TABLE IF NOT EXISTS announcement_roles (
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    role_name VARCHAR(150) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (announcement_id, role_name)
);

-- For specific-user announcements
CREATE TABLE IF NOT EXISTS announcement_users (
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (announcement_id, user_id)
);

-- -------------------------------------
-- Helpful indexes
-- -------------------------------------

-- General sort / feed queries
CREATE INDEX IF NOT EXISTS idx_announcements_created_at
    ON announcements (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_announcements_status
    ON announcements (status);

CREATE INDEX IF NOT EXISTS idx_announcements_level
    ON announcements (level);

CREATE INDEX IF NOT EXISTS idx_announcements_priority_created_at
    ON announcements (priority DESC, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_announcements_publish_at
    ON announcements (publish_at);

CREATE INDEX IF NOT EXISTS idx_announcements_expires_at
    ON announcements (expires_at);

CREATE INDEX IF NOT EXISTS idx_announcements_is_pinned
    ON announcements (is_pinned);

-- Fast active published queries
CREATE INDEX IF NOT EXISTS idx_announcements_active_feed
    ON announcements (is_pinned DESC, priority DESC, publish_at DESC, created_at DESC)
    WHERE deleted_at IS NULL AND status = 'PUBLISHED';

-- Audience mapping indexes
CREATE INDEX IF NOT EXISTS idx_announcement_clients_client_id
    ON announcement_clients (client_id);

CREATE INDEX IF NOT EXISTS idx_announcement_roles_role_name
    ON announcement_roles (role_name);

CREATE INDEX IF NOT EXISTS idx_announcement_users_user_id
    ON announcement_users (user_id);

-- -------------------------------------
-- Trigger to auto-update updated_at + version
-- -------------------------------------
CREATE OR REPLACE FUNCTION set_announcements_updated_fields()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_announcements_set_updated_fields ON announcements;

CREATE TRIGGER trg_announcements_set_updated_fields
BEFORE UPDATE ON announcements
FOR EACH ROW
EXECUTE FUNCTION set_announcements_updated_fields();