package audit

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
