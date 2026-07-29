package health_context

import (
	"context"

	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/authz"
)

type Policy struct {
	service *Service
}

func NewPolicy(service *Service) *Policy {
	return &Policy{service: service}
}

type AuthorizationInput struct {
	UserID             string
	Permissions        []authz.Permission
	RequiredPermission authz.Permission
	Systems            []string
	RequiredSystem     string
	ContextID          uuid.UUID
}

type AliasAuthorizationInput struct {
	UserID             string
	Permissions        []authz.Permission
	RequiredPermission authz.Permission
	Systems            []string
	RequiredSystem     string
	Namespace          string
	ExternalID         string
}

func (p *Policy) Authorize(ctx context.Context, input AuthorizationInput) (AccessExplanation, error) {
	hasPermission := input.RequiredPermission == ""
	for _, permission := range input.Permissions {
		if permission == input.RequiredPermission {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return AccessExplanation{Reason: "required functional permission is missing"}, ErrContextForbidden
	}

	hasSystem := input.RequiredSystem == ""
	for _, system := range input.Systems {
		if system == input.RequiredSystem {
			hasSystem = true
			break
		}
	}
	if !hasSystem {
		return AccessExplanation{Reason: "required system access is missing"}, ErrContextForbidden
	}

	explanation, err := p.service.ExplainAccess(ctx, input.UserID, input.ContextID)
	if err != nil {
		return AccessExplanation{}, err
	}
	if !explanation.Allowed {
		return explanation, ErrContextForbidden
	}
	return explanation, nil
}

func (p *Policy) AuthorizeAlias(
	ctx context.Context,
	input AliasAuthorizationInput,
) (AccessExplanation, error) {
	node, err := p.service.ResolveAlias(ctx, input.Namespace, input.ExternalID)
	if err != nil {
		return AccessExplanation{}, err
	}
	return p.Authorize(ctx, AuthorizationInput{
		UserID:             input.UserID,
		Permissions:        input.Permissions,
		RequiredPermission: input.RequiredPermission,
		Systems:            input.Systems,
		RequiredSystem:     input.RequiredSystem,
		ContextID:          node.ID,
	})
}
