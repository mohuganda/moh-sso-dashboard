package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/moh-sso-dashboard/internal/http/response"
)

type DataQualityHandler struct {
	db *sql.DB
}

func NewDataQualityHandler(db *sql.DB) *DataQualityHandler {
	return &DataQualityHandler{
		db: db,
	}
}

type issueResponse struct {
	IssueID      int64   `json:"issue_id"`
	IssueCode    *string `json:"issue_code,omitempty"`
	Dataset      *string `json:"dataset,omitempty"`
	DataElement  *string `json:"data_element,omitempty"`
	OrgUnit      *string `json:"org_unit,omitempty"`
	Issue        *string `json:"issue,omitempty"`
	IssueType    *string `json:"issue_type,omitempty"`
	DateReported string  `json:"date_reported"`
	ReportedBy   *string `json:"reported_by,omitempty"`
	Status       *string `json:"status,omitempty"`
	Priority     *string `json:"priority,omitempty"`
	Severity     *string `json:"severity,omitempty"`
	UpdatedDate  *string `json:"updated_date,omitempty"`
	UpdatedBy    *string `json:"updated_by,omitempty"`
}

type createIssueRequest struct {
	Dataset     *string `json:"dataset"`
	DataElement *string `json:"data_element"`
	OrgUnit     *string `json:"org_unit"`
	Issue       string  `json:"issue" binding:"required"`
	IssueType   *string `json:"issue_type"`
	ReportedBy  *string `json:"reported_by"`
}

type updateIssueRequest struct {
	Dataset      *string `json:"dataset"`
	DataElement  *string `json:"data_element"`
	OrgUnit      *string `json:"org_unit"`
	Issue        *string `json:"issue"`
	IssueType    *string `json:"issue_type"`
	DateReported *string `json:"date_reported"`
	ReportedBy   *string `json:"reported_by"`
	Status       *string `json:"status"`
	Priority     *string `json:"priority"`
	Severity     *string `json:"severity"`
	UpdatedBy    *string `json:"updated_by"`
}

type createIssueResolutionRequest struct {
	Stage              *string `json:"stage"`
	Status             *string `json:"status"`
	ResolutionAction   *string `json:"resolution_action"`
	ResolvedBy         *string `json:"resolved_by"`
	ResolutionDate     *string `json:"resolution_date"`
	VerificationStatus *string `json:"verification_status"`
	VerifiedBy         *string `json:"verified_by"`
	VerificationDate   *string `json:"verification_date"`
	PreventiveAction   *string `json:"preventive_action"`
	ProcessChange      *string `json:"process_change"`
	PreventiveOwner    *string `json:"preventive_owner"`
	DueDate            *string `json:"due_date"`
}

type issueStageResponse struct {
	ID                 int64   `json:"id"`
	IssueCode          string  `json:"issue_code"`
	Status             *string `json:"status,omitempty"`
	Stage              *string `json:"stage,omitempty"`
	IsCurrent          bool    `json:"is_current"`
	ResolutionAction   *string `json:"resolution_action,omitempty"`
	ResolvedBy         *string `json:"resolved_by,omitempty"`
	ResolutionDate     *string `json:"resolution_date,omitempty"`
	VerificationStatus *string `json:"verification_status,omitempty"`
	VerifiedBy         *string `json:"verified_by,omitempty"`
	VerificationDate   *string `json:"verification_date,omitempty"`
	PreventiveAction   *string `json:"preventive_action,omitempty"`
	ProcessChange      *string `json:"process_change,omitempty"`
	PreventiveOwner    *string `json:"preventive_owner,omitempty"`
	DueDate            *string `json:"due_date,omitempty"`
}

func parseListLimit(c *gin.Context, defaultValue int, maxValue int) int {
	limit := defaultValue
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if limit > maxValue {
		return maxValue
	}

	return limit
}

func parseListOffset(c *gin.Context) int {
	offset := 0
	if raw := c.Query("offset"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return offset
}

func parseDateOnly(value string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(value))
}

func dqNullableString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return sql.NullString{}
	}

	return sql.NullString{
		String: trimmed,
		Valid:  true,
	}
}

func dqNullStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	v := value.String
	return &v
}

func dqNullInt64Ptr(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	v := value.Int64
	return &v
}

func dqNullDatePtr(value sql.NullTime) *string {
	if !value.Valid {
		return nil
	}
	v := value.Time.Format("2006-01-02")
	return &v
}

func optionalTrimmedParam(value *string, emptyAsNil bool) interface{} {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if emptyAsNil && trimmed == "" {
		return nil
	}

	return trimmed
}

func issueDBErrorMessage(err error, fallback string) string {
	if err == nil {
		return fallback
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		if pqErr.Detail != "" {
			return fmt.Sprintf("%s (sqlstate=%s, detail=%s)", pqErr.Message, string(pqErr.Code), pqErr.Detail)
		}
		return fmt.Sprintf("%s (sqlstate=%s)", pqErr.Message, string(pqErr.Code))
	}

	return err.Error()
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return false
	}

	return string(pqErr.Code) == "23505"
}

func scanIssue(scanner interface {
	Scan(dest ...interface{}) error
}) (issueResponse, error) {
	var (
		issueID      int64
		issueCode    sql.NullString
		dataset      sql.NullString
		dataElement  sql.NullString
		orgUnit      sql.NullString
		issueText    sql.NullString
		dateReported time.Time
		reportedBy   sql.NullString
		status       sql.NullString
		priority     sql.NullString
		severity     sql.NullString
		updatedDate  sql.NullTime
		updatedBy    sql.NullString
		issueType    sql.NullString
	)

	if err := scanner.Scan(
		&issueID,
		&issueCode,
		&dataset,
		&dataElement,
		&orgUnit,
		&issueText,
		&dateReported,
		&reportedBy,
		&status,
		&priority,
		&severity,
		&updatedDate,
		&updatedBy,
		&issueType,
	); err != nil {
		return issueResponse{}, err
	}

	return issueResponse{
		IssueID:      issueID,
		IssueCode:    dqNullStringPtr(issueCode),
		Dataset:      dqNullStringPtr(dataset),
		DataElement:  dqNullStringPtr(dataElement),
		OrgUnit:      dqNullStringPtr(orgUnit),
		Issue:        dqNullStringPtr(issueText),
		IssueType:    dqNullStringPtr(issueType),
		DateReported: dateReported.Format("2006-01-02"),
		ReportedBy:   dqNullStringPtr(reportedBy),
		Status:       dqNullStringPtr(status),
		Priority:     dqNullStringPtr(priority),
		Severity:     dqNullStringPtr(severity),
		UpdatedDate:  dqNullDatePtr(updatedDate),
		UpdatedBy:    dqNullStringPtr(updatedBy),
	}, nil
}

func (h *DataQualityHandler) CreateIssue(c *gin.Context) {
	var req createIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}

	issueText := strings.TrimSpace(req.Issue)
	if issueText == "" {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "issue is required")
		return
	}

	if req.IssueType != nil && strings.TrimSpace(*req.IssueType) == "" {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "issue_type cannot be empty")
		return
	}

	reportedBy := dqNullableString(req.ReportedBy)
	if !reportedBy.Valid {
		userID := strings.TrimSpace(c.GetString("user_id"))
		if userID != "" {
			reportedBy = sql.NullString{
				String: userID,
				Valid:  true,
			}
		}
	}

	row := h.db.QueryRowContext(
		c.Request.Context(),
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
			status
		)
		SELECT
			n.issue_id,
			'HMIS-' || LPAD(n.issue_id::text, 4, '0'),
			$1, $2, $3, $4, $5, CURRENT_DATE, $6, 'open'
		FROM next_issue n
		RETURNING issue_id, issue_code, dataset, data_element, org_unit, issue, date_reported, reported_by, status, priority, severity, updated_date, updated_by, issue_type`,
		dqNullableString(req.Dataset),
		dqNullableString(req.DataElement),
		dqNullableString(req.OrgUnit),
		issueText,
		dqNullableString(req.IssueType),
		reportedBy,
	)

	issue, err := scanIssue(row)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "CREATE_ISSUE_FAILED", issueDBErrorMessage(err, "failed to create issue"))
		return
	}

	response.OK(c, http.StatusCreated, issue)
}

func (h *DataQualityHandler) ListIssues(c *gin.Context) {
	limit := parseListLimit(c, 20, 100)
	offset := parseListOffset(c)

	rows, err := h.db.QueryContext(
		c.Request.Context(),
		`SELECT issue_id, issue_code, dataset, data_element, org_unit, issue, date_reported, reported_by, status, priority, severity, updated_date, updated_by, issue_type
		FROM hiv.issue
		ORDER BY date_reported DESC, issue_id DESC
		LIMIT $1 OFFSET $2`,
		limit,
		offset,
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_ISSUES_FAILED", issueDBErrorMessage(err, "failed to list issues"))
		return
	}
	defer rows.Close()

	issues := make([]issueResponse, 0)
	for rows.Next() {
		issue, scanErr := scanIssue(rows)
		if scanErr != nil {
			response.Fail(c, http.StatusInternalServerError, "LIST_ISSUES_FAILED", issueDBErrorMessage(scanErr, "failed to list issues"))
			return
		}
		issues = append(issues, issue)
	}

	if err = rows.Err(); err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_ISSUES_FAILED", issueDBErrorMessage(err, "failed to list issues"))
		return
	}

	response.OK(c, http.StatusOK, issues)
}

func (h *DataQualityHandler) UpdateIssue(c *gin.Context) {
	issueCode := strings.TrimSpace(c.Param("issueCode"))
	if issueCode == "" {
		response.Fail(c, http.StatusBadRequest, "INVALID_ISSUE_CODE", "issue code is required")
		return
	}

	var req updateIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}

	if req.Dataset == nil &&
		req.DataElement == nil &&
		req.OrgUnit == nil &&
		req.Issue == nil &&
		req.IssueType == nil &&
		req.DateReported == nil &&
		req.ReportedBy == nil &&
		req.Status == nil &&
		req.Priority == nil &&
		req.Severity == nil &&
		req.UpdatedBy == nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "at least one field is required")
		return
	}

	var issueParam interface{}
	if req.Issue != nil {
		trimmedIssue := strings.TrimSpace(*req.Issue)
		if trimmedIssue == "" {
			response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "issue cannot be empty")
			return
		}
		issueParam = trimmedIssue
	}

	var dateReportedParam interface{}
	if req.DateReported != nil {
		parsedDate, err := parseDateOnly(*req.DateReported)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "date_reported must be in YYYY-MM-DD format")
			return
		}
		dateReportedParam = parsedDate
	}

	var statusParam interface{}
	if req.Status != nil {
		trimmedStatus := strings.TrimSpace(*req.Status)
		if trimmedStatus == "" {
			response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "status cannot be empty")
			return
		}
		statusParam = trimmedStatus
	}

	var priorityParam interface{}
	if req.Priority != nil {
		trimmedPriority := strings.TrimSpace(*req.Priority)
		if trimmedPriority == "" {
			response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "priority cannot be empty")
			return
		}
		priorityParam = trimmedPriority
	}

	var severityParam interface{}
	if req.Severity != nil {
		trimmedSeverity := strings.TrimSpace(*req.Severity)
		if trimmedSeverity == "" {
			response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "severity cannot be empty")
			return
		}
		severityParam = trimmedSeverity
	}

	var issueTypeParam interface{}
	if req.IssueType != nil {
		trimmedIssueType := strings.TrimSpace(*req.IssueType)
		if trimmedIssueType == "" {
			response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "issue_type cannot be empty")
			return
		}
		issueTypeParam = trimmedIssueType
	}

	updatedByParam := optionalTrimmedParam(req.UpdatedBy, true)
	if updatedByParam == nil {
		userID := strings.TrimSpace(c.GetString("user_id"))
		if userID != "" {
			updatedByParam = userID
		}
	}

	row := h.db.QueryRowContext(
		c.Request.Context(),
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
			issue_type = COALESCE($12, issue_type)
		WHERE issue_code = $1
		RETURNING issue_id, issue_code, dataset, data_element, org_unit, issue, date_reported, reported_by, status, priority, severity, updated_date, updated_by, issue_type`,
		issueCode,
		optionalTrimmedParam(req.Dataset, false),
		optionalTrimmedParam(req.DataElement, false),
		optionalTrimmedParam(req.OrgUnit, false),
		issueParam,
		dateReportedParam,
		optionalTrimmedParam(req.ReportedBy, false),
		statusParam,
		priorityParam,
		severityParam,
		updatedByParam,
		issueTypeParam,
	)

	issue, err := scanIssue(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.Fail(c, http.StatusNotFound, "ISSUE_NOT_FOUND", "issue not found")
			return
		}
		response.Fail(c, http.StatusInternalServerError, "UPDATE_ISSUE_FAILED", issueDBErrorMessage(err, "failed to update issue"))
		return
	}

	response.OK(c, http.StatusOK, issue)
}

func (h *DataQualityHandler) ResolveIssue(c *gin.Context) {
	issueCode := strings.TrimSpace(c.Param("issueCode"))
	if issueCode == "" {
		response.Fail(c, http.StatusBadRequest, "INVALID_ISSUE_CODE", "issue code is required")
		return
	}

	var req createIssueResolutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}

	var status string
	switch {
	case req.Status != nil && strings.TrimSpace(*req.Status) != "":
		status = strings.TrimSpace(*req.Status)
	case req.Stage != nil && strings.TrimSpace(*req.Stage) != "":
		// Backward compatibility for existing clients using "stage".
		status = strings.TrimSpace(*req.Stage)
	default:
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "status is required")
		return
	}

	var resolutionDateParam interface{}
	if req.ResolutionDate != nil {
		parsedDate, err := parseDateOnly(*req.ResolutionDate)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "resolution_date must be in YYYY-MM-DD format")
			return
		}
		resolutionDateParam = parsedDate
	}

	var verificationDateParam interface{}
	if req.VerificationDate != nil {
		parsedDate, err := parseDateOnly(*req.VerificationDate)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "verification_date must be in YYYY-MM-DD format")
			return
		}
		verificationDateParam = parsedDate
	}

	var dueDateParam interface{}
	if req.DueDate != nil {
		parsedDate, err := parseDateOnly(*req.DueDate)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "due_date must be in YYYY-MM-DD format")
			return
		}
		dueDateParam = parsedDate
	}

	resolvedBy := optionalTrimmedParam(req.ResolvedBy, true)
	if resolvedBy == nil {
		userID := strings.TrimSpace(c.GetString("user_id"))
		if userID != "" {
			resolvedBy = userID
		}
	}

	ctx := c.Request.Context()
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "CREATE_ISSUE_STAGE_FAILED", issueDBErrorMessage(err, "failed to begin transaction"))
		return
	}
	defer tx.Rollback()

	var existingCode string
	err = tx.QueryRowContext(ctx, `SELECT issue_code FROM hiv.issue WHERE issue_code = $1`, issueCode).Scan(&existingCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.Fail(c, http.StatusNotFound, "ISSUE_NOT_FOUND", "issue not found")
			return
		}
		response.Fail(c, http.StatusInternalServerError, "CREATE_ISSUE_STAGE_FAILED", issueDBErrorMessage(err, "failed to validate issue"))
		return
	}

	if _, err = tx.ExecContext(
		ctx,
		`UPDATE hiv.issue_resolution
		SET is_current = FALSE
		WHERE issue_code = $1 AND is_current = TRUE`,
		issueCode,
	); err != nil {
		response.Fail(c, http.StatusInternalServerError, "CREATE_ISSUE_STAGE_FAILED", issueDBErrorMessage(err, "failed to update previous resolution status"))
		return
	}

	var stageRow issueStageResponse
	var statusOut sql.NullString
	var resolutionAction sql.NullString
	var resolvedByOut sql.NullString
	var resolutionDate sql.NullTime
	var verificationStatus sql.NullString
	var verifiedBy sql.NullString
	var verificationDate sql.NullTime
	var preventiveAction sql.NullString
	var processChange sql.NullString
	var preventiveOwner sql.NullString
	var dueDate sql.NullTime

	err = tx.QueryRowContext(
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
		issueCode,
		status,
		optionalTrimmedParam(req.ResolutionAction, true),
		resolvedBy,
		resolutionDateParam,
		optionalTrimmedParam(req.VerificationStatus, true),
		optionalTrimmedParam(req.VerifiedBy, true),
		verificationDateParam,
		optionalTrimmedParam(req.PreventiveAction, true),
		optionalTrimmedParam(req.ProcessChange, true),
		optionalTrimmedParam(req.PreventiveOwner, true),
		dueDateParam,
	).Scan(
		&stageRow.ID,
		&stageRow.IssueCode,
		&statusOut,
		&stageRow.IsCurrent,
		&resolutionAction,
		&resolvedByOut,
		&resolutionDate,
		&verificationStatus,
		&verifiedBy,
		&verificationDate,
		&preventiveAction,
		&processChange,
		&preventiveOwner,
		&dueDate,
	)
	if err != nil {
		if isUniqueViolation(err) {
			err = tx.QueryRowContext(
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
				issueCode,
				status,
				optionalTrimmedParam(req.ResolutionAction, true),
				resolvedBy,
				resolutionDateParam,
				optionalTrimmedParam(req.VerificationStatus, true),
				optionalTrimmedParam(req.VerifiedBy, true),
				verificationDateParam,
				optionalTrimmedParam(req.PreventiveAction, true),
				optionalTrimmedParam(req.ProcessChange, true),
				optionalTrimmedParam(req.PreventiveOwner, true),
				dueDateParam,
			).Scan(
				&stageRow.ID,
				&stageRow.IssueCode,
				&statusOut,
				&stageRow.IsCurrent,
				&resolutionAction,
				&resolvedByOut,
				&resolutionDate,
				&verificationStatus,
				&verifiedBy,
				&verificationDate,
				&preventiveAction,
				&processChange,
				&preventiveOwner,
				&dueDate,
			)
		}
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "CREATE_ISSUE_STAGE_FAILED", issueDBErrorMessage(err, "failed to create issue resolution"))
			return
		}
	}
	stageRow.Status = dqNullStringPtr(statusOut)
	stageRow.Stage = stageRow.Status
	stageRow.ResolutionAction = dqNullStringPtr(resolutionAction)
	stageRow.ResolvedBy = dqNullStringPtr(resolvedByOut)
	stageRow.ResolutionDate = dqNullDatePtr(resolutionDate)
	stageRow.VerificationStatus = dqNullStringPtr(verificationStatus)
	stageRow.VerifiedBy = dqNullStringPtr(verifiedBy)
	stageRow.VerificationDate = dqNullDatePtr(verificationDate)
	stageRow.PreventiveAction = dqNullStringPtr(preventiveAction)
	stageRow.ProcessChange = dqNullStringPtr(processChange)
	stageRow.PreventiveOwner = dqNullStringPtr(preventiveOwner)
	stageRow.DueDate = dqNullDatePtr(dueDate)

	if _, err = tx.ExecContext(
		ctx,
		`UPDATE hiv.issue
		SET
			status = $2,
			updated_date = CURRENT_DATE,
			updated_by = COALESCE($3, updated_by)
		WHERE issue_code = $1`,
		issueCode,
		status,
		resolvedBy,
	); err != nil {
		response.Fail(c, http.StatusInternalServerError, "CREATE_ISSUE_STAGE_FAILED", issueDBErrorMessage(err, "failed to sync issue status"))
		return
	}

	if err = tx.Commit(); err != nil {
		response.Fail(c, http.StatusInternalServerError, "CREATE_ISSUE_STAGE_FAILED", issueDBErrorMessage(err, "failed to commit transaction"))
		return
	}

	response.OK(c, http.StatusCreated, stageRow)
}

func (h *DataQualityHandler) ListIssueResolutionTransactions(c *gin.Context) {
	issueCode := strings.TrimSpace(c.Param("issueCode"))
	if issueCode == "" {
		response.Fail(c, http.StatusBadRequest, "INVALID_ISSUE_CODE", "issue code is required")
		return
	}

	limit := parseListLimit(c, 20, 100)
	offset := parseListOffset(c)

	var existingCode string
	err := h.db.QueryRowContext(
		c.Request.Context(),
		`SELECT issue_code FROM hiv.issue WHERE issue_code = $1`,
		issueCode,
	).Scan(&existingCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.Fail(c, http.StatusNotFound, "ISSUE_NOT_FOUND", "issue not found")
			return
		}
		response.Fail(c, http.StatusInternalServerError, "LIST_ISSUE_TRANSACTIONS_FAILED", issueDBErrorMessage(err, "failed to validate issue"))
		return
	}

	rows, err := h.db.QueryContext(
		c.Request.Context(),
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
		response.Fail(c, http.StatusInternalServerError, "LIST_ISSUE_TRANSACTIONS_FAILED", issueDBErrorMessage(err, "failed to list issue transactions"))
		return
	}
	defer rows.Close()

	transactions := make([]issueStageResponse, 0)
	for rows.Next() {
		var row issueStageResponse
		var statusOut sql.NullString
		var resolutionAction sql.NullString
		var resolvedBy sql.NullString
		var resolutionDate sql.NullTime
		var verificationStatus sql.NullString
		var verifiedBy sql.NullString
		var verificationDate sql.NullTime
		var preventiveAction sql.NullString
		var processChange sql.NullString
		var preventiveOwner sql.NullString
		var dueDate sql.NullTime

		scanErr := rows.Scan(
			&row.ID,
			&row.IssueCode,
			&statusOut,
			&row.IsCurrent,
			&resolutionAction,
			&resolvedBy,
			&resolutionDate,
			&verificationStatus,
			&verifiedBy,
			&verificationDate,
			&preventiveAction,
			&processChange,
			&preventiveOwner,
			&dueDate,
		)
		if scanErr != nil {
			response.Fail(c, http.StatusInternalServerError, "LIST_ISSUE_TRANSACTIONS_FAILED", issueDBErrorMessage(scanErr, "failed to list issue transactions"))
			return
		}

		row.Status = dqNullStringPtr(statusOut)
		row.Stage = row.Status
		row.ResolutionAction = dqNullStringPtr(resolutionAction)
		row.ResolvedBy = dqNullStringPtr(resolvedBy)
		row.ResolutionDate = dqNullDatePtr(resolutionDate)
		row.VerificationStatus = dqNullStringPtr(verificationStatus)
		row.VerifiedBy = dqNullStringPtr(verifiedBy)
		row.VerificationDate = dqNullDatePtr(verificationDate)
		row.PreventiveAction = dqNullStringPtr(preventiveAction)
		row.ProcessChange = dqNullStringPtr(processChange)
		row.PreventiveOwner = dqNullStringPtr(preventiveOwner)
		row.DueDate = dqNullDatePtr(dueDate)

		transactions = append(transactions, row)
	}

	if err = rows.Err(); err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_ISSUE_TRANSACTIONS_FAILED", issueDBErrorMessage(err, "failed to list issue transactions"))
		return
	}

	response.OK(c, http.StatusOK, transactions)
}
