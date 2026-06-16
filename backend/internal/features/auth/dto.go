package auth

type AuthMeResponse struct {
	User any `json:"user"`
}

type AuthRefreshCookieDebug struct {
	Environment      string `json:"environment"`
	Secure           bool   `json:"secure"`
	SameSite         string `json:"same_site"`
	Domain           string `json:"domain"`
	HasAccessToken   bool   `json:"has_access_token"`
	HasRefreshToken  bool   `json:"has_refresh_token"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	FrontendBaseURL  string `json:"frontend_base_url"`
	KeycloakRedirect string `json:"keycloak_redirect"`
	KeycloakExternal string `json:"keycloak_external"`
	KeycloakInternal string `json:"keycloak_internal"`
}

type AuthRefreshResponse struct {
	ExpiresIn   int64                  `json:"expires_in"`
	CookieDebug AuthRefreshCookieDebug `json:"cookie_debug"`
}
