package audit

import (
	"time"

	"github.com/google/uuid"
)

type AuditLogResponse struct {
	ID        string         `json:"id"`
	CreatedAt string         `json:"createdAt"`
	UserID    string         `json:"userId,omitempty"`
	Username  string         `json:"username"`
	Action    string         `json:"action"`
	Metadata  map[string]any `json:"metadata"`
	IP        string         `json:"ip,omitempty"`
	ClientID  string         `json:"clientId,omitempty"`
	Success   *bool          `json:"success,omitempty"`
}

type AuditActionsResponse struct {
	Actions []string `json:"actions"`
}

type AuditCursor struct {
	CursorCreatedAt *time.Time `json:"cursor_created_at,omitempty"`
	CursorID        *uuid.UUID `json:"cursor_id,omitempty"`
}

type AuditLogListResponse struct {
	Items      []AuditLogResponse `json:"items"`
	NextCursor AuditCursor        `json:"next_cursor"`
	HasMore    bool               `json:"has_more"`
}

type AuditMetricsOverviewResponse struct {
	TotalEvents      int64 `json:"total_events"`
	TotalFailures    int64 `json:"total_failures"`
	FailedLogins     int64 `json:"failed_logins"`
	SuccessfulLogins int64 `json:"successful_logins"`
}

type AuditFailedLoginsByDayPoint struct {
	Day   int64 `json:"day"`
	Count int64 `json:"count"`
}

type AuditFailedLoginsByDayResponse struct {
	Series []AuditFailedLoginsByDayPoint `json:"series"`
}

type AuditTopFailureIPPoint struct {
	IP    string `json:"ip"`
	Count int64  `json:"count"`
}

type AuditTopFailureIPsResponse struct {
	Items []AuditTopFailureIPPoint `json:"items"`
}

type AuditExportManifest struct {
	GeneratedAt   string            `json:"generated_at"`
	Range         map[string]string `json:"range"`
	RecordCount   int               `json:"record_count"`
	HashAlgo      string            `json:"hash_algo"`
	PayloadSHA256 string            `json:"payload_sha256"`
}

type AuditExportResponse struct {
	Manifest AuditExportManifest `json:"manifest"`
	Items    []AuditLogResponse  `json:"items"`
}

type AuditMetricsSummary struct {
	TotalEvents      int64 `json:"total_events"`
	TotalFailures    int64 `json:"total_failures"`
	FailedLogins     int64 `json:"failed_logins"`
	SuccessfulLogins int64 `json:"successful_logins"`
}

type AuditCountResponse struct {
	Count int64 `json:"count"`
}

type AuditWindowResponse struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}
