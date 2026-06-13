package authz

import "strings"

type Context struct {
	UserID      string              `json:"userId"`
	RealmRoles  []string            `json:"realmRoles"`
	ClientRoles map[string][]string `json:"clientRoles"`
	Permissions []Permission        `json:"permissions"`
	IsAdmin     bool                `json:"isAdmin"`
	IsUser      bool                `json:"isUser"`
}

func NewContext(
	userID string,
	realmRoles []string,
	clientRoles map[string][]string,
) Context {
	normalizedRealmRoles := NormalizeRoles(realmRoles)
	normalizedClientRoles := normalizeClientRoles(clientRoles)

	ctx := Context{
		UserID:      strings.TrimSpace(userID),
		RealmRoles:  normalizedRealmRoles,
		ClientRoles: normalizedClientRoles,
		IsAdmin:     hasRole(normalizedRealmRoles, RoleAdmin),
		IsUser:      hasRole(normalizedRealmRoles, RoleUser),
	}
	ctx.Permissions = PermissionsForRoles(ctx.RealmRoles)

	return ctx
}

func (c Context) HasPermission(permission Permission) bool {
	for _, current := range c.Permissions {
		if current == permission {
			return true
		}
	}

	return false
}

func (c Context) HasAnyPermission(permissions ...Permission) bool {
	for _, permission := range permissions {
		if c.HasPermission(permission) {
			return true
		}
	}

	return false
}

func (c Context) HasRealmRole(role string) bool {
	return hasRole(c.RealmRoles, role)
}

func (c Context) HasClientRole(clientID string, role string) bool {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return false
	}

	return hasRole(c.ClientRoles[clientID], role)
}

func (c Context) PermissionStrings() []string {
	values := make([]string, 0, len(c.Permissions))
	for _, permission := range c.Permissions {
		values = append(values, string(permission))
	}

	return values
}

func normalizeClientRoles(clientRoles map[string][]string) map[string][]string {
	normalized := make(map[string][]string, len(clientRoles))
	for clientID, roles := range clientRoles {
		key := strings.TrimSpace(clientID)
		if key == "" {
			continue
		}
		normalized[key] = NormalizeRoles(roles)
	}

	return normalized
}
