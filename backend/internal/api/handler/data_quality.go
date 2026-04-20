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
	DateReported string  `json:"date_reported"`
	ReportedBy   *string `json:"reported_by,omitempty"`
}

type createIssueRequest struct {
	Dataset     *string `json:"dataset"`
	DataElement *string `json:"data_element"`
	OrgUnit     *string `json:"org_unit"`
	Issue       string  `json:"issue" binding:"required"`
	ReportedBy  *string `json:"reported_by"`
}

type updateIssueRequest struct {
	Dataset      *string `json:"dataset"`
	DataElement  *string `json:"data_element"`
	OrgUnit      *string `json:"org_unit"`
	Issue        *string `json:"issue"`
	DateReported *string `json:"date_reported"`
	ReportedBy   *string `json:"reported_by"`
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
		DateReported: dateReported.Format("2006-01-02"),
		ReportedBy:   dqNullStringPtr(reportedBy),
	}, nil
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
			date_reported,
			reported_by
		)
		SELECT
			n.issue_id,
			'HMIS-' || LPAD(n.issue_id::text, 4, '0'),
			$1, $2, $3, $4, CURRENT_DATE, $5
		FROM next_issue n
		RETURNING issue_id, issue_code, dataset, data_element, org_unit, issue, date_reported, reported_by`,
		dqNullableString(req.Dataset),
		dqNullableString(req.DataElement),
		dqNullableString(req.OrgUnit),
		issueText,
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
		`SELECT issue_id, issue_code, dataset, data_element, org_unit, issue, date_reported, reported_by
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
	rawIssueID := strings.TrimSpace(c.Param("id"))
	if rawIssueID == "" {
		response.Fail(c, http.StatusBadRequest, "INVALID_ID", "issue id is required")
		return
	}
	issueID, err := strconv.ParseInt(rawIssueID, 10, 64)
	if err != nil || issueID <= 0 {
		response.Fail(c, http.StatusBadRequest, "INVALID_ID", "issue id must be a positive integer")
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
		req.DateReported == nil &&
		req.ReportedBy == nil {
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

	row := h.db.QueryRowContext(
		c.Request.Context(),
		`UPDATE hiv.issue
		SET
			dataset = COALESCE($2, dataset),
			data_element = COALESCE($3, data_element),
			org_unit = COALESCE($4, org_unit),
			issue = COALESCE($5, issue),
			date_reported = COALESCE($6, date_reported),
			reported_by = COALESCE($7, reported_by)
		WHERE issue_id = $1
		RETURNING issue_id, issue_code, dataset, data_element, org_unit, issue, date_reported, reported_by`,
		issueID,
		optionalTrimmedParam(req.Dataset, false),
		optionalTrimmedParam(req.DataElement, false),
		optionalTrimmedParam(req.OrgUnit, false),
		issueParam,
		dateReportedParam,
		optionalTrimmedParam(req.ReportedBy, false),
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
