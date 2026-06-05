package users

type userClientRolesForClientRequest struct {
	ClientID   string `json:"clientId"`
	ClientUUID string `json:"clientUuid"`
}

type userClientRolesRequest struct {
	ClientID   string   `json:"clientId"`
	ClientUUID string   `json:"clientUuid"`
	Roles      []string `json:"roles"`
}

type setUserEnabledRequest struct {
	Enabled bool `json:"enabled"`
}
