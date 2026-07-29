DO $$
BEGIN
    ALTER TYPE announcement_audience_type
        ADD VALUE IF NOT EXISTS 'SPECIFIC_HEALTH_CONTEXTS';
    ALTER TYPE announcement_audience_type
        ADD VALUE IF NOT EXISTS 'HEALTH_CONTEXT_AND_DESCENDANTS';
END $$;

CREATE TABLE announcement_health_contexts (
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    context_node_id UUID NOT NULL REFERENCES health_context_nodes(id) ON DELETE RESTRICT,
    include_descendants BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (announcement_id, context_node_id)
);

CREATE INDEX announcement_health_contexts_context_idx
    ON announcement_health_contexts(context_node_id);

CREATE TABLE announcement_audience_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    channel TEXT NOT NULL,
    recipient_count INTEGER NOT NULL CHECK (recipient_count >= 0),
    recipient_user_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX announcement_audience_snapshots_announcement_idx
    ON announcement_audience_snapshots(announcement_id, created_at DESC);
