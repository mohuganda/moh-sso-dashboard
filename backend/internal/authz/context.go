package authz

import (
	"context"
	"strings"
)

type Context struct {
	UserID                  string              `json:"userId"`
	RealmRoles              []string            `json:"realmRoles"`
	ClientRoles             map[string][]string `json:"clientRoles"`
	Permissions             []Permission        `json:"permissions"`
	AccessibleSystems       []string            `json:"systems"`
	AccessibleSystemDetails []SystemAccess      `json:"accessibleSystems"`
	IsAdmin                 bool                `json:"isAdmin"`
	IsUser                  bool                `json:"isUser"`
}

func NewContext(
	userID string,
	realmRoles []string,
	clientRoles map[string][]string,
) Context {
	normalizedRealmRoles := NormalizeRoles(realmRoles)
	normalizedClientRoles := normalizeClientRoles(clientRoles)
	access := ResolveAccess(context.Background(), nil, normalizedRealmRoles, normalizedClientRoles)

	ctx := Context{
		UserID:                  strings.TrimSpace(userID),
		RealmRoles:              normalizedRealmRoles,
		ClientRoles:             normalizedClientRoles,
		Permissions:             dedupePermissions(access.Permissions),
		AccessibleSystems:       systemIDs(access.Systems),
		AccessibleSystemDetails: access.Systems,
		IsAdmin:                 hasRole(normalizedRealmRoles, RoleAdmin),
		IsUser:                  hasRole(normalizedRealmRoles, RoleUser),
	}

	return ctx
}

func NewContextWithResolver(
	ctx context.Context,
	resolver PermissionResolver,
	userID string,
	realmRoles []string,
	clientRoles map[string][]string,
) Context {
	normalizedRealmRoles := NormalizeRoles(realmRoles)
	normalizedClientRoles := normalizeClientRoles(clientRoles)
	access := ResolveAccess(ctx, resolver, normalizedRealmRoles, normalizedClientRoles)

	return Context{
		UserID:                  strings.TrimSpace(userID),
		RealmRoles:              normalizedRealmRoles,
		ClientRoles:             normalizedClientRoles,
		Permissions:             dedupePermissions(access.Permissions),
		AccessibleSystems:       systemIDs(access.Systems),
		AccessibleSystemDetails: access.Systems,
		IsAdmin:                 hasRole(normalizedRealmRoles, RoleAdmin),
		IsUser:                  hasRole(normalizedRealmRoles, RoleUser),
	}
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

func (c Context) HasSystem(system string) bool {
	system = strings.TrimSpace(system)
	if system == "" {
		return false
	}

	for _, current := range c.AccessibleSystems {
		if strings.EqualFold(current, system) {
			return true
		}
	}

	return false
}

func (c Context) HasSystemRole(system string, role string) bool {
	return c.HasClientRole(system, role)
}

func (c Context) PermissionStrings() []string {
	values := make([]string, 0, len(c.Permissions))
	for _, permission := range c.Permissions {
		values = append(values, string(permission))
	}

	return values
}

func (c Context) SystemStrings() []string {
	values := make([]string, 0, len(c.AccessibleSystems))
	values = append(values, c.AccessibleSystems...)
	return values
}

func (c Context) SystemAccess() []SystemAccess {
	values := make([]SystemAccess, 0, len(c.AccessibleSystemDetails))
	values = append(values, c.AccessibleSystemDetails...)
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

func dedupePermissions(permissions []Permission) []Permission {
	seen := map[Permission]bool{}
	values := make([]Permission, 0, len(permissions))
	for _, permission := range permissions {
		if permission == "" || seen[permission] {
			continue
		}
		seen[permission] = true
		values = append(values, permission)
	}
	return values
}

func systemIDs(systems []SystemAccess) []string {
	seen := map[string]bool{}
	values := make([]string, 0, len(systems))
	for _, system := range systems {
		clientID := strings.TrimSpace(system.ClientID)
		if clientID == "" || seen[clientID] {
			continue
		}
		seen[clientID] = true
		values = append(values, clientID)
	}
	return values
}
