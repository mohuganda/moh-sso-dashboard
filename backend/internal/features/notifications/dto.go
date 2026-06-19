package notifications

type CountResponse struct {
	Count int64 `json:"count"`
}

type DeliveryRetryResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type NotificationDeliveryResponse struct {
	ID             string         `json:"id"`
	NotificationID string         `json:"notification_id"`
	Channel        string         `json:"channel"`
	Status         string         `json:"status"`
	Recipient      map[string]any `json:"recipient,omitempty"`
	TemplateName   string         `json:"template_name,omitempty"`
	TemplateData   map[string]any `json:"template_data,omitempty"`
	Payload        map[string]any `json:"payload,omitempty"`
	ScheduledAt    string         `json:"scheduled_at,omitempty"`
	LockedAt       string         `json:"locked_at,omitempty"`
	SentAt         string         `json:"sent_at,omitempty"`
	Attempts       int32          `json:"attempts"`
	MaxAttempts    int32          `json:"max_attempts"`
	LastError      string         `json:"last_error,omitempty"`
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
}

type NotificationResponse struct {
	ID         string         `json:"id"`
	Type       string         `json:"type,omitempty"`
	Title      string         `json:"title"`
	Message    string         `json:"message"`
	Severity   string         `json:"severity"`
	TargetRole string         `json:"target_role,omitempty"`
	ClientID   string         `json:"client_id,omitempty"`
	UserID     string         `json:"user_id,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	Read       bool           `json:"read"`
	CreatedAt  string         `json:"created_at"`
}
