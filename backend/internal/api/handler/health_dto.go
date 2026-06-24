package handler

import "github.com/moh-sso-dashboard/internal/version"

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
	Version       string                 `json:"version,omitempty"`
	Build         *version.BuildInfo     `json:"build,omitempty"`
}
