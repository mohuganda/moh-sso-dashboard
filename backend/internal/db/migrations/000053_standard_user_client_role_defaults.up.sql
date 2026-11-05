-- Replace only the realm-user default mappings. Direct user roles, group grants,
-- administrator/manager mappings, and the definitions of each role are preserved.
CREATE TABLE IF NOT EXISTS ihp_standard_user_defaults_backup (
    id BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (id),
    snapshot JSONB NOT NULL
);

INSERT INTO ihp_standard_user_defaults_backup (id, snapshot)
SELECT TRUE, jsonb_build_object(
    'roleIds', COALESCE((SELECT jsonb_agg(system_role_id::text)
        FROM ihp_realm_role_system_roles WHERE realm_role = 'user'), '[]'::jsonb),
    'portalUserAccess', EXISTS (SELECT 1 FROM ihp_system_access_roles ar
        JOIN ihp_systems s ON s.id = ar.system_id
        WHERE s.client_id = 'dashboard-web' AND ar.role_name = 'portal_user')
)
ON CONFLICT (id) DO NOTHING;

DELETE FROM ihp_realm_role_system_roles WHERE realm_role = 'user';

WITH defaults (client_id, role_name) AS (VALUES
    ('dashboard-web', 'dashboard-web_access'),
    ('dashboard-web', 'portal_user'),
    ('data-statistics', 'report-scheduler_access'),
    ('data-statistics', 'report_viewer'),
    ('data-statistics', 'data-statistics_access'),
    ('data-statistics', 'issue_tracker_viewer'),
    ('data-statistics', 'document_viewer'),
    ('settings', 'settings_access')
)
INSERT INTO ihp_realm_role_system_roles (realm_role, system_role_id)
SELECT 'user', sr.id FROM defaults d
JOIN ihp_systems s ON s.client_id = d.client_id
JOIN ihp_system_roles sr ON sr.system_id = s.id AND sr.role_name = d.role_name
ON CONFLICT DO NOTHING;

INSERT INTO ihp_system_access_roles (system_id, role_name)
SELECT id, 'portal_user' FROM ihp_systems WHERE client_id = 'dashboard-web'
ON CONFLICT DO NOTHING;
