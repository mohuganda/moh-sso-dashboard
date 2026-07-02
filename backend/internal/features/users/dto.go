package users

import "time"

type UserResponse struct {
	ID               string              `json:"id"`
	Username         string              `json:"username"`
	Email            string              `json:"email"`
	FullName         string              `json:"fullName"`
	IsAdmin          bool                `json:"isAdmin"`
	RealmRoles       []string            `json:"realmRoles"`
	ClientRoles      map[string][]string `json:"clientRoles"`
	IsActive         bool                `json:"isActive"`
	EmailVerified    bool                `json:"emailVerified"`
	RequirePwdChange bool                `json:"requirePwdChange"`
	LastLoginAt      *time.Time          `json:"lastLoginAt"`
	CreatedAt        *time.Time          `json:"createdAt"`
}

type UserClientRoleAssignmentResponse struct {
	ID         string   `json:"id"`
	ClientID   string   `json:"clientId"`
	ClientName string   `json:"clientName"`
	Roles      []string `json:"roles"`
}

type ClientRoleResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

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
