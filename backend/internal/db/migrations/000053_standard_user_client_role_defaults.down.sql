DELETE FROM ihp_realm_role_system_roles WHERE realm_role = 'user';

INSERT INTO ihp_realm_role_system_roles (realm_role, system_role_id)
SELECT 'user', sr.id FROM ihp_standard_user_defaults_backup b
CROSS JOIN LATERAL jsonb_array_elements_text(b.snapshot->'roleIds') AS saved(role_id)
JOIN ihp_system_roles sr ON sr.id::text = saved.role_id
ON CONFLICT DO NOTHING;

DELETE FROM ihp_system_access_roles ar USING ihp_systems s
WHERE ar.system_id = s.id AND s.client_id = 'dashboard-web'
  AND ar.role_name = 'portal_user'
  AND EXISTS (SELECT 1 FROM ihp_standard_user_defaults_backup
      WHERE NOT (snapshot->>'portalUserAccess')::boolean);

DROP TABLE ihp_standard_user_defaults_backup;
