package authz

import (
	"context"
	"errors"
)

type PermissionResolver interface {
	Resolve(ctx context.Context, realmRoles []string, clientRoles map[string][]string) (ResolvedAccess, error)
}

type ResolvedAccess struct {
	Permissions []Permission   `json:"permissions"`
	Systems     []SystemAccess `json:"systems"`
}

type SystemAccess struct {
	ClientID    string   `json:"clientId"`
	DisplayName string   `json:"displayName"`
	LaunchURL   string   `json:"launchUrl,omitempty"`
	Icon        string   `json:"icon,omitempty"`
	Category    string   `json:"category,omitempty"`
	Navigation  string   `json:"navigation,omitempty"`
	Roles       []string `json:"roles"`
}

type StaticResolver struct{}

func NewStaticResolver() StaticResolver {
	return StaticResolver{}
}

func (StaticResolver) Resolve(
	_ context.Context,
	realmRoles []string,
	clientRoles map[string][]string,
) (ResolvedAccess, error) {
	return ResolvedAccess{
		Permissions: PermissionsForContext(realmRoles, clientRoles),
		Systems:     AccessibleSystemDetailsForContext(clientRoles),
	}, nil
}

type CompositeResolver struct {
	Primary  PermissionResolver
	Fallback PermissionResolver
}

func NewCompositeResolver(primary PermissionResolver, fallback PermissionResolver) CompositeResolver {
	if fallback == nil {
		fallback = NewStaticResolver()
	}

	return CompositeResolver{
		Primary:  primary,
		Fallback: fallback,
	}
}

func (r CompositeResolver) Resolve(
	ctx context.Context,
	realmRoles []string,
	clientRoles map[string][]string,
) (ResolvedAccess, error) {
	if r.Primary != nil {
		access, err := r.Primary.Resolve(ctx, realmRoles, clientRoles)
		if err == nil && (len(access.Permissions) > 0 || len(access.Systems) > 0) {
			return access, nil
		}
	}

	if r.Fallback == nil {
		return ResolvedAccess{}, errors.New("authz resolver is not configured")
	}

	return r.Fallback.Resolve(ctx, realmRoles, clientRoles)
}

func ResolveAccess(
	ctx context.Context,
	resolver PermissionResolver,
	realmRoles []string,
	clientRoles map[string][]string,
) ResolvedAccess {
	if resolver == nil {
		resolver = NewStaticResolver()
	}

	access, err := resolver.Resolve(ctx, realmRoles, clientRoles)
	if err != nil {
		return ResolvedAccess{
			Permissions: PermissionsForContext(realmRoles, clientRoles),
			Systems:     AccessibleSystemDetailsForContext(clientRoles),
		}
	}

	return access
}
