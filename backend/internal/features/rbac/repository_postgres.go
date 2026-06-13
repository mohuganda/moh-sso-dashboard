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
		       COALESCE(launch_url, ''), COALESCE(category, ''), COALESCE(owner_team, ''),
		       COALESCE(owner_name, ''), COALESCE(owner_email, ''), COALESCE(support_url, ''),
		       COALESCE(documentation_url, ''), COALESCE(environment, ''), COALESCE(criticality, ''),
		       enabled, sort_order
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
			&system.OwnerTeam,
			&system.OwnerName,
			&system.OwnerEmail,
			&system.SupportURL,
			&system.DocumentationURL,
			&system.Environment,
			&system.Criticality,
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
		       COALESCE(launch_url, ''), COALESCE(category, ''), COALESCE(owner_team, ''),
		       COALESCE(owner_name, ''), COALESCE(owner_email, ''), COALESCE(support_url, ''),
		       COALESCE(documentation_url, ''), COALESCE(environment, ''), COALESCE(criticality, ''),
		       enabled, sort_order
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
		&detail.OwnerTeam,
		&detail.OwnerName,
		&detail.OwnerEmail,
		&detail.SupportURL,
		&detail.DocumentationURL,
		&detail.Environment,
		&detail.Criticality,
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
			client_id, display_name, description, icon, launch_url, category, owner_team, owner_name,
			owner_email, support_url, documentation_url, environment, criticality, enabled, sort_order
		) VALUES (
			$1, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''),
			NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''), NULLIF($11, ''), NULLIF($12, ''),
			NULLIF($13, ''), $14, $15
		)
		ON CONFLICT (client_id) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			description = EXCLUDED.description,
			icon = EXCLUDED.icon,
			launch_url = EXCLUDED.launch_url,
			category = EXCLUDED.category,
			owner_team = EXCLUDED.owner_team,
			owner_name = EXCLUDED.owner_name,
			owner_email = EXCLUDED.owner_email,
			support_url = EXCLUDED.support_url,
			documentation_url = EXCLUDED.documentation_url,
			environment = EXCLUDED.environment,
			criticality = EXCLUDED.criticality,
			enabled = EXCLUDED.enabled,
			sort_order = EXCLUDED.sort_order,
			updated_at = now()
		RETURNING id::text, client_id, display_name, COALESCE(description, ''), COALESCE(icon, ''),
		          COALESCE(launch_url, ''), COALESCE(category, ''), COALESCE(owner_team, ''),
		          COALESCE(owner_name, ''), COALESCE(owner_email, ''), COALESCE(support_url, ''),
		          COALESCE(documentation_url, ''), COALESCE(environment, ''), COALESCE(criticality, ''),
		          enabled, sort_order
	`, input.ClientID, input.DisplayName, input.Description, input.Icon, input.LaunchURL, input.Category, input.OwnerTeam, input.OwnerName, input.OwnerEmail, input.SupportURL, input.DocumentationURL, input.Environment, input.Criticality, enabled, input.SortOrder).Scan(
		&system.ID,
		&system.ClientID,
		&system.DisplayName,
		&system.Description,
		&system.Icon,
		&system.LaunchURL,
		&system.Category,
		&system.OwnerTeam,
		&system.OwnerName,
		&system.OwnerEmail,
		&system.SupportURL,
		&system.DocumentationURL,
		&system.Environment,
		&system.Criticality,
		&system.Enabled,
		&system.SortOrder,
	)
	return system, err
}

func (r *postgresRepository) ListPermissions(ctx context.Context) ([]Permission, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, permission_key, COALESCE(display_name, ''), COALESCE(description, ''), COALESCE(category, ''), COALESCE(status, 'active')
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

func (r *postgresRepository) GetSystemRole(ctx context.Context, roleID string) (SystemRole, error) {
	var role SystemRole
	err := r.db.QueryRowContext(ctx, `
		SELECT sr.id::text, sr.system_id::text, s.client_id, sr.role_name,
		       COALESCE(sr.display_name, ''), COALESCE(sr.description, ''), sr.enabled
		FROM ihp_system_roles sr
		JOIN ihp_systems s ON s.id = sr.system_id
		WHERE sr.id = $1::uuid
	`, roleID).Scan(
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

func (r *postgresRepository) UpsertSystemRole(ctx context.Context, clientID string, input RoleInput) (SystemRole, error) {
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
		ON CONFLICT (system_id, role_name) DO UPDATE SET
			display_name = COALESCE(EXCLUDED.display_name, ihp_system_roles.display_name),
			description = COALESCE(EXCLUDED.description, ihp_system_roles.description),
			enabled = EXCLUDED.enabled,
			updated_at = now()
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
	if err != nil {
		return SystemRole{}, err
	}
	role.Permissions, err = r.listPermissionsForSystemRole(ctx, role.ID)
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

func (r *postgresRepository) ListAuditEvents(ctx context.Context, limit int) ([]AuditEvent, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, COALESCE(actor_user_id, ''), action, resource_type, COALESCE(resource_id, ''),
		       COALESCE(system_client_id, ''), COALESCE(role_name, ''), COALESCE(permission_key, ''),
		       details::text, created_at::text
		FROM ihp_rbac_audit_events
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]AuditEvent, 0)
	for rows.Next() {
		var event AuditEvent
		if err := rows.Scan(
			&event.ID,
			&event.ActorUserID,
			&event.Action,
			&event.ResourceType,
			&event.ResourceID,
			&event.SystemClientID,
			&event.RoleName,
			&event.PermissionKey,
			&event.Details,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (r *postgresRepository) RecordAuditEvent(ctx context.Context, event AuditEvent) error {
	details := string(event.Details)
	if strings.TrimSpace(details) == "" {
		details = "{}"
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO ihp_rbac_audit_events (
			actor_user_id, action, resource_type, resource_id, system_client_id, role_name, permission_key, details
		) VALUES (NULLIF($1, ''), $2, $3, NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), $8::jsonb)
	`, event.ActorUserID, event.Action, event.ResourceType, event.ResourceID, event.SystemClientID, event.RoleName, event.PermissionKey, details)
	return err
}

func (r *postgresRepository) CreateAccessRequest(ctx context.Context, input AccessRequestInput) (AccessRequest, error) {
	var request AccessRequest
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO ihp_access_requests (
			user_id, username, email, system_client_id, requested_role, reason, requested_by
		) VALUES (NULLIF($1, ''), NULLIF($2, ''), NULLIF($3, ''), $4, $5, NULLIF($6, ''), NULLIF($1, ''))
		RETURNING id::text, COALESCE(user_id, ''), COALESCE(username, ''), COALESCE(email, ''), system_client_id,
		          requested_role, COALESCE(reason, ''), status, COALESCE(requested_by, ''), COALESCE(reviewed_by, ''),
		          COALESCE(reviewed_at::text, ''), COALESCE(decision_note, ''), created_at::text, updated_at::text
	`, input.UserID, input.Username, input.Email, input.SystemClientID, input.RequestedRole, input.Reason).Scan(
		&request.ID,
		&request.UserID,
		&request.Username,
		&request.Email,
		&request.SystemClientID,
		&request.RequestedRole,
		&request.Reason,
		&request.Status,
		&request.RequestedBy,
		&request.ReviewedBy,
		&request.ReviewedAt,
		&request.DecisionNote,
		&request.CreatedAt,
		&request.UpdatedAt,
	)
	return request, err
}

func (r *postgresRepository) ListAccessRequests(ctx context.Context) ([]AccessRequest, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, COALESCE(user_id, ''), COALESCE(username, ''), COALESCE(email, ''), system_client_id,
		       requested_role, COALESCE(reason, ''), status, COALESCE(requested_by, ''), COALESCE(reviewed_by, ''),
		       COALESCE(reviewed_at::text, ''), COALESCE(decision_note, ''), created_at::text, updated_at::text
		FROM ihp_access_requests
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	requests := make([]AccessRequest, 0)
	for rows.Next() {
		request, err := scanAccessRequest(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, rows.Err()
}

func (r *postgresRepository) UpdateAccessRequestStatus(ctx context.Context, id string, status string, note string, reviewer string) (AccessRequest, error) {
	var request AccessRequest
	err := r.db.QueryRowContext(ctx, `
		UPDATE ihp_access_requests
		SET status = $2,
		    decision_note = NULLIF($3, ''),
		    reviewed_by = NULLIF($4, ''),
		    reviewed_at = now(),
		    updated_at = now()
		WHERE id = $1::uuid
		RETURNING id::text, COALESCE(user_id, ''), COALESCE(username, ''), COALESCE(email, ''), system_client_id,
		          requested_role, COALESCE(reason, ''), status, COALESCE(requested_by, ''), COALESCE(reviewed_by, ''),
		          COALESCE(reviewed_at::text, ''), COALESCE(decision_note, ''), created_at::text, updated_at::text
	`, id, status, note, reviewer).Scan(
		&request.ID,
		&request.UserID,
		&request.Username,
		&request.Email,
		&request.SystemClientID,
		&request.RequestedRole,
		&request.Reason,
		&request.Status,
		&request.RequestedBy,
		&request.ReviewedBy,
		&request.ReviewedAt,
		&request.DecisionNote,
		&request.CreatedAt,
		&request.UpdatedAt,
	)
	return request, err
}

func (r *postgresRepository) CreateChangeRequest(ctx context.Context, input ChangeRequestInput) (ChangeRequest, error) {
	payload := string(input.Payload)
	if strings.TrimSpace(payload) == "" {
		payload = "{}"
	}
	var request ChangeRequest
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO ihp_rbac_change_requests (
			action, resource_type, resource_id, payload, risk_level, reason
		) VALUES ($1, $2, NULLIF($3, ''), $4::jsonb, COALESCE(NULLIF($5, ''), 'medium'), NULLIF($6, ''))
		RETURNING id::text, COALESCE(requested_by, ''), COALESCE(reviewed_by, ''), status, action, resource_type,
		          COALESCE(resource_id, ''), payload::text, risk_level, COALESCE(reason, ''), COALESCE(decision_note, ''),
		          COALESCE(reviewed_at::text, ''), created_at::text, updated_at::text
	`, input.Action, input.ResourceType, input.ResourceID, payload, input.RiskLevel, input.Reason).Scan(
		&request.ID,
		&request.RequestedBy,
		&request.ReviewedBy,
		&request.Status,
		&request.Action,
		&request.ResourceType,
		&request.ResourceID,
		&request.Payload,
		&request.RiskLevel,
		&request.Reason,
		&request.DecisionNote,
		&request.ReviewedAt,
		&request.CreatedAt,
		&request.UpdatedAt,
	)
	return request, err
}

func (r *postgresRepository) ListChangeRequests(ctx context.Context) ([]ChangeRequest, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, COALESCE(requested_by, ''), COALESCE(reviewed_by, ''), status, action, resource_type,
		       COALESCE(resource_id, ''), payload::text, risk_level, COALESCE(reason, ''), COALESCE(decision_note, ''),
		       COALESCE(reviewed_at::text, ''), created_at::text, updated_at::text
		FROM ihp_rbac_change_requests
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	requests := make([]ChangeRequest, 0)
	for rows.Next() {
		request, err := scanChangeRequest(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, rows.Err()
}

func (r *postgresRepository) UpdateChangeRequestStatus(ctx context.Context, id string, status string, note string, reviewer string) (ChangeRequest, error) {
	var request ChangeRequest
	err := r.db.QueryRowContext(ctx, `
		UPDATE ihp_rbac_change_requests
		SET status = $2,
		    decision_note = NULLIF($3, ''),
		    reviewed_by = NULLIF($4, ''),
		    reviewed_at = now(),
		    updated_at = now()
		WHERE id = $1::uuid
		RETURNING id::text, COALESCE(requested_by, ''), COALESCE(reviewed_by, ''), status, action, resource_type,
		          COALESCE(resource_id, ''), payload::text, risk_level, COALESCE(reason, ''), COALESCE(decision_note, ''),
		          COALESCE(reviewed_at::text, ''), created_at::text, updated_at::text
	`, id, status, note, reviewer).Scan(
		&request.ID,
		&request.RequestedBy,
		&request.ReviewedBy,
		&request.Status,
		&request.Action,
		&request.ResourceType,
		&request.ResourceID,
		&request.Payload,
		&request.RiskLevel,
		&request.Reason,
		&request.DecisionNote,
		&request.ReviewedAt,
		&request.CreatedAt,
		&request.UpdatedAt,
	)
	return request, err
}

func (r *postgresRepository) UpdatePermissionMetadata(ctx context.Context, permissionKey string, input PermissionMetadataInput) (Permission, error) {
	var permission Permission
	err := r.db.QueryRowContext(ctx, `
		UPDATE ihp_permissions
		SET display_name = NULLIF($2, ''),
		    description = NULLIF($3, ''),
		    category = NULLIF($4, ''),
		    status = COALESCE(NULLIF($5, ''), 'active'),
		    updated_at = now()
		WHERE permission_key = $1
		RETURNING id::text, permission_key, COALESCE(display_name, ''), COALESCE(description, ''), COALESCE(category, ''), COALESCE(status, 'active')
	`, permissionKey, input.DisplayName, input.Description, input.Category, input.Status).Scan(
		&permission.ID,
		&permission.Key,
		&permission.DisplayName,
		&permission.Description,
		&permission.Category,
		&permission.Status,
	)
	return permission, err
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
		SELECT p.id::text, p.permission_key, COALESCE(p.display_name, ''), COALESCE(p.description, ''), COALESCE(p.category, ''), COALESCE(p.status, 'active')
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
		SELECT p.id::text, p.permission_key, COALESCE(p.display_name, ''), COALESCE(p.description, ''), COALESCE(p.category, ''), COALESCE(p.status, 'active')
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

type accessRequestScanner interface {
	Scan(dest ...any) error
}

type changeRequestScanner interface {
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
		&permission.Status,
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

func scanAccessRequest(row accessRequestScanner) (AccessRequest, error) {
	var request AccessRequest
	err := row.Scan(
		&request.ID,
		&request.UserID,
		&request.Username,
		&request.Email,
		&request.SystemClientID,
		&request.RequestedRole,
		&request.Reason,
		&request.Status,
		&request.RequestedBy,
		&request.ReviewedBy,
		&request.ReviewedAt,
		&request.DecisionNote,
		&request.CreatedAt,
		&request.UpdatedAt,
	)
	return request, err
}

func scanChangeRequest(row changeRequestScanner) (ChangeRequest, error) {
	var request ChangeRequest
	err := row.Scan(
		&request.ID,
		&request.RequestedBy,
		&request.ReviewedBy,
		&request.Status,
		&request.Action,
		&request.ResourceType,
		&request.ResourceID,
		&request.Payload,
		&request.RiskLevel,
		&request.Reason,
		&request.DecisionNote,
		&request.ReviewedAt,
		&request.CreatedAt,
		&request.UpdatedAt,
	)
	return request, err
}

func normalizeRoleName(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}

func notFoundError(entity string) error {
	return fmt.Errorf("%s not found: %w", entity, sql.ErrNoRows)
}
