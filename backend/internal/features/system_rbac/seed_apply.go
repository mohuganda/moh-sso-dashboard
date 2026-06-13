package system_rbac

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/moh-sso-dashboard/internal/authz"
)

func ApplySeed(ctx context.Context, db *sql.DB, seed SeedFile) error {
	if db == nil {
		return fmt.Errorf("database is required")
	}
	if err := ValidateSeed(seed); err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	permissionIDs := map[string]string{}
	for _, permission := range collectPermissions(seed) {
		if permission == "*" {
			continue
		}

		permissionID, err := upsertPermission(ctx, tx, permission)
		if err != nil {
			return err
		}
		permissionIDs[permission] = permissionID
	}

	for _, system := range seed.Systems {
		systemID, err := upsertSystem(ctx, tx, system)
		if err != nil {
			return err
		}

		for _, accessRole := range system.AccessRoles {
			if err := assignSystemAccessRole(ctx, tx, systemID, accessRole); err != nil {
				return err
			}
		}

		for _, role := range system.Roles {
			roleID, err := upsertSystemRole(ctx, tx, systemID, role)
			if err != nil {
				return err
			}

			for _, permission := range role.Permissions {
				if permission == "*" {
					continue
				}
				if err := assignSystemRolePermission(ctx, tx, roleID, permissionIDs[permission]); err != nil {
					return err
				}
			}
		}
	}

	for _, role := range seed.RealmRoles {
		realmRole := strings.ToLower(strings.TrimSpace(role.Name))
		for _, permission := range role.Permissions {
			if permission == "*" {
				for _, permissionID := range permissionIDs {
					if err := assignRealmRolePermission(ctx, tx, realmRole, permissionID); err != nil {
						return err
					}
				}
				continue
			}
			if err := assignRealmRolePermission(ctx, tx, realmRole, permissionIDs[permission]); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func collectPermissions(seed SeedFile) []string {
	seen := map[string]bool{}
	values := make([]string, 0)

	add := func(permission string) {
		permission = strings.TrimSpace(permission)
		if permission == "" || seen[permission] {
			return
		}
		seen[permission] = true
		values = append(values, permission)
	}
	addAll := func() {
		for _, permission := range authz.AllPermissions {
			add(string(permission))
		}
	}

	for _, system := range seed.Systems {
		for _, role := range system.Roles {
			for _, permission := range role.Permissions {
				if permission == "*" {
					addAll()
					continue
				}
				add(permission)
			}
		}
	}
	for _, role := range seed.RealmRoles {
		for _, permission := range role.Permissions {
			if permission == "*" {
				addAll()
				continue
			}
			add(permission)
		}
	}

	return values
}

func upsertPermission(ctx context.Context, tx *sql.Tx, permission string) (string, error) {
	var id string
	err := tx.QueryRowContext(ctx, `
		INSERT INTO ihp_permissions (permission_key)
		VALUES ($1)
		ON CONFLICT (permission_key) DO UPDATE SET updated_at = now()
		RETURNING id::text
	`, permission).Scan(&id)

	return id, err
}

func upsertSystem(ctx context.Context, tx *sql.Tx, system SeedSystem) (string, error) {
	enabled := true
	if system.Enabled != nil {
		enabled = *system.Enabled
	}

	metadata, err := json.Marshal(map[string]any{})
	if err != nil {
		return "", err
	}

	var id string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO ihp_systems (
			client_id, display_name, description, icon, launch_url, category, enabled, metadata
		) VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), $7, $8)
		ON CONFLICT (client_id) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			description = EXCLUDED.description,
			icon = EXCLUDED.icon,
			launch_url = EXCLUDED.launch_url,
			category = EXCLUDED.category,
			enabled = EXCLUDED.enabled,
			metadata = EXCLUDED.metadata,
			updated_at = now()
		RETURNING id::text
	`, system.ClientID, system.DisplayName, system.Description, system.Icon, system.LaunchURL, system.Category, enabled, metadata).Scan(&id)

	return id, err
}

func upsertSystemRole(ctx context.Context, tx *sql.Tx, systemID string, role SeedRole) (string, error) {
	metadata, err := json.Marshal(map[string]any{})
	if err != nil {
		return "", err
	}

	var id string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO ihp_system_roles (
			system_id, role_name, display_name, description, enabled, metadata
		) VALUES ($1::uuid, $2, NULLIF($3, ''), NULLIF($4, ''), TRUE, $5)
		ON CONFLICT (system_id, role_name) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			description = EXCLUDED.description,
			enabled = EXCLUDED.enabled,
			metadata = EXCLUDED.metadata,
			updated_at = now()
		RETURNING id::text
	`, systemID, strings.ToLower(strings.TrimSpace(role.Name)), role.DisplayName, role.Description, metadata).Scan(&id)

	return id, err
}

func assignSystemAccessRole(ctx context.Context, tx *sql.Tx, systemID string, roleName string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO ihp_system_access_roles (system_id, role_name)
		VALUES ($1::uuid, $2)
		ON CONFLICT DO NOTHING
	`, systemID, strings.ToLower(strings.TrimSpace(roleName)))

	return err
}

func assignSystemRolePermission(ctx context.Context, tx *sql.Tx, roleID string, permissionID string) error {
	if permissionID == "" {
		return nil
	}

	_, err := tx.ExecContext(ctx, `
		INSERT INTO ihp_system_role_permissions (system_role_id, permission_id)
		VALUES ($1::uuid, $2::uuid)
		ON CONFLICT DO NOTHING
	`, roleID, permissionID)

	return err
}

func assignRealmRolePermission(ctx context.Context, tx *sql.Tx, realmRole string, permissionID string) error {
	if permissionID == "" {
		return nil
	}

	_, err := tx.ExecContext(ctx, `
		INSERT INTO ihp_realm_role_permissions (realm_role, permission_id)
		VALUES ($1, $2::uuid)
		ON CONFLICT DO NOTHING
	`, realmRole, permissionID)

	return err
}
