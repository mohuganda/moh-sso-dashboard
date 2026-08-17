package data_quality

import (
	"database/sql"
	"strings"
	"time"
)

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

func toIssueResponse(issue issueResponse) issueResponse {
	return issue
}

func toIssueResponses(issues []issueResponse) []issueResponse {
	if issues == nil {
		return []issueResponse{}
	}
	out := make([]issueResponse, 0, len(issues))
	for _, issue := range issues {
		out = append(out, toIssueResponse(issue))
	}
	return out
}

func toIssueProgramSummaryResponses(rows []issueProgramSummaryResponse) []issueProgramSummaryResponse {
	if rows == nil {
		return []issueProgramSummaryResponse{}
	}
	return rows
}

func toIssueStageResponse(stage issueStageResponse) issueStageResponse {
	return stage
}

func toIssueStageResponses(stages []issueStageResponse) []issueStageResponse {
	if stages == nil {
		return []issueStageResponse{}
	}
	out := make([]issueStageResponse, 0, len(stages))
	for _, stage := range stages {
		out = append(out, toIssueStageResponse(stage))
	}
	return out
}

func toValidationRuleResponses(rules []validationRuleResponse) []validationRuleResponse {
	if rules == nil {
		return []validationRuleResponse{}
	}
	return rules
}

func trimStringPtr(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func normalizeValidationRuleRequest(req validationRuleImportRequest, createdBy interface{}) (validationRuleInput, string) {
	code := strings.TrimSpace(req.Code)
	severity := strings.ToLower(strings.TrimSpace(req.Severity))
	column := strings.TrimSpace(req.Column)
	op := strings.ToLower(strings.TrimSpace(req.Op))

	if code == "" {
		return validationRuleInput{}, "code is required"
	}
	if !isValidValidationSeverity(severity) {
		return validationRuleInput{}, "severity must be one of error, warning, info"
	}
	if column == "" {
		return validationRuleInput{}, "column is required"
	}
	if !isValidValidationOperator(op) {
		return validationRuleInput{}, "op must be one of contains, eq, gt, gte, isnull, lt, lte, ne, notnull"
	}

	value := trimStringPtr(req.Value)
	valueColumn := trimStringPtr(req.ValueColumn)
	if op != "isnull" && op != "notnull" {
		if value == nil && valueColumn == nil {
			return validationRuleInput{}, "value or value_column is required for this operator"
		}
		if value != nil && valueColumn != nil {
			return validationRuleInput{}, "only one of value or value_column can be provided"
		}
	}

	description := ""
	if req.Description != nil {
		description = strings.TrimSpace(*req.Description)
	}

	return validationRuleInput{
		TableID:     dqNullableString(trimStringPtr(req.TableID)),
		Program:     dqNullableString(trimStringPtr(req.Program)),
		Category:    dqNullableString(trimStringPtr(req.Category)),
		Code:        code,
		Severity:    severity,
		Description: description,
		Column:      column,
		Op:          op,
		Value:       dqNullableString(value),
		ValueColumn: dqNullableString(valueColumn),
		CreatedBy:   createdBy,
	}, ""
}

func isValidValidationSeverity(value string) bool {
	switch value {
	case "error", "warning", "info":
		return true
	default:
		return false
	}
}

func isValidValidationOperator(value string) bool {
	switch value {
	case "contains", "eq", "gt", "gte", "isnull", "lt", "lte", "ne", "notnull":
		return true
	default:
		return false
	}
}

func scanIssue(scanner interface {
	Scan(dest ...interface{}) error
}) (issueResponse, error) {
	var (
		issueID      int64
		issueCode    sql.NullString
		dataset      sql.NullString
		dataElement  sql.NullString
		region       sql.NullString
		district     sql.NullString
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
		timePeriod   sql.NullString
		assignedTo   sql.NullString
	)

	if err := scanner.Scan(
		&issueID,
		&issueCode,
		&dataset,
		&dataElement,
		&region,
		&district,
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
		&timePeriod,
		&assignedTo,
	); err != nil {
		return issueResponse{}, err
	}

	return issueResponse{
		IssueID:      issueID,
		IssueCode:    dqNullStringPtr(issueCode),
		Dataset:      dqNullStringPtr(dataset),
		DataElement:  dqNullStringPtr(dataElement),
		Region:       dqNullStringPtr(region),
		District:     dqNullStringPtr(district),
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
		TimePeriod:   dqNullStringPtr(timePeriod),
		AssignedTo:   dqNullStringPtr(assignedTo),
	}, nil
}

func scanIssueStage(scanner interface {
	Scan(dest ...interface{}) error
}) (issueStageResponse, error) {
	var row issueStageResponse
	var status sql.NullString
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
	var assignedTo sql.NullString

	if err := scanner.Scan(
		&row.ID,
		&row.IssueCode,
		&status,
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
		&assignedTo,
	); err != nil {
		return issueStageResponse{}, err
	}

	row.Status = dqNullStringPtr(status)
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
	row.AssignedTo = dqNullStringPtr(assignedTo)

	return row, nil
}

func scanValidationRule(scanner interface {
	Scan(dest ...interface{}) error
}) (validationRuleResponse, error) {
	rule, _, err := scanValidationRuleInternal(scanner, false)
	return rule, err
}

func scanValidationRuleWithCreated(scanner interface {
	Scan(dest ...interface{}) error
}) (validationRuleResponse, bool, error) {
	return scanValidationRuleInternal(scanner, true)
}

func scanValidationRuleInternal(scanner interface {
	Scan(dest ...interface{}) error
}, includeCreated bool) (validationRuleResponse, bool, error) {
	var (
		rule        validationRuleResponse
		tableID     sql.NullString
		program     sql.NullString
		category    sql.NullString
		value       sql.NullString
		valueColumn sql.NullString
		createdBy   sql.NullString
		updatedBy   sql.NullString
		deletedAt   sql.NullTime
		created     bool
	)

	dest := []interface{}{
		&rule.ID,
		&tableID,
		&program,
		&category,
		&rule.Code,
		&rule.Severity,
		&rule.Description,
		&rule.Column,
		&rule.Op,
		&value,
		&valueColumn,
		&rule.IsActive,
		&createdBy,
		&updatedBy,
		&rule.CreatedAt,
		&rule.UpdatedAt,
		&deletedAt,
	}

	if includeCreated {
		dest = append(dest, &created)
	}

	if err := scanner.Scan(dest...); err != nil {
		return validationRuleResponse{}, false, err
	}

	rule.TableID = dqNullStringPtr(tableID)
	rule.Program = dqNullStringPtr(program)
	rule.Category = dqNullStringPtr(category)
	rule.Value = dqNullStringPtr(value)
	rule.ValueColumn = dqNullStringPtr(valueColumn)
	rule.CreatedBy = dqNullStringPtr(createdBy)
	rule.UpdatedBy = dqNullStringPtr(updatedBy)
	if deletedAt.Valid {
		rule.DeletedAt = &deletedAt.Time
	}

	return rule, created, nil
}
