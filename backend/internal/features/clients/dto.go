package clients

type ClientResponse struct {
	ID                                 string            `json:"id"`
	ClientID                           string            `json:"clientId"`
	Name                               string            `json:"name"`
	Description                        string            `json:"description"`
	RootURL                            string            `json:"rootUrl"`
	BaseURL                            string            `json:"baseUrl"`
	AdminURL                           string            `json:"adminUrl"`
	SurrogateAuthRequired              bool              `json:"surrogateAuthRequired"`
	Enabled                            bool              `json:"enabled"`
	AlwaysDisplayInConsole             bool              `json:"alwaysDisplayInConsole"`
	ClientAuthenticatorType            string            `json:"clientAuthenticatorType"`
	RedirectUris                       []string          `json:"redirectUris"`
	WebOrigins                         []string          `json:"webOrigins"`
	NotBefore                          int               `json:"notBefore"`
	BearerOnly                         bool              `json:"bearerOnly"`
	ConsentRequired                    bool              `json:"consentRequired"`
	StandardFlow                       bool              `json:"standardFlowEnabled"`
	ImplicitFlow                       bool              `json:"implicitFlowEnabled"`
	DirectAccess                       bool              `json:"directAccessGrantsEnabled"`
	ServiceAccounts                    bool              `json:"serviceAccountsEnabled"`
	PublicClient                       bool              `json:"publicClient"`
	FrontChannelLogout                 bool              `json:"frontchannelLogout"`
	Protocol                           string            `json:"protocol"`
	FullScopeAllowed                   bool              `json:"fullScopeAllowed"`
	NodeReRegistrationTimeout          int               `json:"nodeReRegistrationTimeout"`
	DefaultClientScopes                []string          `json:"defaultClientScopes"`
	OptionalClientScopes               []string          `json:"optionalClientScopes"`
	Access                             map[string]bool   `json:"access"`
	ClientTemplate                     string            `json:"clientTemplate"`
	UseTemplateConfig                  bool              `json:"useTemplateConfig"`
	RootClientRealm                    string            `json:"rootClientRealm"`
	AuthorizationServicesEnabled       bool              `json:"authorizationServicesEnabled"`
	ProtocolMappers                    []ProtocolMapper  `json:"protocolMappers"`
	DefaultRoles                       []string          `json:"defaultRoles"`
	AuthorizationURL                   string            `json:"authorizationUrl"`
	TlsRequired                        string            `json:"tlsRequired"`
	Attributes                         map[string]string `json:"attributes"`
	AuthenticationFlowBindingOverrides map[string]string `json:"authenticationFlowBindingOverrides"`
	RegisteredNodes                    map[string]int    `json:"registeredNodes"`
}

type ProtocolMapper struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Protocol        string            `json:"protocol"`
	ProtocolMapper  string            `json:"protocolMapper"`
	ConsentRequired bool              `json:"consentRequired"`
	ConsentText     string            `json:"consentText"`
	Config          map[string]string `json:"config"`
}

type ClientRoleResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

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
