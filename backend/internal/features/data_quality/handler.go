package data_quality

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/keycloak"
	sharedservice "github.com/moh-sso-dashboard/internal/service"
)

type Handler struct {
	service        Service
	keyAdminClient *keycloak.KeyAdminClient
}

func NewHandler(dwhDB *sql.DB, primaryDB *sql.DB, keyAdminClient *keycloak.KeyAdminClient, emailService ...sharedservice.EmailService) *Handler {
	var emailSvc sharedservice.EmailService
	if len(emailService) > 0 {
		emailSvc = emailService[0]
	}
	return &Handler{
		service:        NewService(NewRepository(dwhDB, primaryDB), emailSvc),
		keyAdminClient: keyAdminClient,
	}
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

func issueDBErrorMessage(err error, fallback string) string {
	if err == nil {
		return fallback
	}
	return fallback
}

func (h *Handler) CreateIssue(c *gin.Context) {
	var req createIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "invalid request payload")
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

	issue, err := h.service.CreateIssue(c.Request.Context(), createIssueInput{
		Dataset:     dqNullableString(req.Dataset),
		DataElement: dqNullableString(req.DataElement),
		OrgUnit:     dqNullableString(req.OrgUnit),
		Issue:       issueText,
		IssueType:   dqNullableString(req.IssueType),
		ReportedBy:  reportedBy,
		TimePeriod:  dqNullableString(req.TimePeriod),
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "CREATE_ISSUE_FAILED", issueDBErrorMessage(err, "failed to create issue"))
		return
	}

	response.OK(c, http.StatusCreated, toIssueResponse(issue))
}

func (h *Handler) ListIssues(c *gin.Context) {
	limit := parseListLimit(c, 20, 10000)
	offset := parseListOffset(c)
	program := strings.TrimSpace(c.Query("program"))

	issues, err := h.service.ListIssues(c.Request.Context(), limit, offset, program)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_ISSUES_FAILED", issueDBErrorMessage(err, "failed to list issues"))
		return
	}

	total, err := h.service.CountIssues(c.Request.Context(), program)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_ISSUES_FAILED", issueDBErrorMessage(err, "failed to count issues"))
		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.Header("Access-Control-Expose-Headers", "X-Total-Count")

	response.OK(c, http.StatusOK, toIssueResponses(issues))
}

func (h *Handler) ListIssueSummaryByProgram(c *gin.Context) {
	limit := parseListLimit(c, 50, 500)
	offset := parseListOffset(c)

	summary, err := h.service.ListIssueSummaryByProgram(c.Request.Context(), limit, offset)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_ISSUE_SUMMARY_FAILED", issueDBErrorMessage(err, "failed to list issue summary"))
		return
	}

	response.OK(c, http.StatusOK, toIssueProgramSummaryResponses(summary))
}

func (h *Handler) UpdateIssue(c *gin.Context) {
	issueCode := strings.TrimSpace(c.Param("issueCode"))
	if issueCode == "" {
		response.Fail(c, http.StatusBadRequest, "INVALID_ISSUE_CODE", "issue code is required")
		return
	}

	var req updateIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "invalid request payload")
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
		req.UpdatedBy == nil &&
		req.TimePeriod == nil {
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

	issue, err := h.service.UpdateIssue(c.Request.Context(), updateIssueInput{
		IssueCode:    issueCode,
		Dataset:      optionalTrimmedParam(req.Dataset, false),
		DataElement:  optionalTrimmedParam(req.DataElement, false),
		OrgUnit:      optionalTrimmedParam(req.OrgUnit, false),
		Issue:        issueParam,
		DateReported: dateReportedParam,
		ReportedBy:   optionalTrimmedParam(req.ReportedBy, false),
		Status:       statusParam,
		Priority:     priorityParam,
		Severity:     severityParam,
		UpdatedBy:    updatedByParam,
		IssueType:    issueTypeParam,
		TimePeriod:   optionalTrimmedParam(req.TimePeriod, false),
	})
	if err != nil {
		if isNotFound(err) {
			response.Fail(c, http.StatusNotFound, "ISSUE_NOT_FOUND", "issue not found")
			return
		}
		response.Fail(c, http.StatusInternalServerError, "UPDATE_ISSUE_FAILED", issueDBErrorMessage(err, "failed to update issue"))
		return
	}

	response.OK(c, http.StatusOK, toIssueResponse(issue))
}

func (h *Handler) ResolveIssue(c *gin.Context) {
	issueCode := strings.TrimSpace(c.Param("issueCode"))
	if issueCode == "" {
		response.Fail(c, http.StatusBadRequest, "INVALID_ISSUE_CODE", "issue code is required")
		return
	}

	var req createIssueResolutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "invalid request payload")
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

	stageRow, err := h.service.ResolveIssue(c.Request.Context(), resolveIssueInput{
		IssueCode:          issueCode,
		Status:             status,
		ResolutionAction:   optionalTrimmedParam(req.ResolutionAction, true),
		ResolvedBy:         resolvedBy,
		ResolutionDate:     resolutionDateParam,
		VerificationStatus: optionalTrimmedParam(req.VerificationStatus, true),
		VerifiedBy:         optionalTrimmedParam(req.VerifiedBy, true),
		VerificationDate:   verificationDateParam,
		PreventiveAction:   optionalTrimmedParam(req.PreventiveAction, true),
		ProcessChange:      optionalTrimmedParam(req.ProcessChange, true),
		PreventiveOwner:    optionalTrimmedParam(req.PreventiveOwner, true),
		DueDate:            dueDateParam,
	})
	if err != nil {
		if isNotFound(err) {
			response.Fail(c, http.StatusNotFound, "ISSUE_NOT_FOUND", "issue not found")
			return
		}
		response.Fail(c, http.StatusInternalServerError, "CREATE_ISSUE_STAGE_FAILED", issueDBErrorMessage(err, "failed to create issue resolution"))
		return
	}

	response.OK(c, http.StatusCreated, toIssueStageResponse(stageRow))
}

func (h *Handler) ListIssueResolutionTransactions(c *gin.Context) {
	issueCode := strings.TrimSpace(c.Param("issueCode"))
	if issueCode == "" {
		response.Fail(c, http.StatusBadRequest, "INVALID_ISSUE_CODE", "issue code is required")
		return
	}

	limit := parseListLimit(c, 20, 100)
	offset := parseListOffset(c)

	transactions, err := h.service.ListIssueResolutionTransactions(c.Request.Context(), issueCode, limit, offset)
	if err != nil {
		if isNotFound(err) {
			response.Fail(c, http.StatusNotFound, "ISSUE_NOT_FOUND", "issue not found")
			return
		}
		response.Fail(c, http.StatusInternalServerError, "LIST_ISSUE_TRANSACTIONS_FAILED", issueDBErrorMessage(err, "failed to list issue transactions"))
		return
	}

	response.OK(c, http.StatusOK, toIssueStageResponses(transactions))
}

func (h *Handler) ImportValidationRules(c *gin.Context) {
	requests, err := readValidationRuleImportPayload(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}

	createdBy := optionalRequestUserID(c)
	inputs := make([]validationRuleInput, 0, len(requests))
	result := validationRuleImportResult{
		Errors: []validationRuleRowError{},
		Rules:  []validationRuleResponse{},
	}

	for index, req := range requests {
		input, validationMessage := normalizeValidationRuleRequest(req, createdBy)
		if validationMessage != "" {
			result.Skipped++
			result.Errors = append(result.Errors, validationRuleRowError{
				Index:   index,
				Code:    strings.TrimSpace(req.Code),
				Message: validationMessage,
			})
			continue
		}

		inputs = append(inputs, input)
	}

	imported, err := h.service.ImportValidationRules(c.Request.Context(), inputs)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "IMPORT_VALIDATION_RULES_FAILED", "failed to import validation rules")
		return
	}

	result.Imported = imported.Imported
	result.Created = imported.Created
	result.Updated = imported.Updated
	result.Rules = imported.Rules

	response.OK(c, http.StatusCreated, result)
}

func (h *Handler) ListValidationRules(c *gin.Context) {
	limit := parseListLimit(c, 50, 500)
	offset := parseListOffset(c)

	rules, err := h.service.ListValidationRules(c.Request.Context(), limit, offset)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_VALIDATION_RULES_FAILED", "failed to list validation rules")
		return
	}

	response.OK(c, http.StatusOK, toValidationRuleResponses(rules))
}

func readValidationRuleImportPayload(c *gin.Context) ([]validationRuleImportRequest, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, errors.New("invalid request payload")
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return nil, errors.New("request body is required")
	}

	var direct []validationRuleImportRequest
	if err := json.Unmarshal(body, &direct); err == nil {
		if len(direct) == 0 {
			return nil, errors.New("at least one rule is required")
		}
		return direct, nil
	}

	var envelope validationRuleImportEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, errors.New("payload must be an array of rules or an object with a rules array")
	}
	if len(envelope.Rules) == 0 {
		return nil, errors.New("at least one rule is required")
	}

	return envelope.Rules, nil
}

func optionalRequestUserID(c *gin.Context) interface{} {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		return nil
	}
	return userID
}

func (h *Handler) ListKeycloakGroups(c *gin.Context) {
	if h.keyAdminClient == nil {
		response.OK(c, http.StatusOK, []keycloakGroupResponse{})
		return
	}

	groups, err := h.keyAdminClient.ListGroups(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "KEYCLOAK_GROUPS_FAILED", "failed to fetch Keycloak groups")
		return
	}

	out := make([]keycloakGroupResponse, 0, len(groups))
	for _, g := range groups {
		out = append(out, keycloakGroupResponse{
			ID:   g.ID,
			Name: g.Name,
			Path: g.Path,
		})
	}

	response.OK(c, http.StatusOK, out)
}

func (h *Handler) ListKeycloakGroupMembers(c *gin.Context) {
	groupID := strings.TrimSpace(c.Param("groupId"))
	if groupID == "" {
		response.Fail(c, http.StatusBadRequest, "INVALID_GROUP_ID", "group ID is required")
		return
	}

	if h.keyAdminClient == nil {
		response.OK(c, http.StatusOK, []keycloakGroupMemberResponse{})
		return
	}

	members, err := h.keyAdminClient.ListGroupMembers(c.Request.Context(), groupID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "KEYCLOAK_MEMBERS_FAILED", "failed to fetch Keycloak group members")
		return
	}

	out := make([]keycloakGroupMemberResponse, 0, len(members))
	for _, m := range members {
		out = append(out, keycloakGroupMemberResponse{
			ID:        m.ID,
			Username:  m.Username,
			Email:     m.Email,
			FirstName: m.FirstName,
			LastName:  m.LastName,
		})
	}

	response.OK(c, http.StatusOK, out)
}

func (h *Handler) AssignIssues(c *gin.Context) {
	var req assignIssuesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "invalid request payload")
		return
	}

	assignedTo := strings.TrimSpace(req.AssignedTo)
	if assignedTo == "" {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "assigned_to (user email) is required")
		return
	}

	issueCodes := make([]string, 0)
	for _, code := range req.IssueCodes {
		trimmed := strings.TrimSpace(code)
		if trimmed != "" {
			issueCodes = append(issueCodes, trimmed)
		}
	}
	if len(issueCodes) == 0 {
		paramCode := strings.TrimSpace(c.Param("issueCode"))
		if paramCode != "" {
			issueCodes = append(issueCodes, paramCode)
		} else if req.IssueCode != nil && strings.TrimSpace(*req.IssueCode) != "" {
			issueCodes = append(issueCodes, strings.TrimSpace(*req.IssueCode))
		}
	}

	if len(issueCodes) == 0 {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "at least one issue code is required")
		return
	}

	assignedBy := strings.TrimSpace(c.GetString("user_id"))
	if assignedBy == "" {
		assignedBy = "System"
	}

	comment := ""
	if req.Comment != nil {
		comment = strings.TrimSpace(*req.Comment)
	}

	results, err := h.service.AssignIssues(c.Request.Context(), assignIssuesInput{
		IssueCodes: issueCodes,
		AssignedTo: assignedTo,
		AssignedBy: assignedBy,
		Comment:    comment,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "ASSIGN_ISSUES_FAILED", issueDBErrorMessage(err, "failed to assign issue(s)"))
		return
	}

	response.OK(c, http.StatusOK, toIssueStageResponses(results))
}
