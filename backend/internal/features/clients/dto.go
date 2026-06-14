package clients

type toggleClientEnabledRequest struct {
	Enabled bool `json:"enabled"`
}

type UpdateClientRequest struct {
	Name         string            `json:"name" validate:"required"`
	Description  string            `json:"description,omitempty"`
	Icon         string            `json:"icon,omitempty"`
	PublicClient bool              `json:"publicClient"`
	Enabled      bool              `json:"enabled"`
	RootURL      string            `json:"rootUrl"`
	BaseURL      string            `json:"baseUrl"`
	AdminURL     string            `json:"adminUrl"`
	RedirectURIs []string          `json:"redirectUris"`
	WebOrigins   []string          `json:"webOrigins"`
	Attributes   map[string]string `json:"attributes"`
}
