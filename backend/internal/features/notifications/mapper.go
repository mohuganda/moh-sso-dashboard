package notifications

import (
	"database/sql"
	"encoding/json"
	"time"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/service"
	"github.com/sqlc-dev/pqtype"
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

func toNotificationDeliveryResponse(delivery db.NotificationDelivery) NotificationDeliveryResponse {
	return NotificationDeliveryResponse{
		ID:             delivery.ID.String(),
		NotificationID: delivery.NotificationID.String(),
		Channel:        delivery.Channel,
		Status:         delivery.Status,
		Recipient:      toNullRawMessageMap(delivery.Recipient),
		TemplateName:   nullStringValue(delivery.TemplateName),
		TemplateData:   toNullRawMessageMap(delivery.TemplateData),
		Payload:        toNullRawMessageMap(delivery.Payload),
		ScheduledAt:    nullTimeValue(delivery.ScheduledAt),
		LockedAt:       nullTimeValue(delivery.LockedAt),
		SentAt:         nullTimeValue(delivery.SentAt),
		Attempts:       delivery.Attempts,
		MaxAttempts:    delivery.MaxAttempts,
		LastError:      nullStringValue(delivery.LastError),
		Provider:       nullStringValue(delivery.Provider),
		ProviderMsgID:  nullStringValue(delivery.ProviderMessageID),
		ProviderStatus: nullStringValue(delivery.ProviderStatus),
		ProviderMeta:   toNullRawMessageMap(delivery.ProviderResponse),
		CreatedAt:      delivery.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:      delivery.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toNotificationDeliveryResponses(deliveries []db.NotificationDelivery) []NotificationDeliveryResponse {
	out := make([]NotificationDeliveryResponse, 0, len(deliveries))
	for _, delivery := range deliveries {
		out = append(out, toNotificationDeliveryResponse(delivery))
	}
	return out
}

func toNotificationDeliveryMetricResponse(
	metric db.ListNotificationDeliveryMetricsRow,
) NotificationDeliveryMetricResponse {
	return NotificationDeliveryMetricResponse{
		Channel:              metric.Channel,
		Provider:             metric.Provider,
		Total:                metric.Total,
		Pending:              metric.Pending,
		Processing:           metric.Processing,
		Sent:                 metric.Sent,
		Failed:               metric.Failed,
		Retry:                metric.Retry,
		Cancelled:            metric.Cancelled,
		AvgProcessingSeconds: metric.AvgProcessingSeconds,
	}
}

func toNotificationDeliveryMetricResponses(
	metrics []db.ListNotificationDeliveryMetricsRow,
) []NotificationDeliveryMetricResponse {
	out := make([]NotificationDeliveryMetricResponse, 0, len(metrics))
	for _, metric := range metrics {
		out = append(out, toNotificationDeliveryMetricResponse(metric))
	}
	return out
}

func toNotificationPreferencesResponse(
	preferences service.NotificationPreferences,
) NotificationPreferencesResponse {
	return NotificationPreferencesResponse{
		UserID:          preferences.UserID,
		EmailEnabled:    preferences.EmailEnabled,
		SMSEnabled:      preferences.SMSEnabled,
		PhoneNumber:     preferences.PhoneNumber,
		PhoneVerified:   preferences.PhoneVerified,
		QuietHoursStart: preferences.QuietHoursStart,
		QuietHoursEnd:   preferences.QuietHoursEnd,
		CreatedAt:       timeValue(preferences.CreatedAt),
		UpdatedAt:       timeValue(preferences.UpdatedAt),
	}
}

func toNullRawMessageMap(raw pqtype.NullRawMessage) map[string]any {
	if !raw.Valid || len(raw.RawMessage) == 0 {
		return nil
	}

	var value map[string]any
	if err := json.Unmarshal(raw.RawMessage, &value); err != nil {
		return nil
	}

	return value
}

func nullStringValue(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func nullTimeValue(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return value.Time.UTC().Format(time.RFC3339)
}

func timeValue(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.UTC().Format(time.RFC3339)
}
