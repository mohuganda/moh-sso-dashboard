package handler

type HealthComponentStatus struct {
	Database string `json:"database"`
	RemoteDB string `json:"remoteDB"`
	Keycloak string `json:"keycloak"`
	Redis    string `json:"redis"`
}

type HealthResponse struct {
	Status        string                 `json:"status"`
	UptimeSeconds int                    `json:"uptime_seconds"`
	Reason        string                 `json:"reason,omitempty"`
	Service       string                 `json:"service,omitempty"`
	Components    *HealthComponentStatus `json:"components,omitempty"`
}
