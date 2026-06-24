package data_quality

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Repository interface {
	CreateIssue(ctx context.Context, input createIssueInput) (issueResponse, error)
	ListIssues(ctx context.Context, limit int, offset int) ([]issueResponse, error)
	UpdateIssue(ctx context.Context, input updateIssueInput) (issueResponse, error)
	ResolveIssue(ctx context.Context, input resolveIssueInput) (issueStageResponse, error)
	ListIssueResolutionTransactions(ctx context.Context, issueCode string, limit int, offset int) ([]issueStageResponse, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) CreateIssue(ctx context.Context, input createIssueInput) (issueResponse, error) {
	row := r.db.QueryRowContext(
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
	return scanIssue(row)
}

func (r *postgresRepository) ListIssues(ctx context.Context, limit int, offset int) ([]issueResponse, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT issue_id, issue_code, dataset, data_element, org_unit, issue, date_reported, reported_by, status, priority, severity, updated_date, updated_by, issue_type,time_period
		FROM hiv.issue
		ORDER BY date_reported DESC, issue_id DESC
		LIMIT $1 OFFSET $2`,
		limit,
		offset,
	)
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
		issues = append(issues, issue)
	}
	return issues, rows.Err()
}

func (r *postgresRepository) UpdateIssue(ctx context.Context, input updateIssueInput) (issueResponse, error) {
	row := r.db.QueryRowContext(
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

func (r *postgresRepository) ResolveIssue(ctx context.Context, input resolveIssueInput) (issueStageResponse, error) {
	tx, err := r.db.BeginTx(ctx, nil)
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

func (r *postgresRepository) ListIssueResolutionTransactions(ctx context.Context, issueCode string, limit int, offset int) ([]issueStageResponse, error) {
	var existingCode string
	err := r.db.QueryRowContext(
		ctx,
		`SELECT issue_code FROM hiv.issue WHERE issue_code = $1`,
		issueCode,
	).Scan(&existingCode)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(
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
