DO $$
DECLARE
    legacy_system_id UUID;
    target_system_id UUID;
BEGIN
    SELECT id INTO legacy_system_id
    FROM ihp_systems
    WHERE client_id = 'integrated-outbreak-system';

    IF legacy_system_id IS NULL THEN
        RETURN;
    END IF;

    SELECT id INTO target_system_id
    FROM ihp_systems
    WHERE client_id = 'outbreak-management';

    IF target_system_id IS NULL THEN
        UPDATE ihp_systems
        SET client_id = 'outbreak-management',
            display_name = 'Outbreak Management',
            description = 'Signals, alerts, surveillance, points of entry, and case management.',
            icon = 'warning-alt',
            launch_url = '/portal/apps/outbreak-management',
            updated_at = now()
        WHERE id = legacy_system_id;

        UPDATE ihp_system_roles
        SET role_name = 'outbreak-management_access',
            display_name = 'Outbreak Management Access',
            updated_at = now()
        WHERE system_id = legacy_system_id
          AND role_name = 'integrated-outbreak-system_access';

        UPDATE ihp_system_access_roles
        SET role_name = 'outbreak-management_access'
        WHERE system_id = legacy_system_id
          AND role_name = 'integrated-outbreak-system_access';

        RETURN;
    END IF;

    INSERT INTO ihp_system_roles (
        system_id, role_name, display_name, description, enabled, metadata, created_at, updated_at
    )
    SELECT target_system_id,
           CASE
               WHEN role_name = 'integrated-outbreak-system_access' THEN 'outbreak-management_access'
               ELSE role_name
           END,
           CASE
               WHEN role_name = 'integrated-outbreak-system_access' THEN 'Outbreak Management Access'
               ELSE display_name
           END,
           description, enabled, metadata, created_at, now()
    FROM ihp_system_roles
    WHERE system_id = legacy_system_id
    ON CONFLICT (system_id, role_name) DO UPDATE SET
        description = COALESCE(EXCLUDED.description, ihp_system_roles.description),
        enabled = EXCLUDED.enabled,
        metadata = ihp_system_roles.metadata || EXCLUDED.metadata,
        updated_at = now();

    INSERT INTO ihp_system_role_permissions (system_role_id, permission_id)
    SELECT target_role.id, legacy_permission.permission_id
    FROM ihp_system_roles legacy_role
    JOIN ihp_system_role_permissions legacy_permission
      ON legacy_permission.system_role_id = legacy_role.id
    JOIN ihp_system_roles target_role
      ON target_role.system_id = target_system_id
     AND target_role.role_name = CASE
         WHEN legacy_role.role_name = 'integrated-outbreak-system_access' THEN 'outbreak-management_access'
         ELSE legacy_role.role_name
     END
    WHERE legacy_role.system_id = legacy_system_id
    ON CONFLICT DO NOTHING;

    INSERT INTO ihp_system_access_roles (system_id, role_name)
    SELECT target_system_id,
           CASE
               WHEN role_name = 'integrated-outbreak-system_access' THEN 'outbreak-management_access'
               ELSE role_name
           END
    FROM ihp_system_access_roles
    WHERE system_id = legacy_system_id
    ON CONFLICT DO NOTHING;

    INSERT INTO ihp_realm_role_system_roles (realm_role, system_role_id)
    SELECT mapping.realm_role, target_role.id
    FROM ihp_realm_role_system_roles mapping
    JOIN ihp_system_roles legacy_role ON legacy_role.id = mapping.system_role_id
    JOIN ihp_system_roles target_role
      ON target_role.system_id = target_system_id
     AND target_role.role_name = CASE
         WHEN legacy_role.role_name = 'integrated-outbreak-system_access' THEN 'outbreak-management_access'
         ELSE legacy_role.role_name
     END
    WHERE legacy_role.system_id = legacy_system_id
    ON CONFLICT DO NOTHING;

    DELETE FROM ihp_systems WHERE id = legacy_system_id;
END $$;

UPDATE ihp_rbac_audit_events
SET system_client_id = 'outbreak-management'
WHERE system_client_id = 'integrated-outbreak-system';

UPDATE ihp_access_requests
SET system_client_id = 'outbreak-management', updated_at = now()
WHERE system_client_id = 'integrated-outbreak-system';
