package health_context

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type postgresRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) CreateNode(ctx context.Context, input CreateNodeInput) (Node, error) {
	if r.db == nil {
		return Node{}, errors.New("health context database is unavailable")
	}

	metadata := input.Metadata
	if len(metadata) == 0 {
		metadata = json.RawMessage(`{}`)
	}

	row := r.db.QueryRowContext(ctx, `
		INSERT INTO health_context_nodes (
			code, name, context_type, parent_id, source, metadata, enabled
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, code, name, context_type, parent_id, source, metadata,
		          enabled, version, created_at, updated_at
	`, input.Code, input.Name, input.ContextType, input.ParentID, input.Source, metadata, input.Enabled)

	return scanNode(row)
}

func (r *postgresRepository) UpdateNode(ctx context.Context, input UpdateNodeInput) (Node, error) {
	if r.db == nil {
		return Node{}, errors.New("health context database is unavailable")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Node{}, err
	}
	defer tx.Rollback()

	var currentParent uuid.NullUUID
	var currentVersion int
	if err := tx.QueryRowContext(ctx, `
		SELECT parent_id, version
		FROM health_context_nodes
		WHERE id = $1
		FOR UPDATE
	`, input.ID).Scan(&currentParent, &currentVersion); errors.Is(err, sql.ErrNoRows) {
		return Node{}, ErrNotFound
	} else if err != nil {
		return Node{}, err
	}
	if currentVersion != input.Version {
		return Node{}, ErrVersionConflict
	}

	parentChanged := currentParent.Valid != (input.ParentID != nil)
	if currentParent.Valid && input.ParentID != nil {
		parentChanged = currentParent.UUID != *input.ParentID
	}
	if parentChanged {
		if _, err := tx.ExecContext(ctx, `
			DELETE FROM health_context_closure
			WHERE descendant_id IN (
				SELECT descendant_id
				FROM health_context_closure
				WHERE ancestor_id = $1
			)
			AND ancestor_id IN (
				SELECT ancestor_id
				FROM health_context_closure
				WHERE descendant_id = $1 AND ancestor_id <> $1
			)
		`, input.ID); err != nil {
			return Node{}, err
		}

		if input.ParentID != nil {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO health_context_closure (ancestor_id, descendant_id, depth)
				SELECT parent_tree.ancestor_id,
				       child_tree.descendant_id,
				       parent_tree.depth + child_tree.depth + 1
				FROM health_context_closure parent_tree
				CROSS JOIN health_context_closure child_tree
				WHERE parent_tree.descendant_id = $1
				  AND child_tree.ancestor_id = $2
			`, input.ParentID, input.ID); err != nil {
				return Node{}, err
			}
		}
	}

	metadata := input.Metadata
	if len(metadata) == 0 {
		metadata = json.RawMessage(`{}`)
	}
	node, err := scanNode(tx.QueryRowContext(ctx, `
		UPDATE health_context_nodes
		SET code = $2,
		    name = $3,
		    context_type = $4,
		    parent_id = $5,
		    source = $6,
		    metadata = $7,
		    enabled = $8,
		    version = version + 1,
		    updated_at = NOW()
		WHERE id = $1 AND version = $9
		RETURNING id, code, name, context_type, parent_id, source, metadata,
		          enabled, version, created_at, updated_at
	`, input.ID, input.Code, input.Name, input.ContextType, input.ParentID,
		input.Source, metadata, input.Enabled, input.Version))
	if err != nil {
		return Node{}, err
	}
	if err := tx.Commit(); err != nil {
		return Node{}, err
	}
	return node, nil
}

func (r *postgresRepository) DeleteNode(ctx context.Context, id uuid.UUID) error {
	if r.db == nil {
		return errors.New("health context database is unavailable")
	}

	var references int
	if err := r.db.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM health_context_nodes WHERE parent_id = $1) +
			(SELECT COUNT(*) FROM user_health_context_assignments WHERE context_node_id = $1) +
			(SELECT COUNT(*) FROM group_health_context_assignments WHERE context_node_id = $1)
	`, id).Scan(&references); err != nil {
		return err
	}
	if references > 0 {
		return ErrContextInUse
	}

	result, err := r.db.ExecContext(ctx, `DELETE FROM health_context_nodes WHERE id = $1`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *postgresRepository) GetNode(ctx context.Context, id uuid.UUID) (Node, error) {
	if r.db == nil {
		return Node{}, errors.New("health context database is unavailable")
	}

	return scanNode(r.db.QueryRowContext(ctx, `
		SELECT id, code, name, context_type, parent_id, source, metadata,
		       enabled, version, created_at, updated_at
		FROM health_context_nodes
		WHERE id = $1
	`, id))
}

func (r *postgresRepository) ListNodes(
	ctx context.Context,
	parentID *uuid.UUID,
	contextType ContextType,
	includeDisabled bool,
) ([]Node, error) {
	if r.db == nil {
		return nil, errors.New("health context database is unavailable")
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, code, name, context_type, parent_id, source, metadata,
		       enabled, version, created_at, updated_at
		FROM health_context_nodes
		WHERE ($1::uuid IS NULL OR parent_id = $1)
		  AND ($2 = '' OR context_type = $2)
		  AND ($3 OR enabled)
		ORDER BY context_type, name, code
	`, parentID, contextType, includeDisabled)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanNodes(rows)
}

func (r *postgresRepository) ListAncestors(ctx context.Context, id uuid.UUID) ([]Node, error) {
	return r.listClosure(ctx, id, true)
}

func (r *postgresRepository) ListDescendants(ctx context.Context, id uuid.UUID) ([]Node, error) {
	return r.listClosure(ctx, id, false)
}

func (r *postgresRepository) listClosure(ctx context.Context, id uuid.UUID, ancestors bool) ([]Node, error) {
	if r.db == nil {
		return nil, errors.New("health context database is unavailable")
	}

	joinColumn := "hc.descendant_id"
	nodeColumn := "hc.ancestor_id"
	if !ancestors {
		joinColumn = "hc.ancestor_id"
		nodeColumn = "hc.descendant_id"
	}

	query := fmt.Sprintf(`
		SELECT n.id, n.code, n.name, n.context_type, n.parent_id, n.source,
		       n.metadata, n.enabled, n.version, n.created_at, n.updated_at
		FROM health_context_closure hc
		JOIN health_context_nodes n ON n.id = %s
		WHERE %s = $1
		ORDER BY hc.depth, n.name
	`, nodeColumn, joinColumn)

	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanNodes(rows)
}

func (r *postgresRepository) IsDescendant(
	ctx context.Context,
	ancestorID uuid.UUID,
	descendantID uuid.UUID,
) (bool, error) {
	if r.db == nil {
		return false, errors.New("health context database is unavailable")
	}

	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM health_context_closure
			WHERE ancestor_id = $1 AND descendant_id = $2
		)
	`, ancestorID, descendantID).Scan(&exists)
	return exists, err
}

func (r *postgresRepository) UpsertAlias(ctx context.Context, input UpsertAliasInput) (Alias, error) {
	if r.db == nil {
		return Alias{}, errors.New("health context database is unavailable")
	}

	metadata := input.Metadata
	if len(metadata) == 0 {
		metadata = json.RawMessage(`{}`)
	}

	var alias Alias
	var rawMetadata []byte
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO health_context_aliases (
			context_node_id, namespace, external_id, metadata
		) VALUES ($1, $2, $3, $4)
		ON CONFLICT (context_node_id, namespace) DO UPDATE
		SET external_id = EXCLUDED.external_id,
		    metadata = EXCLUDED.metadata
		RETURNING id, context_node_id, namespace, external_id, metadata, created_at
	`, input.ContextNodeID, input.Namespace, input.ExternalID, metadata).Scan(
		&alias.ID, &alias.ContextNodeID, &alias.Namespace, &alias.ExternalID,
		&rawMetadata, &alias.CreatedAt,
	)
	if err != nil {
		return Alias{}, err
	}
	alias.Metadata = json.RawMessage(rawMetadata)
	return alias, nil
}

func (r *postgresRepository) GetNodeByAlias(
	ctx context.Context,
	namespace string,
	externalID string,
) (Node, error) {
	if r.db == nil {
		return Node{}, errors.New("health context database is unavailable")
	}

	return scanNode(r.db.QueryRowContext(ctx, `
		SELECT n.id, n.code, n.name, n.context_type, n.parent_id, n.source,
		       n.metadata, n.enabled, n.version, n.created_at, n.updated_at
		FROM health_context_aliases alias
		JOIN health_context_nodes n ON n.id = alias.context_node_id
		WHERE alias.namespace = $1 AND alias.external_id = $2
	`, namespace, externalID))
}

func (r *postgresRepository) ListAliases(ctx context.Context, contextID uuid.UUID) ([]Alias, error) {
	if r.db == nil {
		return nil, errors.New("health context database is unavailable")
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, context_node_id, namespace, external_id, metadata, created_at
		FROM health_context_aliases
		WHERE context_node_id = $1
		ORDER BY namespace
	`, contextID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Alias, 0)
	for rows.Next() {
		var alias Alias
		var metadata []byte
		if err := rows.Scan(
			&alias.ID, &alias.ContextNodeID, &alias.Namespace, &alias.ExternalID,
			&metadata, &alias.CreatedAt,
		); err != nil {
			return nil, err
		}
		alias.Metadata = json.RawMessage(metadata)
		result = append(result, alias)
	}
	return result, rows.Err()
}

func (r *postgresRepository) ListAliasesInScope(
	ctx context.Context,
	contextID uuid.UUID,
	includeDescendants bool,
) ([]Alias, error) {
	if r.db == nil {
		return nil, errors.New("health context database is unavailable")
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT alias.id, alias.context_node_id, alias.namespace, alias.external_id,
		       alias.metadata, alias.created_at
		FROM health_context_aliases alias
		JOIN health_context_nodes node ON node.id = alias.context_node_id
		JOIN health_context_closure closure
		  ON closure.descendant_id = alias.context_node_id
		WHERE closure.ancestor_id = $1
		  AND ($2 OR closure.depth = 0)
		  AND node.enabled
		ORDER BY closure.depth, alias.namespace, alias.external_id
	`, contextID, includeDescendants)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Alias, 0)
	for rows.Next() {
		var alias Alias
		var metadata []byte
		if err := rows.Scan(
			&alias.ID,
			&alias.ContextNodeID,
			&alias.Namespace,
			&alias.ExternalID,
			&metadata,
			&alias.CreatedAt,
		); err != nil {
			return nil, err
		}
		alias.Metadata = json.RawMessage(metadata)
		result = append(result, alias)
	}
	return result, rows.Err()
}

func (r *postgresRepository) DeleteAlias(
	ctx context.Context,
	contextID uuid.UUID,
	aliasID uuid.UUID,
) error {
	if r.db == nil {
		return errors.New("health context database is unavailable")
	}

	result, err := r.db.ExecContext(ctx, `
		DELETE FROM health_context_aliases
		WHERE id = $1 AND context_node_id = $2
	`, aliasID, contextID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *postgresRepository) ListUserAssignments(ctx context.Context, userID string) ([]Assignment, error) {
	if r.db == nil {
		return nil, errors.New("health context database is unavailable")
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, NULL::uuid, context_node_id, scope_mode, is_default,
		       valid_from, valid_until, source, source_reference, created_by,
		       created_at, updated_at
		FROM user_health_context_assignments
		WHERE user_id = $1
		ORDER BY is_default DESC, created_at
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAssignments(rows)
}

func (r *postgresRepository) ReplaceUserAssignments(
	ctx context.Context,
	userID string,
	actorID string,
	inputs []AssignmentInput,
) error {
	if r.db == nil {
		return errors.New("health context database is unavailable")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx, `DELETE FROM user_health_context_assignments WHERE user_id = $1`, userID); err != nil {
		return err
	}

	for _, input := range inputs {
		source := defaultString(input.Source, "PORTAL")
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO user_health_context_assignments (
				user_id, context_node_id, scope_mode, is_default, valid_from,
				valid_until, source, source_reference, created_by
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		`, userID, input.ContextNodeID, input.ScopeMode, input.IsDefault, input.ValidFrom,
			input.ValidUntil, source, input.SourceReference, actorID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *postgresRepository) ListGroupAssignments(ctx context.Context, groupID uuid.UUID) ([]Assignment, error) {
	if r.db == nil {
		return nil, errors.New("health context database is unavailable")
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, ''::text, group_id, context_node_id, scope_mode, FALSE,
		       NULL::timestamptz, NULL::timestamptz, source, NULL::text,
		       created_by, created_at, updated_at
		FROM group_health_context_assignments
		WHERE group_id = $1
		ORDER BY created_at
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAssignments(rows)
}

func (r *postgresRepository) ReplaceGroupAssignments(
	ctx context.Context,
	groupID uuid.UUID,
	actorID string,
	inputs []AssignmentInput,
) error {
	if r.db == nil {
		return errors.New("health context database is unavailable")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx, `DELETE FROM group_health_context_assignments WHERE group_id = $1`, groupID); err != nil {
		return err
	}

	for _, input := range inputs {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO group_health_context_assignments (
				group_id, context_node_id, scope_mode, source, created_by
			) VALUES ($1,$2,$3,$4,$5)
		`, groupID, input.ContextNodeID, input.ScopeMode,
			defaultString(input.Source, "PORTAL"), actorID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *postgresRepository) ListGroupMappingDrift(
	ctx context.Context,
) ([]GroupMappingDrift, error) {
	if r.db == nil {
		return nil, errors.New("health context database is unavailable")
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT g.id,
		       g.path,
		       COALESCE(g.keycloak_group_id, ''),
		       g.enabled,
		       a.context_node_id,
		       COALESCE(n.code, ''),
		       COALESCE(n.name, ''),
		       COALESCE(n.enabled, FALSE),
		       COALESCE(a.scope_mode, '')
		FROM ihp_rbac_groups g
		LEFT JOIN group_health_context_assignments a ON a.group_id = g.id
		LEFT JOIN health_context_nodes n ON n.id = a.context_node_id
		ORDER BY g.path, n.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]GroupMappingDrift, 0)
	for rows.Next() {
		var item GroupMappingDrift
		var contextNodeID uuid.NullUUID
		var scopeMode string
		if err := rows.Scan(
			&item.GroupID,
			&item.GroupPath,
			&item.KeycloakGroupID,
			&item.GroupEnabled,
			&contextNodeID,
			&item.ContextCode,
			&item.ContextName,
			&item.ContextEnabled,
			&scopeMode,
		); err != nil {
			return nil, err
		}
		if contextNodeID.Valid {
			item.ContextNodeID = &contextNodeID.UUID
		}
		item.ScopeMode = ScopeMode(scopeMode)
		switch {
		case strings.TrimSpace(item.KeycloakGroupID) == "":
			item.Status = "MISSING_KEYCLOAK_LINK"
			item.RecommendedAction = "Run live Keycloak group sync before applying mappings."
		case !item.GroupEnabled:
			item.Status = "DISABLED_GROUP"
			item.RecommendedAction = "Review the disabled group and remove stale mappings if appropriate."
		case item.ContextNodeID == nil:
			item.Status = "UNMAPPED"
			item.RecommendedAction = "Map the group to an enabled health context."
		case !item.ContextEnabled:
			item.Status = "INACTIVE_CONTEXT"
			item.RecommendedAction = "Enable the context or move the group mapping."
		default:
			item.Status = "IN_SYNC"
			item.RecommendedAction = "No action required."
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *postgresRepository) ApplyGroupContextMappings(
	ctx context.Context,
	actorID string,
	mappings []GroupContextSyncMapping,
	replaceExisting bool,
) (GroupContextSyncResult, error) {
	if r.db == nil {
		return GroupContextSyncResult{}, errors.New("health context database is unavailable")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return GroupContextSyncResult{}, err
	}
	defer tx.Rollback()

	groups := make(map[uuid.UUID]struct{}, len(mappings))
	for _, mapping := range mappings {
		groups[mapping.GroupID] = struct{}{}
	}
	if replaceExisting {
		for groupID := range groups {
			if _, err := tx.ExecContext(
				ctx,
				`DELETE FROM group_health_context_assignments WHERE group_id = $1`,
				groupID,
			); err != nil {
				return GroupContextSyncResult{}, err
			}
		}
	}

	for _, mapping := range mappings {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO group_health_context_assignments (
				group_id, context_node_id, scope_mode, source, created_by
			) VALUES ($1, $2, $3, 'SYNC', $4)
			ON CONFLICT (group_id, context_node_id) DO UPDATE
			SET scope_mode = EXCLUDED.scope_mode,
			    source = EXCLUDED.source,
			    created_by = EXCLUDED.created_by,
			    updated_at = NOW()
		`, mapping.GroupID, mapping.ContextNodeID, mapping.ScopeMode, actorID); err != nil {
			return GroupContextSyncResult{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return GroupContextSyncResult{}, err
	}
	return GroupContextSyncResult{
		AppliedMappings: len(mappings),
		AffectedGroups:  len(groups),
	}, nil
}

func (r *postgresRepository) ListEffectiveContexts(ctx context.Context, userID string) ([]EffectiveContext, error) {
	if r.db == nil {
		return nil, errors.New("health context database is unavailable")
	}

	rows, err := r.db.QueryContext(ctx, `
		WITH assignments AS (
			SELECT a.id, a.context_node_id, a.scope_mode, 'DIRECT'::text assignment_type,
			       NULL::uuid group_id, NULL::text group_path, a.is_default
			FROM user_health_context_assignments a
			WHERE a.user_id = $1
			  AND (a.valid_from IS NULL OR a.valid_from <= NOW())
			  AND (a.valid_until IS NULL OR a.valid_until > NOW())
			UNION ALL
			SELECT a.id, a.context_node_id, a.scope_mode, 'GROUP'::text,
			       g.id, g.path, FALSE
			FROM ihp_rbac_group_members gm
			JOIN ihp_rbac_groups g ON g.id = gm.group_id AND g.enabled
			JOIN group_health_context_assignments a ON a.group_id = gm.group_id
			WHERE gm.user_id = $1
		)
		SELECT n.id, n.code, n.name, n.context_type, n.parent_id, n.source,
		       n.metadata, n.enabled, n.version, n.created_at, n.updated_at,
		       a.scope_mode, a.assignment_type, a.id, a.group_id, a.group_path,
		       a.is_default, COALESCE(active.context_node_id = n.id, FALSE)
		FROM assignments a
		JOIN health_context_nodes n ON n.id = a.context_node_id
		LEFT JOIN user_active_health_contexts active ON active.user_id = $1
		WHERE n.enabled
		ORDER BY a.is_default DESC, n.context_type, n.name
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]EffectiveContext, 0)
	for rows.Next() {
		var item EffectiveContext
		var parentID uuid.NullUUID
		var metadata []byte
		var contextType string
		var scopeMode string
		var groupID uuid.NullUUID
		var groupPath sql.NullString
		if err := rows.Scan(
			&item.ID, &item.Code, &item.Name, &contextType, &parentID,
			&item.Source, &metadata, &item.Enabled, &item.Version,
			&item.CreatedAt, &item.UpdatedAt, &scopeMode, &item.AssignmentType,
			&item.AssignmentID, &groupID, &groupPath, &item.IsDefault, &item.IsActive,
		); err != nil {
			return nil, err
		}
		item.ContextType = ContextType(contextType)
		item.ScopeMode = ScopeMode(scopeMode)
		item.Metadata = json.RawMessage(metadata)
		if parentID.Valid {
			item.ParentID = &parentID.UUID
		}
		if groupID.Valid {
			item.SourceGroupID = &groupID.UUID
		}
		if groupPath.Valid {
			item.SourceGroupPath = &groupPath.String
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *postgresRepository) ListAudienceUserIDs(
	ctx context.Context,
	contextIDs []uuid.UUID,
	includeDescendants bool,
) ([]string, error) {
	if r.db == nil {
		return nil, errors.New("health context database is unavailable")
	}
	if len(contextIDs) == 0 {
		return nil, nil
	}

	rows, err := r.db.QueryContext(ctx, `
		WITH selected_contexts AS (
			SELECT DISTINCT n.id
			FROM health_context_nodes n
			WHERE n.enabled
			  AND (
				n.id = ANY($1::uuid[])
				OR (
					$2
					AND EXISTS (
						SELECT 1
						FROM health_context_closure c
						WHERE c.ancestor_id = ANY($1::uuid[])
						  AND c.descendant_id = n.id
					)
				)
			  )
		),
		direct_users AS (
			SELECT DISTINCT a.user_id
			FROM user_health_context_assignments a
			JOIN health_context_nodes assigned ON assigned.id = a.context_node_id AND assigned.enabled
			WHERE (a.valid_from IS NULL OR a.valid_from <= NOW())
			  AND (a.valid_until IS NULL OR a.valid_until > NOW())
			  AND (
				a.context_node_id IN (SELECT id FROM selected_contexts)
				OR (
					a.scope_mode = 'NODE_AND_DESCENDANTS'
					AND EXISTS (
						SELECT 1
						FROM health_context_closure c
						JOIN selected_contexts selected ON selected.id = c.descendant_id
						WHERE c.ancestor_id = a.context_node_id
					)
				)
			  )
		),
		group_users AS (
			SELECT DISTINCT gm.user_id
			FROM group_health_context_assignments a
			JOIN ihp_rbac_groups g ON g.id = a.group_id AND g.enabled
			JOIN ihp_rbac_group_members gm ON gm.group_id = g.id
			JOIN health_context_nodes assigned ON assigned.id = a.context_node_id AND assigned.enabled
			WHERE (
				a.context_node_id IN (SELECT id FROM selected_contexts)
				OR (
					a.scope_mode = 'NODE_AND_DESCENDANTS'
					AND EXISTS (
						SELECT 1
						FROM health_context_closure c
						JOIN selected_contexts selected ON selected.id = c.descendant_id
						WHERE c.ancestor_id = a.context_node_id
					)
				)
			)
		)
		SELECT user_id FROM direct_users
		UNION
		SELECT user_id FROM group_users
		ORDER BY user_id
	`, pq.Array(contextIDs), includeDescendants)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]string, 0)
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		result = append(result, userID)
	}
	return result, rows.Err()
}

func (r *postgresRepository) GetActiveContext(ctx context.Context, userID string) (*uuid.UUID, error) {
	if r.db == nil {
		return nil, errors.New("health context database is unavailable")
	}

	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, `
		SELECT context_node_id FROM user_active_health_contexts WHERE user_id = $1
	`, userID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &id, err
}

func (r *postgresRepository) SetActiveContext(ctx context.Context, userID string, contextID uuid.UUID) error {
	if r.db == nil {
		return errors.New("health context database is unavailable")
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_active_health_contexts (user_id, context_node_id, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id) DO UPDATE
		SET context_node_id = EXCLUDED.context_node_id, updated_at = NOW()
	`, userID, contextID)
	return err
}

type rowScanner interface {
	Scan(...any) error
}

func scanNode(row rowScanner) (Node, error) {
	var node Node
	var parentID uuid.NullUUID
	var contextType string
	var metadata []byte
	err := row.Scan(
		&node.ID, &node.Code, &node.Name, &contextType, &parentID, &node.Source,
		&metadata, &node.Enabled, &node.Version, &node.CreatedAt, &node.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Node{}, ErrNotFound
	}
	if err != nil {
		return Node{}, err
	}
	node.ContextType = ContextType(contextType)
	node.Metadata = json.RawMessage(metadata)
	if parentID.Valid {
		node.ParentID = &parentID.UUID
	}
	return node, nil
}

func scanNodes(rows *sql.Rows) ([]Node, error) {
	result := make([]Node, 0)
	for rows.Next() {
		node, err := scanNode(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, node)
	}
	return result, rows.Err()
}

func scanAssignments(rows *sql.Rows) ([]Assignment, error) {
	result := make([]Assignment, 0)
	for rows.Next() {
		var item Assignment
		var groupID uuid.NullUUID
		var scopeMode string
		var validFrom, validUntil sql.NullTime
		var sourceReference, createdBy sql.NullString
		if err := rows.Scan(
			&item.ID, &item.UserID, &groupID, &item.ContextNodeID, &scopeMode,
			&item.IsDefault, &validFrom, &validUntil, &item.Source,
			&sourceReference, &createdBy, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		item.ScopeMode = ScopeMode(scopeMode)
		if groupID.Valid {
			item.GroupID = &groupID.UUID
		}
		if validFrom.Valid {
			item.ValidFrom = &validFrom.Time
		}
		if validUntil.Valid {
			item.ValidUntil = &validUntil.Time
		}
		if sourceReference.Valid {
			item.SourceReference = &sourceReference.String
		}
		if createdBy.Valid {
			item.CreatedBy = &createdBy.String
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}
