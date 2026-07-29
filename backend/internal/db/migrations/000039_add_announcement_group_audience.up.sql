DO $$
BEGIN
    ALTER TYPE announcement_audience_type ADD VALUE IF NOT EXISTS 'SPECIFIC_GROUPS';
END $$;

CREATE TABLE IF NOT EXISTS announcement_groups (
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES ihp_rbac_groups(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (announcement_id, group_id)
);

CREATE INDEX IF NOT EXISTS idx_announcement_groups_group_id
    ON announcement_groups(group_id);
