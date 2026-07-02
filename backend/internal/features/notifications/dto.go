package notifications

type CountResponse struct {
	Count int64 `json:"count"`
}

type DeliveryRetryResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type DeliveryActionResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type NotificationDeliveryListResponse struct {
	Items  []NotificationDeliveryResponse `json:"items"`
	Total  int64                          `json:"total"`
	Limit  int32                          `json:"limit"`
	Offset int32                          `json:"offset"`
}

type NotificationDeliveryMetricsResponse struct {
	Items []NotificationDeliveryMetricResponse `json:"items"`
}

type NotificationDeliveryMetricResponse struct {
	Channel              string  `json:"channel"`
	Provider             string  `json:"provider,omitempty"`
	Total                int64   `json:"total"`
	Pending              int64   `json:"pending"`
	Processing           int64   `json:"processing"`
	Sent                 int64   `json:"sent"`
	Failed               int64   `json:"failed"`
	Retry                int64   `json:"retry"`
	Cancelled            int64   `json:"cancelled"`
	AvgProcessingSeconds float64 `json:"avg_processing_seconds"`
}

type TestSMSRequest struct {
	To      string `json:"to" binding:"required"`
	Message string `json:"message" binding:"required"`
}

type TestSMSResponse struct {
	NotificationID string `json:"notification_id"`
	Status         string `json:"status"`
}

type NotificationPreferencesResponse struct {
	UserID          string `json:"user_id"`
	EmailEnabled    bool   `json:"email_enabled"`
	SMSEnabled      bool   `json:"sms_enabled"`
	PhoneNumber     string `json:"phone_number,omitempty"`
	PhoneVerified   bool   `json:"phone_verified"`
	QuietHoursStart string `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd   string `json:"quiet_hours_end,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
	UpdatedAt       string `json:"updated_at,omitempty"`
}

type UpdateNotificationPreferencesRequest struct {
	EmailEnabled    *bool   `json:"email_enabled"`
	SMSEnabled      *bool   `json:"sms_enabled"`
	PhoneNumber     *string `json:"phone_number"`
	QuietHoursStart *string `json:"quiet_hours_start"`
	QuietHoursEnd   *string `json:"quiet_hours_end"`
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
	Provider       string         `json:"provider,omitempty"`
	ProviderMsgID  string         `json:"provider_message_id,omitempty"`
	ProviderStatus string         `json:"provider_status,omitempty"`
	ProviderMeta   map[string]any `json:"provider_metadata,omitempty"`
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
