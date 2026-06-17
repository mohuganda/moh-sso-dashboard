package audit

import (
	"time"

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
