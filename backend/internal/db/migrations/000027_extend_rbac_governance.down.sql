DROP TABLE IF EXISTS ihp_rbac_change_request_events;
DROP TABLE IF EXISTS ihp_rbac_change_requests;
DROP TABLE IF EXISTS ihp_access_request_events;
DROP TABLE IF EXISTS ihp_access_requests;
DROP TABLE IF EXISTS ihp_rbac_audit_events;

ALTER TABLE ihp_systems
    DROP COLUMN IF EXISTS criticality,
    DROP COLUMN IF EXISTS environment,
    DROP COLUMN IF EXISTS documentation_url,
    DROP COLUMN IF EXISTS support_url,
    DROP COLUMN IF EXISTS owner_email,
    DROP COLUMN IF EXISTS owner_name,
    DROP COLUMN IF EXISTS owner_team;

ALTER TABLE ihp_permissions
    DROP COLUMN IF EXISTS metadata,
    DROP COLUMN IF EXISTS deprecated_at,
    DROP COLUMN IF EXISTS status;
