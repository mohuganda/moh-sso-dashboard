package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

func toAuditLogResponse(row db.ListAuditLogsRow) AuditLogResponse {
	return auditLogResponse(row.ID.String(), row.CreatedAt.Valid, row.CreatedAt.Time, row.UserID.Valid, row.UserID.UUID.String(), row.Username, row.Action, row.Metadata)
}

func toAuditLogResponseFromGet(row db.GetAuditLogRow) AuditLogResponse {
	return auditLogResponse(row.ID.String(), row.CreatedAt.Valid, row.CreatedAt.Time, row.UserID.Valid, row.UserID.UUID.String(), row.Username, row.Action, row.Metadata)
}

func toAuditLogResponseFromExport(row db.ExportAuditLogsRow) AuditLogResponse {
	return auditLogResponse(row.ID.String(), row.CreatedAt.Valid, row.CreatedAt.Time, row.UserID.Valid, row.UserID.UUID.String(), row.Username, row.Action, row.Metadata)
}

func toAuditLogResponses(rows []db.ListAuditLogsRow) []AuditLogResponse {
	out := make([]AuditLogResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toAuditLogResponse(row))
	}
	return out
}

func toAuditActionsResponse(rows []string) AuditActionsResponse {
	return AuditActionsResponse{Actions: rows}
}

func toAuditCursorResponse(nextCreatedAt *time.Time, nextID *uuid.UUID) AuditCursor {
	return AuditCursor{
		CursorCreatedAt: nextCreatedAt,
		CursorID:        nextID,
	}
}

func toAuditMetricsOverviewResponse(row db.AuditMetricsOverviewRow) AuditMetricsOverviewResponse {
	return AuditMetricsOverviewResponse{
		TotalEvents:      row.TotalEvents,
		TotalFailures:    row.TotalFailures,
		FailedLogins:     row.FailedLogins,
		SuccessfulLogins: row.SuccessfulLogins,
	}
}

func toAuditFailedLoginsByDayResponse(rows []db.FailedLoginsByDayRow) AuditFailedLoginsByDayResponse {
	out := make([]AuditFailedLoginsByDayPoint, 0, len(rows))
	for _, row := range rows {
		out = append(out, AuditFailedLoginsByDayPoint{Day: row.Day, Count: row.Count})
	}
	return AuditFailedLoginsByDayResponse{Series: out}
}

func toAuditTopFailureIPsResponse(rows []db.TopFailureIPsRow) AuditTopFailureIPsResponse {
	out := make([]AuditTopFailureIPPoint, 0, len(rows))
	for _, row := range rows {
		out = append(out, AuditTopFailureIPPoint{
			IP:    toAuditString(row.Ip),
			Count: row.Count,
		})
	}
	return AuditTopFailureIPsResponse{Items: out}
}

func toAuditExportResponse(manifest AuditExportManifest, items []AuditLogResponse) AuditExportResponse {
	return AuditExportResponse{Manifest: manifest, Items: items}
}

func toAuditLogListResponse(items []AuditLogResponse, nextCursor AuditCursor, hasMore bool) AuditLogListResponse {
	return AuditLogListResponse{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}
}

func toAuditExportManifest(from, to time.Time, rows any, recordCount int) AuditExportManifest {
	payload, _ := json.Marshal(rows)
	sum := sha256.Sum256(payload)
	return AuditExportManifest{
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Range:         map[string]string{"from": from.UTC().Format(time.RFC3339), "to": to.UTC().Format(time.RFC3339)},
		RecordCount:   recordCount,
		HashAlgo:      "sha256",
		PayloadSHA256: hex.EncodeToString(sum[:]),
	}
}

func toExportAuditLogResponses(rows []db.ExportAuditLogsRow) []AuditLogResponse {
	out := make([]AuditLogResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toAuditLogResponseFromExport(row))
	}
	return out
}

func auditLogResponse(
	id string,
	createdAtValid bool,
	createdAt time.Time,
	userIDValid bool,
	userID string,
	username string,
	action string,
	rawMetadata pqtype.NullRawMessage,
) AuditLogResponse {
	metadata := extractAuditMetadata(rawMetadata)
	if metadata == nil {
		metadata = map[string]any{}
	}

	response := AuditLogResponse{
		ID:       id,
		Username: username,
		Action:   action,
		Metadata: metadata,
		IP:       stringValue(metadata["ip"]),
		ClientID: stringValue(metadata["client_id"]),
		Success:  boolPtr(metadata["success"]),
	}
	if createdAtValid {
		response.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	}
	if userIDValid {
		response.UserID = userID
	}

	return response
}

func extractAuditMetadata(r pqtype.NullRawMessage) map[string]any {
	if !r.Valid || len(r.RawMessage) == 0 {
		return nil
	}

	var meta map[string]any
	if err := json.Unmarshal(r.RawMessage, &meta); err != nil {
		return nil
	}
	return meta
}

func stringValue(value any) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func boolPtr(value any) *bool {
	switch v := value.(type) {
	case bool:
		return &v
	case string:
		if v == "true" {
			out := true
			return &out
		}
		if v == "false" {
			out := false
			return &out
		}
	}
	return nil
}

func toAuditString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case uuid.UUID:
		return v.String()
	default:
		return ""
	}
}
