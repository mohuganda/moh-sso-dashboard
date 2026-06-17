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
		timePeriod   sql.NullString
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
		&timePeriod,
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
		TimePeriod:   dqNullStringPtr(timePeriod),
	}, nil
}
