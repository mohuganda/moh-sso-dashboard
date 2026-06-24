ALTER TABLE ihp_permissions
    ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active',
    ADD COLUMN IF NOT EXISTS deprecated_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS metadata JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE ihp_systems
    ADD COLUMN IF NOT EXISTS owner_team TEXT,
    ADD COLUMN IF NOT EXISTS owner_name TEXT,
    ADD COLUMN IF NOT EXISTS owner_email TEXT,
    ADD COLUMN IF NOT EXISTS support_url TEXT,
    ADD COLUMN IF NOT EXISTS documentation_url TEXT,
    ADD COLUMN IF NOT EXISTS environment TEXT,
    ADD COLUMN IF NOT EXISTS criticality TEXT;

CREATE TABLE IF NOT EXISTS ihp_rbac_audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id TEXT,
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id TEXT,
    system_client_id TEXT,
    role_name TEXT,
    permission_key TEXT,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS ihp_access_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT,
    username TEXT,
    email TEXT,
    system_client_id TEXT NOT NULL,
    requested_role TEXT NOT NULL,
    reason TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    requested_by TEXT,
    reviewed_by TEXT,
    reviewed_at TIMESTAMPTZ,
    decision_note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS ihp_access_request_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES ihp_access_requests(id) ON DELETE CASCADE,
    actor_user_id TEXT,
    action TEXT NOT NULL,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS ihp_rbac_change_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    requested_by TEXT,
    reviewed_by TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id TEXT,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    risk_level TEXT NOT NULL DEFAULT 'medium',
    reason TEXT,
    decision_note TEXT,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS ihp_rbac_change_request_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES ihp_rbac_change_requests(id) ON DELETE CASCADE,
    actor_user_id TEXT,
    action TEXT NOT NULL,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ihp_rbac_audit_events_created_at_idx ON ihp_rbac_audit_events(created_at DESC);
CREATE INDEX IF NOT EXISTS ihp_access_requests_status_idx ON ihp_access_requests(status);
CREATE INDEX IF NOT EXISTS ihp_rbac_change_requests_status_idx ON ihp_rbac_change_requests(status);
