package notifications

type CountResponse struct {
	Count int64 `json:"count"`
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
