package rbac

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type postgresRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) ListSystems(ctx context.Context) ([]System, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, client_id, display_name, COALESCE(description, ''), COALESCE(icon, ''),
		       COALESCE(launch_url, ''), COALESCE(category, ''), enabled, sort_order
		FROM ihp_systems
		ORDER BY sort_order, display_name, client_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	systems := make([]System, 0)
	for rows.Next() {
		var system System
		if err := rows.Scan(
			&system.ID,
			&system.ClientID,
			&system.DisplayName,
			&system.Description,
			&system.Icon,
			&system.LaunchURL,
			&system.Category,
			&system.Enabled,
			&system.SortOrder,
		); err != nil {
			return nil, err
		}
		systems = append(systems, system)
	}
	return systems, rows.Err()
}

func (r *postgresRepository) GetSystem(ctx context.Context, clientID string) (SystemDetail, error) {
	var detail SystemDetail
	err := r.db.QueryRowContext(ctx, `
		SELECT id::text, client_id, display_name, COALESCE(description, ''), COALESCE(icon, ''),
		       COALESCE(launch_url, ''), COALESCE(category, ''), enabled, sort_order
		FROM ihp_systems
		WHERE client_id = $1
	`, clientID).Scan(
		&detail.ID,
		&detail.ClientID,
		&detail.DisplayName,
		&detail.Description,
		&detail.Icon,
		&detail.LaunchURL,
		&detail.Category,
		&detail.Enabled,
		&detail.SortOrder,
	)
	if err != nil {
		return SystemDetail{}, err
	}

	accessRoles, err := r.listSystemAccessRoles(ctx, clientID)
	if err != nil {
		return SystemDetail{}, err
	}
	roles, err := r.ListSystemRoles(ctx, clientID)
	if err != nil {
		return SystemDetail{}, err
	}
	detail.AccessRoles = accessRoles
	detail.Roles = roles
	return detail, nil
}

func (r *postgresRepository) UpsertSystem(ctx context.Context, input UpsertSystemInput) (System, error) {
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	var system System
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO ihp_systems (
			client_id, display_name, description, icon, launch_url, category, enabled, sort_order
		) VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), $7, $8)
		ON CONFLICT (client_id) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			description = EXCLUDED.description,
			icon = EXCLUDED.icon,
			launch_url = EXCLUDED.launch_url,
			category = EXCLUDED.category,
			enabled = EXCLUDED.enabled,
			sort_order = EXCLUDED.sort_order,
			updated_at = now()
		RETURNING id::text, client_id, display_name, COALESCE(description, ''), COALESCE(icon, ''),
		          COALESCE(launch_url, ''), COALESCE(category, ''), enabled, sort_order
	`, input.ClientID, input.DisplayName, input.Description, input.Icon, input.LaunchURL, input.Category, enabled, input.SortOrder).Scan(
		&system.ID,
		&system.ClientID,
		&system.DisplayName,
		&system.Description,
		&system.Icon,
		&system.LaunchURL,
		&system.Category,
		&system.Enabled,
		&system.SortOrder,
	)
	return system, err
}

func (r *postgresRepository) ListPermissions(ctx context.Context) ([]Permission, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, permission_key, COALESCE(display_name, ''), COALESCE(description, ''), COALESCE(category, '')
		FROM ihp_permissions
		ORDER BY category NULLS LAST, permission_key
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := make([]Permission, 0)
	for rows.Next() {
		permission, err := scanPermission(rows)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, rows.Err()
}

func (r *postgresRepository) ListSystemRoles(ctx context.Context, clientID string) ([]SystemRole, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT sr.id::text, sr.system_id::text, s.client_id, sr.role_name,
		       COALESCE(sr.display_name, ''), COALESCE(sr.description, ''), sr.enabled
		FROM ihp_system_roles sr
		JOIN ihp_systems s ON s.id = sr.system_id
		WHERE s.client_id = $1
		ORDER BY sr.role_name
	`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := make([]SystemRole, 0)
	for rows.Next() {
		var role SystemRole
		if err := rows.Scan(
			&role.ID,
			&role.SystemID,
			&role.ClientID,
			&role.Name,
			&role.DisplayName,
			&role.Description,
			&role.Enabled,
		); err != nil {
			return nil, err
		}
		permissions, err := r.listPermissionsForSystemRole(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		role.Permissions = permissions
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *postgresRepository) CreateSystemRole(ctx context.Context, clientID string, input RoleInput) (SystemRole, error) {
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	var role SystemRole
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO ihp_system_roles (system_id, role_name, display_name, description, enabled)
		SELECT id, $2, NULLIF($3, ''), NULLIF($4, ''), $5
		FROM ihp_systems
		WHERE client_id = $1
		RETURNING id::text, system_id::text, $1::text, role_name, COALESCE(display_name, ''), COALESCE(description, ''), enabled
	`, clientID, input.Name, input.DisplayName, input.Description, enabled).Scan(
		&role.ID,
		&role.SystemID,
		&role.ClientID,
		&role.Name,
		&role.DisplayName,
		&role.Description,
		&role.Enabled,
	)
	return role, err
}

func (r *postgresRepository) UpdateSystemRole(ctx context.Context, roleID string, input RoleInput) (SystemRole, error) {
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	var role SystemRole
	err := r.db.QueryRowContext(ctx, `
		UPDATE ihp_system_roles sr
		SET role_name = $2,
		    display_name = NULLIF($3, ''),
		    description = NULLIF($4, ''),
		    enabled = $5,
		    updated_at = now()
		FROM ihp_systems s
		WHERE sr.system_id = s.id AND sr.id = $1::uuid
		RETURNING sr.id::text, sr.system_id::text, s.client_id, sr.role_name,
		          COALESCE(sr.display_name, ''), COALESCE(sr.description, ''), sr.enabled
	`, roleID, input.Name, input.DisplayName, input.Description, enabled).Scan(
		&role.ID,
		&role.SystemID,
		&role.ClientID,
		&role.Name,
		&role.DisplayName,
		&role.Description,
		&role.Enabled,
	)
	if err != nil {
		return SystemRole{}, err
	}
	role.Permissions, err = r.listPermissionsForSystemRole(ctx, role.ID)
	return role, err
}

func (r *postgresRepository) DeleteSystemRole(ctx context.Context, roleID string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM ihp_system_roles WHERE id = $1::uuid`, roleID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err == nil && affected == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (r *postgresRepository) AssignSystemRolePermission(ctx context.Context, roleID string, permissionKey string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO ihp_system_role_permissions (system_role_id, permission_id)
		SELECT $1::uuid, id
		FROM ihp_permissions
		WHERE permission_key = $2
		ON CONFLICT DO NOTHING
	`, roleID, permissionKey)
	return err
}

func (r *postgresRepository) RemoveSystemRolePermission(ctx context.Context, roleID string, permissionKey string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM ihp_system_role_permissions srp
		USING ihp_permissions p
		WHERE srp.permission_id = p.id
		  AND srp.system_role_id = $1::uuid
		  AND p.permission_key = $2
	`, roleID, permissionKey)
	return err
}

func (r *postgresRepository) ListRealmRolePermissions(ctx context.Context) ([]RealmRolePermissionGroup, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT realm_role
		FROM ihp_realm_role_permissions
		GROUP BY realm_role
		ORDER BY realm_role
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]RealmRolePermissionGroup, 0)
	for rows.Next() {
		var group RealmRolePermissionGroup
		if err := rows.Scan(&group.RealmRole); err != nil {
			return nil, err
		}
		permissions, err := r.listPermissionsForRealmRole(ctx, group.RealmRole)
		if err != nil {
			return nil, err
		}
		group.Permissions = permissions
		groups = append(groups, group)
	}
	return groups, rows.Err()
}

func (r *postgresRepository) AssignRealmRolePermission(ctx context.Context, realmRole string, permissionKey string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO ihp_realm_role_permissions (realm_role, permission_id)
		SELECT $1, id
		FROM ihp_permissions
		WHERE permission_key = $2
		ON CONFLICT DO NOTHING
	`, realmRole, permissionKey)
	return err
}

func (r *postgresRepository) RemoveRealmRolePermission(ctx context.Context, realmRole string, permissionKey string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM ihp_realm_role_permissions rrp
		USING ihp_permissions p
		WHERE rrp.permission_id = p.id
		  AND rrp.realm_role = $1
		  AND p.permission_key = $2
	`, realmRole, permissionKey)
	return err
}

func (r *postgresRepository) AddSystemAccessRole(ctx context.Context, clientID string, roleName string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO ihp_system_access_roles (system_id, role_name)
		SELECT id, $2
		FROM ihp_systems
		WHERE client_id = $1
		ON CONFLICT DO NOTHING
	`, clientID, roleName)
	return err
}

func (r *postgresRepository) RemoveSystemAccessRole(ctx context.Context, clientID string, roleName string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM ihp_system_access_roles sar
		USING ihp_systems s
		WHERE sar.system_id = s.id
		  AND s.client_id = $1
		  AND sar.role_name = $2
	`, clientID, roleName)
	return err
}

func (r *postgresRepository) CountSystemAccessRoles(ctx context.Context, clientID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM ihp_system_access_roles sar
		JOIN ihp_systems s ON s.id = sar.system_id
		WHERE s.client_id = $1
	`, clientID).Scan(&count)
	return count, err
}

func (r *postgresRepository) PermissionExists(ctx context.Context, permissionKey string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM ihp_permissions WHERE permission_key = $1)`, permissionKey).Scan(&exists)
	return exists, err
}

func (r *postgresRepository) listSystemAccessRoles(ctx context.Context, clientID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT sar.role_name
		FROM ihp_system_access_roles sar
		JOIN ihp_systems s ON s.id = sar.system_id
		WHERE s.client_id = $1
		ORDER BY sar.role_name
	`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := make([]string, 0)
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *postgresRepository) listPermissionsForSystemRole(ctx context.Context, roleID string) ([]Permission, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id::text, p.permission_key, COALESCE(p.display_name, ''), COALESCE(p.description, ''), COALESCE(p.category, '')
		FROM ihp_system_role_permissions srp
		JOIN ihp_permissions p ON p.id = srp.permission_id
		WHERE srp.system_role_id = $1::uuid
		ORDER BY p.permission_key
	`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPermissions(rows)
}

func (r *postgresRepository) listPermissionsForRealmRole(ctx context.Context, realmRole string) ([]Permission, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id::text, p.permission_key, COALESCE(p.display_name, ''), COALESCE(p.description, ''), COALESCE(p.category, '')
		FROM ihp_realm_role_permissions rrp
		JOIN ihp_permissions p ON p.id = rrp.permission_id
		WHERE rrp.realm_role = $1
		ORDER BY p.permission_key
	`, realmRole)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPermissions(rows)
}

type permissionScanner interface {
	Scan(dest ...any) error
}

func scanPermission(row permissionScanner) (Permission, error) {
	var permission Permission
	err := row.Scan(
		&permission.ID,
		&permission.Key,
		&permission.DisplayName,
		&permission.Description,
		&permission.Category,
	)
	return permission, err
}

func scanPermissions(rows *sql.Rows) ([]Permission, error) {
	permissions := make([]Permission, 0)
	for rows.Next() {
		permission, err := scanPermission(rows)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, rows.Err()
}

func normalizeRoleName(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}

func notFoundError(entity string) error {
	return fmt.Errorf("%s not found: %w", entity, sql.ErrNoRows)
}
