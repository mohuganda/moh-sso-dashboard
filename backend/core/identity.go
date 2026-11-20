package core

import (
	"context"
	"errors"
	"strings"
	"time"
)

// Identity contains verified subject/role claims, never bearer or refresh
// tokens. The host must authenticate credentials before constructing it.
type Identity struct {
	Subject     string
	RealmRoles  []string
	ClientRoles map[string][]string
}

// SessionStore describes host-owned opaque-session persistence. T is defined
// by the trusted host; core declares no credential payload or storage provider.
type SessionStore[T any] interface {
	Create(context.Context, T, time.Duration) (string, error)
	Get(context.Context, string) (*T, error)
	Update(context.Context, string, T, time.Duration) error
	Delete(context.Context, string) error
}

var ErrSessionNotFound = errors.New("auth session not found")

// Access carries only the explicit permission/system decision of a trusted
// resolver. Core does not derive admin grants, realm defaults or fallback rules.
type Access struct {
	Permissions []string
	Systems     []SystemAccess
}

type SystemAccess struct {
	ClientID               string   `json:"clientId"`
	DisplayName            string   `json:"displayName"`
	LaunchURL              string   `json:"launchUrl,omitempty"`
	AuthenticatedLaunchURL string   `json:"authenticatedLaunchUrl,omitempty"`
	Icon                   string   `json:"icon,omitempty"`
	Category               string   `json:"category,omitempty"`
	Navigation             string   `json:"navigation,omitempty"`
	SystemType             string   `json:"systemType"`
	DisplayInLauncher      bool     `json:"displayInLauncher"`
	DisplayInSideNav       bool     `json:"displayInSideNav"`
	LaunchMode             string   `json:"launchMode"`
	Roles                  []string `json:"roles"`
}

type AuthorizationResolver interface {
	Resolve(context.Context, Identity) (Access, error)
}

var ErrAuthorizationUnavailable = errors.New("authorization resolver is not configured")

// ResolveAuthorization discards partial grants on failure. Provider selection
// and any deliberate static policy belong to trusted host composition.
func ResolveAuthorization(ctx context.Context, resolver AuthorizationResolver, identity Identity) (Access, error) {
	if resolver == nil {
		return Access{}, ErrAuthorizationUnavailable
	}
	access, err := resolver.Resolve(ctx, identity)
	if err != nil {
		return Access{}, err
	}
	return access, nil
}

func HasPermission[P ~string](permissions []P, permission P) bool {
	for _, current := range permissions {
		if current == permission {
			return true
		}
	}
	return false
}
func HasAnyPermission[P ~string](permissions []P, requested ...P) bool {
	for _, permission := range requested {
		if HasPermission(permissions, permission) {
			return true
		}
	}
	return false
}

// MissingPermissions returns unmet requirements in their original order.
func MissingPermissions[P ~string](granted []P, required ...P) []P {
	var missing []P
	for _, permission := range required {
		if !HasPermission(granted, permission) {
			missing = append(missing, permission)
		}
	}
	return missing
}

func HasRole(roles []string, role string) bool {
	role = strings.ToLower(strings.TrimSpace(role))
	if role == "" {
		return false
	}
	for _, current := range roles {
		if strings.EqualFold(strings.TrimSpace(current), role) {
			return true
		}
	}
	return false
}
func HasSystem(systems []string, system string) bool {
	system = strings.TrimSpace(system)
	if system == "" {
		return false
	}
	for _, current := range systems {
		if strings.EqualFold(current, system) {
			return true
		}
	}
	return false
}
