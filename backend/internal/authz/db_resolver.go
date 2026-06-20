package authz

import (
	"context"
	"database/sql"
	"strings"

	"github.com/lib/pq"
)

type DBResolver struct {
	db *sql.DB
}

func NewDBResolver(db *sql.DB) *DBResolver {
	return &DBResolver{db: db}
}

func (r *DBResolver) Resolve(
	ctx context.Context,
	realmRoles []string,
	clientRoles map[string][]string,
) (ResolvedAccess, error) {
	if r == nil || r.db == nil {
		return ResolvedAccess{}, sql.ErrConnDone
	}

	permissions := make([]Permission, 0)
	permissionSeen := map[Permission]bool{}

	addPermission := func(permission Permission) {
		if permission == "" || permissionSeen[permission] {
			return
		}
		permissionSeen[permission] = true
		permissions = append(permissions, permission)
	}

	realmPermissionKeys, err := r.listRealmPermissions(ctx, NormalizeRoles(realmRoles))
	if err != nil {
		return ResolvedAccess{}, err
	}
	for _, key := range realmPermissionKeys {
		if key == "*" {
			for _, permission := range AllPermissions {
				addPermission(permission)
			}
			continue
		}
		addPermission(Permission(key))
	}
	realmSystemPermissionKeys, err := r.listRealmSystemPermissions(ctx, NormalizeRoles(realmRoles))
	if err != nil {
		return ResolvedAccess{}, err
	}
	for _, key := range realmSystemPermissionKeys {
		addPermission(Permission(key))
	}

	systemSeen := map[string]bool{}
	systems := make([]SystemAccess, 0)
	realmSystems, err := r.listRealmAccessibleSystems(ctx, NormalizeRoles(realmRoles))
	if err != nil {
		return ResolvedAccess{}, err
	}
	for _, system := range realmSystems {
		if system.ClientID == "" || systemSeen[system.ClientID] {
			continue
		}
		systemSeen[system.ClientID] = true
		systems = append(systems, system)
	}

	for clientID, roles := range normalizeClientRoles(clientRoles) {
		if len(roles) == 0 {
			continue
		}

		systemPermissionKeys, err := r.listSystemPermissions(ctx, clientID, roles)
		if err != nil {
			return ResolvedAccess{}, err
		}
		for _, key := range systemPermissionKeys {
			if key == "*" {
				for _, permission := range AllPermissions {
					addPermission(permission)
				}
				continue
			}
			addPermission(Permission(key))
		}

		accessibleSystems, err := r.listAccessibleSystems(ctx, clientID, roles)
		if err != nil {
			return ResolvedAccess{}, err
		}
		for _, system := range accessibleSystems {
			if system.ClientID == "" || systemSeen[system.ClientID] {
				continue
			}
			systemSeen[system.ClientID] = true
			systems = append(systems, system)
		}
	}

	return ResolvedAccess{
		Permissions: permissions,
		Systems:     systems,
	}, nil
}

func (r *DBResolver) listRealmSystemPermissions(ctx context.Context, realmRoles []string) ([]string, error) {
	if len(realmRoles) == 0 {
		return nil, nil
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT p.permission_key
		FROM ihp_realm_role_system_roles rrsr
		JOIN ihp_system_roles sr ON sr.id = rrsr.system_role_id
		JOIN ihp_systems s ON s.id = sr.system_id
		JOIN ihp_system_role_permissions srp ON srp.system_role_id = sr.id
		JOIN ihp_permissions p ON p.id = srp.permission_id
		WHERE rrsr.realm_role = ANY($1::text[])
		  AND s.enabled = TRUE
		  AND sr.enabled = TRUE
		ORDER BY p.permission_key ASC
	`, pq.Array(realmRoles))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanStrings(rows)
}

func (r *DBResolver) listRealmAccessibleSystems(ctx context.Context, realmRoles []string) ([]SystemAccess, error) {
	if len(realmRoles) == 0 {
		return nil, nil
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			s.client_id,
			s.display_name,
			COALESCE(s.launch_url, '') AS launch_url,
			COALESCE(s.icon, '') AS icon,
			COALESCE(s.category, '') AS category,
			COALESCE(s.metadata->>'navigation', '') AS navigation,
			s.system_type,
			s.display_in_launcher,
			s.display_in_sidenav,
			s.launch_mode,
			ARRAY_AGG(DISTINCT sr.role_name ORDER BY sr.role_name)::text[] AS roles
		FROM ihp_realm_role_system_roles rrsr
		JOIN ihp_system_roles sr ON sr.id = rrsr.system_role_id
		JOIN ihp_systems s ON s.id = sr.system_id
		JOIN ihp_system_access_roles sar
		  ON sar.system_id = s.id AND sar.role_name = sr.role_name
		WHERE rrsr.realm_role = ANY($1::text[])
		  AND s.enabled = TRUE
		  AND sr.enabled = TRUE
		GROUP BY s.id, s.client_id, s.display_name, s.launch_url, s.icon, s.category, s.metadata,
		         s.system_type, s.display_in_launcher, s.display_in_sidenav, s.launch_mode
		ORDER BY s.display_name ASC
	`, pq.Array(realmRoles))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	systems := make([]SystemAccess, 0)
	for rows.Next() {
		var system SystemAccess
		if err := rows.Scan(
			&system.ClientID,
			&system.DisplayName,
			&system.LaunchURL,
			&system.Icon,
			&system.Category,
			&system.Navigation,
			&system.SystemType,
			&system.DisplayInLauncher,
			&system.DisplayInSideNav,
			&system.LaunchMode,
			pq.Array(&system.Roles),
		); err != nil {
			return nil, err
		}
		systems = append(systems, system)
	}
	return systems, rows.Err()
}

func (r *DBResolver) listRealmPermissions(ctx context.Context, roles []string) ([]string, error) {
	if len(roles) == 0 {
		return nil, nil
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT p.permission_key
		FROM ihp_realm_role_permissions rrp
		JOIN ihp_permissions p ON p.id = rrp.permission_id
		WHERE rrp.realm_role = ANY($1::text[])
		ORDER BY p.permission_key ASC
	`, pq.Array(roles))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanStrings(rows)
}

func (r *DBResolver) listSystemPermissions(
	ctx context.Context,
	clientID string,
	roles []string,
) ([]string, error) {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" || len(roles) == 0 {
		return nil, nil
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT p.permission_key
		FROM ihp_system_role_permissions srp
		JOIN ihp_permissions p ON p.id = srp.permission_id
		JOIN ihp_system_roles sr ON sr.id = srp.system_role_id
		JOIN ihp_systems s ON s.id = sr.system_id
		WHERE s.client_id = $1
		  AND sr.role_name = ANY($2::text[])
		  AND s.enabled = TRUE
		  AND sr.enabled = TRUE
		ORDER BY p.permission_key ASC
	`, clientID, pq.Array(roles))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanStrings(rows)
}

func (r *DBResolver) listAccessibleSystems(
	ctx context.Context,
	clientID string,
	roles []string,
) ([]SystemAccess, error) {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" || len(roles) == 0 {
		return nil, nil
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT
			s.client_id,
			s.display_name,
			COALESCE(s.launch_url, '') AS launch_url,
			COALESCE(s.icon, '') AS icon,
			COALESCE(s.category, '') AS category,
			COALESCE(s.metadata->>'navigation', '') AS navigation,
			s.system_type,
			s.display_in_launcher,
			s.display_in_sidenav,
			s.launch_mode,
			ARRAY(
				SELECT DISTINCT ar.role_name
				FROM ihp_system_access_roles ar
				WHERE ar.system_id = s.id
				  AND ar.role_name = ANY($2::text[])
				ORDER BY ar.role_name ASC
			)::text[] AS roles
		FROM ihp_systems s
		JOIN ihp_system_access_roles ar ON ar.system_id = s.id
		WHERE s.client_id = $1
		  AND s.enabled = TRUE
		  AND ar.role_name = ANY($2::text[])
		ORDER BY s.display_name ASC
	`, clientID, pq.Array(roles))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	systems := make([]SystemAccess, 0)
	for rows.Next() {
		var system SystemAccess
		if err := rows.Scan(
			&system.ClientID,
			&system.DisplayName,
			&system.LaunchURL,
			&system.Icon,
			&system.Category,
			&system.Navigation,
			&system.SystemType,
			&system.DisplayInLauncher,
			&system.DisplayInSideNav,
			&system.LaunchMode,
			pq.Array(&system.Roles),
		); err != nil {
			return nil, err
		}
		systems = append(systems, system)
	}

	return systems, rows.Err()
}

func scanStrings(rows *sql.Rows) ([]string, error) {
	values := make([]string, 0)
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	return values, rows.Err()
}
