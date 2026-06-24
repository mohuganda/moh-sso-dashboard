-- name: UpsertIHPSystem :one
INSERT INTO ihp_systems (
    client_id,
    display_name,
    description,
    icon,
    launch_url,
    category,
    enabled,
    sort_order,
    metadata
) VALUES (
    sqlc.arg(client_id),
    sqlc.arg(display_name),
    sqlc.arg(description),
    sqlc.arg(icon),
    sqlc.arg(launch_url),
    sqlc.arg(category),
    sqlc.arg(enabled),
    sqlc.arg(sort_order),
    sqlc.arg(metadata)
)
ON CONFLICT (client_id) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    icon = EXCLUDED.icon,
    launch_url = EXCLUDED.launch_url,
    category = EXCLUDED.category,
    enabled = EXCLUDED.enabled,
    sort_order = EXCLUDED.sort_order,
    metadata = EXCLUDED.metadata,
    updated_at = now()
RETURNING *;

-- name: ListIHPSystems :many
SELECT *
FROM ihp_systems
ORDER BY sort_order ASC, display_name ASC;

-- name: ListEnabledIHPSystems :many
SELECT *
FROM ihp_systems
WHERE enabled = TRUE
ORDER BY sort_order ASC, display_name ASC;

-- name: GetIHPSystemByClientID :one
SELECT *
FROM ihp_systems
WHERE client_id = $1;

-- name: UpsertIHPSystemRole :one
INSERT INTO ihp_system_roles (
    system_id,
    role_name,
    display_name,
    description,
    enabled,
    metadata
) VALUES (
    sqlc.arg(system_id),
    sqlc.arg(role_name),
    sqlc.arg(display_name),
    sqlc.arg(description),
    sqlc.arg(enabled),
    sqlc.arg(metadata)
)
ON CONFLICT (system_id, role_name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    enabled = EXCLUDED.enabled,
    metadata = EXCLUDED.metadata,
    updated_at = now()
RETURNING *;

-- name: ListIHPSystemRoles :many
SELECT sr.*
FROM ihp_system_roles sr
JOIN ihp_systems s ON s.id = sr.system_id
WHERE s.client_id = $1
ORDER BY sr.role_name ASC;

-- name: UpsertIHPPermission :one
INSERT INTO ihp_permissions (
    permission_key,
    display_name,
    description,
    category
) VALUES (
    sqlc.arg(permission_key),
    sqlc.arg(display_name),
    sqlc.arg(description),
    sqlc.arg(category)
)
ON CONFLICT (permission_key) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    updated_at = now()
RETURNING *;

-- name: ListIHPPermissions :many
SELECT *
FROM ihp_permissions
ORDER BY permission_key ASC;

-- name: AssignIHPPermissionToSystemRole :exec
INSERT INTO ihp_system_role_permissions (system_role_id, permission_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveIHPPermissionFromSystemRole :exec
DELETE FROM ihp_system_role_permissions
WHERE system_role_id = $1
  AND permission_id = $2;

-- name: AssignIHPPermissionToRealmRole :exec
INSERT INTO ihp_realm_role_permissions (realm_role, permission_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveIHPPermissionFromRealmRole :exec
DELETE FROM ihp_realm_role_permissions
WHERE realm_role = $1
  AND permission_id = $2;

-- name: AssignIHPSystemAccessRole :exec
INSERT INTO ihp_system_access_roles (system_id, role_name)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveIHPSystemAccessRole :exec
DELETE FROM ihp_system_access_roles
WHERE system_id = $1
  AND role_name = $2;

-- name: ListIHPPermissionsForRealmRoles :many
SELECT DISTINCT p.permission_key
FROM ihp_realm_role_permissions rrp
JOIN ihp_permissions p ON p.id = rrp.permission_id
WHERE rrp.realm_role = ANY(sqlc.arg(realm_roles)::text[])
ORDER BY p.permission_key ASC;

-- name: ListIHPPermissionsForSystemRoles :many
SELECT DISTINCT p.permission_key
FROM ihp_system_role_permissions srp
JOIN ihp_permissions p ON p.id = srp.permission_id
JOIN ihp_system_roles sr ON sr.id = srp.system_role_id
JOIN ihp_systems s ON s.id = sr.system_id
WHERE s.client_id = sqlc.arg(client_id)
  AND sr.role_name = ANY(sqlc.arg(role_names)::text[])
  AND s.enabled = TRUE
  AND sr.enabled = TRUE
ORDER BY p.permission_key ASC;

-- name: ListIHPAccessibleSystemsForClientRoles :many
SELECT DISTINCT
    s.client_id,
    s.display_name,
    s.launch_url,
    s.icon,
    s.category,
    ARRAY(
        SELECT DISTINCT ar.role_name
        FROM ihp_system_access_roles ar
        WHERE ar.system_id = s.id
          AND ar.role_name = ANY(sqlc.arg(role_names)::text[])
        ORDER BY ar.role_name ASC
    )::text[] AS roles
FROM ihp_systems s
JOIN ihp_system_access_roles ar ON ar.system_id = s.id
WHERE s.client_id = sqlc.arg(client_id)
  AND s.enabled = TRUE
  AND ar.role_name = ANY(sqlc.arg(role_names)::text[])
ORDER BY s.display_name ASC;

