package notifications

import (
	"encoding/json"
	"time"

	"github.com/moh-sso-dashboard/internal/model"
)

func toNotificationResponse(notification model.Notification) NotificationResponse {
	return NotificationResponse{
		ID:         notification.ID.String(),
		Type:       notification.Type,
		Title:      notification.Title,
		Message:    notification.Message,
		Severity:   notification.Severity,
		TargetRole: notification.TargetRole,
		ClientID:   notification.ClientID,
		UserID:     notification.UserID,
		Metadata:   toNotificationMetadata(notification.Metadata),
		Read:       notification.Read,
		CreatedAt:  notification.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func toNotificationResponses(notifications []model.Notification) []NotificationResponse {
	out := make([]NotificationResponse, 0, len(notifications))
	for _, notification := range notifications {
		out = append(out, toNotificationResponse(notification))
	}
	return out
}

func toNotificationMetadata(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return nil
	}

	var metadata map[string]any
	if err := json.Unmarshal(raw, &metadata); err != nil {
		return nil
	}
	return metadata
}
