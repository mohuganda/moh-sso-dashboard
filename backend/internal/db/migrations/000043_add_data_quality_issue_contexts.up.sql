CREATE TABLE data_quality_issue_contexts (
    issue_code TEXT PRIMARY KEY,
    health_context_id UUID NOT NULL REFERENCES health_context_nodes(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX data_quality_issue_contexts_health_context_idx
    ON data_quality_issue_contexts(health_context_id);

COMMENT ON TABLE data_quality_issue_contexts IS
    'Portal-owned health-context ownership for issues stored in the external DWH database.';
