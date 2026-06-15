package model

import "time"

type NotificationChannel string

const (
	NotificationChannelInApp   NotificationChannel = "in_app"
	NotificationChannelEmail   NotificationChannel = "email"
	NotificationChannelSMS     NotificationChannel = "sms"
	NotificationChannelWebhook NotificationChannel = "webhook"
)

type NotificationDeliveryStatus string

const (
	NotificationDeliveryPending    NotificationDeliveryStatus = "PENDING"
	NotificationDeliveryProcessing NotificationDeliveryStatus = "PROCESSING"
	NotificationDeliverySent       NotificationDeliveryStatus = "SENT"
	NotificationDeliveryFailed     NotificationDeliveryStatus = "FAILED"
	NotificationDeliveryRetry      NotificationDeliveryStatus = "RETRY"
	NotificationDeliveryCancelled  NotificationDeliveryStatus = "CANCELLED"
)

type CreateNotificationRequest struct {
	Type       string         `json:"type"`
	Title      string         `json:"title"`
	Message    string         `json:"message"`
	Severity   string         `json:"severity"`
	TargetRole string         `json:"target_role"`
	Metadata   map[string]any `json:"metadata,omitempty"`

	Deliveries []NotificationDeliveryRequest `json:"deliveries,omitempty"`
}

type NotificationDeliveryRequest struct {
	Channel      NotificationChannel `json:"channel"`
	Recipient    map[string]any      `json:"recipient,omitempty"`
	TemplateName string              `json:"template_name,omitempty"`
	TemplateData map[string]any      `json:"template_data,omitempty"`
	Payload      map[string]any      `json:"payload,omitempty"`
	Attachments  []Attachment        `json:"attachments,omitempty"`
	ScheduledAt  *time.Time          `json:"scheduled_at,omitempty"`
	MaxAttempts  int32               `json:"max_attempts,omitempty"`
}
