package data_quality

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Repository interface {
	CreateIssue(ctx context.Context, input createIssueInput, scope healthContextScope) (issueResponse, error)
	ListIssues(ctx context.Context, limit int, offset int, program string, scope healthContextScope) ([]issueResponse, error)
	ListIssueSummaryByProgram(ctx context.Context, limit int, offset int, scope healthContextScope) ([]issueProgramSummaryResponse, error)
	UpdateIssue(ctx context.Context, input updateIssueInput, scope healthContextScope) (issueResponse, error)
	ResolveIssue(ctx context.Context, input resolveIssueInput, scope healthContextScope) (issueStageResponse, error)
	ListIssueResolutionTransactions(ctx context.Context, issueCode string, limit int, offset int, scope healthContextScope) ([]issueStageResponse, error)
	ImportValidationRules(ctx context.Context, inputs []validationRuleInput) (validationRuleImportResult, error)
	ListValidationRules(ctx context.Context, limit int, offset int) ([]validationRuleResponse, error)
	CountIssues(ctx context.Context, program string, scope healthContextScope) (int64, error)
}

type postgresRepository struct {
	dwhDB                 *sql.DB
	primaryDB             *sql.DB
	ownershipBackfillMu   sync.Mutex
	ownershipBackfilledAt time.Time
}

func NewRepository(dwhDB *sql.DB, primaryDB *sql.DB) Repository {
	return &postgresRepository{
		dwhDB:     dwhDB,
		primaryDB: primaryDB,
	}
}

func (r *postgresRepository) CreateIssue(ctx context.Context, input createIssueInput, scope healthContextScope) (issueResponse, error) {
	row := r.dwhDB.QueryRowContext(
		ctx,
		`WITH next_issue AS (
			SELECT nextval(pg_get_serial_sequence('hiv.issue', 'issue_id'))::bigint AS issue_id
		)
		INSERT INTO hiv.issue (
			issue_id,
			issue_code,
			dataset,
			data_element,
			org_unit,
			issue,
			issue_type,
			date_reported,
			reported_by,
			status,
			time_period
		)
		SELECT
			n.issue_id,
			'HMIS-' || LPAD(n.issue_id::text, 4, '0'),
			$1, $2, $3, $4, $5, CURRENT_DATE, $6, 'OPEN', $7
		FROM next_issue n
		RETURNING issue_id, issue_code, dataset, data_element, org_unit, issue, date_reported, reported_by, status, priority, severity, updated_date, updated_by, issue_type,time_period`,
		input.Dataset,
		input.DataElement,
		input.OrgUnit,
		input.Issue,
		input.IssueType,
		input.ReportedBy,
		input.TimePeriod,
	)
	issue, err := scanIssue(row)
	if err != nil {
		return issueResponse{}, err
	}
	if issue.IssueCode == nil {
		return issueResponse{}, errors.New("created issue has no issue code")
	}
	contextID, err := r.contextForNewIssue(ctx, scope)
	if err != nil {
		r.deleteIssueAfterOwnershipFailure(ctx, *issue.IssueCode)
		return issueResponse{}, err
	}
	if err := r.upsertIssueOwnership(ctx, *issue.IssueCode, contextID); err != nil {
		r.deleteIssueAfterOwnershipFailure(ctx, *issue.IssueCode)
		return issueResponse{}, err
	}
	issue.HealthContextID = &contextID
	return issue, nil
}

func (r *postgresRepository) ListIssues(ctx context.Context, limit int, offset int, program string, scope healthContextScope) ([]issueResponse, error) {
	ownerships, err := r.issueOwnershipsInScope(ctx, scope)
	if err != nil {
		return nil, err
	}
	query := `SELECT issue_id, issue_code, dataset, data_element, org_unit, issue, date_reported, reported_by, status, priority, severity, updated_date, updated_by, issue_type,time_period
		FROM hiv.issue
		WHERE ($3 = '' OR LOWER(BTRIM(program)) = LOWER(BTRIM($3)))`
	args := []any{limit, offset, program}
	if scope.ContextID != uuid.Nil {
		query += ` AND issue_code = ANY($4)`
		args = append(args, pq.Array(ownershipCodes(ownerships)))
	}
	query += ` ORDER BY date_reported DESC, issue_id DESC LIMIT $1 OFFSET $2`
	rows, err := r.dwhDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	issues := make([]issueResponse, 0)
	for rows.Next() {
		issue, err := scanIssue(rows)
		if err != nil {
			return nil, err
		}
		if issue.IssueCode != nil {
			if contextID, ok := ownerships[*issue.IssueCode]; ok {
				id := contextID
				issue.HealthContextID = &id
			}
		}
		issues = append(issues, issue)
	}
	return issues, rows.Err()
}

func (r *postgresRepository) ListIssueSummaryByProgram(ctx context.Context, limit int, offset int, scope healthContextScope) ([]issueProgramSummaryResponse, error) {
	ownerships, err := r.issueOwnershipsInScope(ctx, scope)
	if err != nil {
		return nil, err
	}
	query := `SELECT
			COALESCE(NULLIF(BTRIM(program), ''), 'Unspecified') AS program,
			COUNT(*)::bigint AS issue_count,
			COUNT(CASE WHEN UPPER(BTRIM(status)) NOT IN ('RESOLVED', 'CLOSED') THEN 1 END)::bigint AS open_count,
			COUNT(CASE WHEN UPPER(BTRIM(status)) IN ('RESOLVED', 'CLOSED') THEN 1 END)::bigint AS resolved_count
		FROM hiv.issue`
	args := []any{limit, offset}
	if scope.ContextID != uuid.Nil {
		query += ` WHERE issue_code = ANY($3)`
		args = append(args, pq.Array(ownershipCodes(ownerships)))
	}
	query += `
		GROUP BY 1
		ORDER BY issue_count DESC, program ASC
		LIMIT $1 OFFSET $2`
	rows, err := r.dwhDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summary := make([]issueProgramSummaryResponse, 0)
	for rows.Next() {
		var row issueProgramSummaryResponse
		if err := rows.Scan(&row.Program, &row.IssueCount, &row.OpenCount, &row.ResolvedCount); err != nil {
			return nil, err
		}
		summary = append(summary, row)
	}

	return summary, rows.Err()
}

func (r *postgresRepository) UpdateIssue(ctx context.Context, input updateIssueInput, scope healthContextScope) (issueResponse, error) {
	if err := r.requireIssueInScope(ctx, input.IssueCode, scope); err != nil {
		return issueResponse{}, err
	}
	row := r.dwhDB.QueryRowContext(
		ctx,
		`UPDATE hiv.issue
		SET
			dataset = COALESCE($2, dataset),
			data_element = COALESCE($3, data_element),
			org_unit = COALESCE($4, org_unit),
			issue = COALESCE($5, issue),
			date_reported = COALESCE($6, date_reported),
			reported_by = COALESCE($7, reported_by),
			status = COALESCE($8, status),
			priority = COALESCE($9, priority),
			severity = COALESCE($10, severity),
			updated_date = CURRENT_DATE,
			updated_by = COALESCE($11, updated_by),
			issue_type = COALESCE($12, issue_type),
			time_period = COALESCE($13, time_period)
		WHERE issue_code = $1
		RETURNING issue_id, issue_code, dataset, data_element, org_unit, issue, date_reported, reported_by, status, priority, severity, updated_date, updated_by, issue_type, time_period`,
		input.IssueCode,
		input.Dataset,
		input.DataElement,
		input.OrgUnit,
		input.Issue,
		input.DateReported,
		input.ReportedBy,
		input.Status,
		input.Priority,
		input.Severity,
		input.UpdatedBy,
		input.IssueType,
		input.TimePeriod,
	)
	return scanIssue(row)
}

func (r *postgresRepository) ResolveIssue(ctx context.Context, input resolveIssueInput, scope healthContextScope) (issueStageResponse, error) {
	if err := r.requireIssueInScope(ctx, input.IssueCode, scope); err != nil {
		return issueStageResponse{}, err
	}
	tx, err := r.dwhDB.BeginTx(ctx, nil)
	if err != nil {
		return issueStageResponse{}, err
	}
	defer tx.Rollback()

	var existingCode string
	err = tx.QueryRowContext(ctx, `SELECT issue_code FROM hiv.issue WHERE issue_code = $1`, input.IssueCode).Scan(&existingCode)
	if err != nil {
		return issueStageResponse{}, err
	}

	if _, err = tx.ExecContext(
		ctx,
		`UPDATE hiv.issue_resolution
		SET is_current = FALSE
		WHERE issue_code = $1 AND is_current = TRUE`,
		input.IssueCode,
	); err != nil {
		return issueStageResponse{}, err
	}

	rowScanner := tx.QueryRowContext(
		ctx,
		`INSERT INTO hiv.issue_resolution (
			issue_code,
			status,
			is_current,
			resolution_action,
			resolved_by,
			resolution_date,
			verification_status,
			verified_by,
			verification_date,
			preventive_action,
			process_change,
			preventive_owner,
			due_date
		) VALUES (
			$1, $2, TRUE, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
		RETURNING id, issue_code, status, is_current, resolution_action, resolved_by, resolution_date,
			verification_status, verified_by, verification_date, preventive_action, process_change, preventive_owner, due_date`,
		input.IssueCode,
		input.Status,
		input.ResolutionAction,
		input.ResolvedBy,
		input.ResolutionDate,
		input.VerificationStatus,
		input.VerifiedBy,
		input.VerificationDate,
		input.PreventiveAction,
		input.ProcessChange,
		input.PreventiveOwner,
		input.DueDate,
	)
	stageRow, err := scanIssueStage(rowScanner)
	if err != nil && isUniqueViolation(err) {
		stageRow, err = scanIssueStage(tx.QueryRowContext(
			ctx,
			`UPDATE hiv.issue_resolution
				SET
					status = $2,
					is_current = TRUE,
					resolution_action = $3,
					resolved_by = $4,
					resolution_date = $5,
					verification_status = $6,
					verified_by = $7,
					verification_date = $8,
					preventive_action = $9,
					process_change = $10,
					preventive_owner = $11,
					due_date = $12
				WHERE issue_code = $1
				RETURNING id, issue_code, status, is_current, resolution_action, resolved_by, resolution_date,
					verification_status, verified_by, verification_date, preventive_action, process_change, preventive_owner, due_date`,
			input.IssueCode,
			input.Status,
			input.ResolutionAction,
			input.ResolvedBy,
			input.ResolutionDate,
			input.VerificationStatus,
			input.VerifiedBy,
			input.VerificationDate,
			input.PreventiveAction,
			input.ProcessChange,
			input.PreventiveOwner,
			input.DueDate,
		))
	}
	if err != nil {
		return issueStageResponse{}, err
	}

	if _, err = tx.ExecContext(
		ctx,
		`UPDATE hiv.issue
		SET
			status = $2,
			updated_date = CURRENT_DATE,
			updated_by = COALESCE($3, updated_by)
		WHERE issue_code = $1`,
		input.IssueCode,
		input.Status,
		input.ResolvedBy,
	); err != nil {
		return issueStageResponse{}, err
	}

	if err = tx.Commit(); err != nil {
		return issueStageResponse{}, err
	}

	return stageRow, nil
}

func (r *postgresRepository) ListIssueResolutionTransactions(ctx context.Context, issueCode string, limit int, offset int, scope healthContextScope) ([]issueStageResponse, error) {
	if err := r.requireIssueInScope(ctx, issueCode, scope); err != nil {
		return nil, err
	}
	var existingCode string
	err := r.dwhDB.QueryRowContext(
		ctx,
		`SELECT issue_code FROM hiv.issue WHERE issue_code = $1`,
		issueCode,
	).Scan(&existingCode)
	if err != nil {
		return nil, err
	}

	rows, err := r.dwhDB.QueryContext(
		ctx,
		`SELECT
			id,
			issue_code,
			status,
			is_current,
			resolution_action,
			resolved_by,
			resolution_date,
			verification_status,
			verified_by,
			verification_date,
			preventive_action,
			process_change,
			preventive_owner,
			due_date
		FROM hiv.issue_resolution
		WHERE issue_code = $1
		ORDER BY id DESC
		LIMIT $2 OFFSET $3`,
		issueCode,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := make([]issueStageResponse, 0)
	for rows.Next() {
		transaction, err := scanIssueStage(rows)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(transactions) == 0 && existingCode == "" {
		return nil, sql.ErrNoRows
	}

	return transactions, nil
}

func (r *postgresRepository) ImportValidationRules(ctx context.Context, inputs []validationRuleInput) (validationRuleImportResult, error) {
	result := validationRuleImportResult{
		Errors: []validationRuleRowError{},
		Rules:  []validationRuleResponse{},
	}

	if len(inputs) == 0 {
		return result, nil
	}

	tx, err := r.primaryDB.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer tx.Rollback()

	for _, input := range inputs {
		rule, created, err := upsertValidationRule(ctx, tx, input)
		if err != nil {
			return result, err
		}

		result.Imported++
		if created {
			result.Created++
		} else {
			result.Updated++
		}
		result.Rules = append(result.Rules, rule)
	}

	if err = tx.Commit(); err != nil {
		return result, err
	}

	return result, nil
}

func (r *postgresRepository) ListValidationRules(ctx context.Context, limit int, offset int) ([]validationRuleResponse, error) {
	rows, err := r.primaryDB.QueryContext(
		ctx,
		`SELECT
			id,
			table_id,
			program,
			category,
			code,
			severity,
			description,
			column_name,
			operator,
			value,
			value_column,
			is_active,
			created_by,
			updated_by,
			created_at,
			updated_at,
			deleted_at
		FROM data_quality_validation_rules
		WHERE deleted_at IS NULL
		ORDER BY updated_at DESC, id DESC
		LIMIT $1 OFFSET $2`,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := make([]validationRuleResponse, 0)
	for rows.Next() {
		rule, err := scanValidationRule(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	return rules, rows.Err()
}

func upsertValidationRule(ctx context.Context, tx *sql.Tx, input validationRuleInput) (validationRuleResponse, bool, error) {
	row := tx.QueryRowContext(
		ctx,
		`INSERT INTO data_quality_validation_rules (
			table_id,
			program,
			category,
			code,
			severity,
			description,
			column_name,
			operator,
			value,
			value_column,
			is_active,
			created_by,
			updated_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, TRUE, $11, $11
		)
		ON CONFLICT (
			COALESCE(table_id, ''),
			COALESCE(program, ''),
			COALESCE(category, ''),
			code
		)
		WHERE deleted_at IS NULL
		DO UPDATE SET
			severity = EXCLUDED.severity,
			description = EXCLUDED.description,
			column_name = EXCLUDED.column_name,
			operator = EXCLUDED.operator,
			value = EXCLUDED.value,
			value_column = EXCLUDED.value_column,
			is_active = TRUE,
			updated_by = EXCLUDED.updated_by,
			updated_at = now()
		RETURNING
			id,
			table_id,
			program,
			category,
			code,
			severity,
			description,
			column_name,
			operator,
			value,
			value_column,
			is_active,
			created_by,
			updated_by,
			created_at,
			updated_at,
			deleted_at,
			(xmax = 0) AS created`,
		input.TableID,
		input.Program,
		input.Category,
		input.Code,
		input.Severity,
		input.Description,
		input.Column,
		input.Op,
		input.Value,
		input.ValueColumn,
		input.CreatedBy,
	)

	return scanValidationRuleWithCreated(row)
}

func isNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return false
	}

	return string(pqErr.Code) == "23505"
}

func (r *postgresRepository) CountIssues(ctx context.Context, program string, scope healthContextScope) (int64, error) {
	ownerships, err := r.issueOwnershipsInScope(ctx, scope)
	if err != nil {
		return 0, err
	}
	var count int64
	query := `SELECT COUNT(*)::bigint FROM hiv.issue WHERE ($1 = '' OR LOWER(BTRIM(program)) = LOWER(BTRIM($1)))`
	args := []any{program}
	if scope.ContextID != uuid.Nil {
		query += ` AND issue_code = ANY($2)`
		args = append(args, pq.Array(ownershipCodes(ownerships)))
	}
	err = r.dwhDB.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}
