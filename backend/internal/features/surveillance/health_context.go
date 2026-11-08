package surveillance

import (
	"context"
	"errors"

	"github.com/google/uuid"
	healthcontext "github.com/moh-sso-dashboard/internal/features/health_context"
)

const (
	healthContextNamespaceRegion    = "surveillance-region"
	healthContextNamespaceDistrict  = "surveillance-district"
	healthContextNamespaceSubCounty = "surveillance-sub-county"
	healthContextNamespaceFacility  = "surveillance-facility"
)

type HealthContextScope struct {
	service            *healthcontext.Service
	ID                 uuid.UUID
	IncludeDescendants bool
}

func (s HealthContextScope) containsAlias(
	ctx context.Context,
	service *healthcontext.Service,
	namespace string,
	externalID uuid.UUID,
) (bool, error) {
	node, err := service.ResolveAlias(ctx, namespace, externalID.String())
	if err != nil {
		return false, err
	}
	if node.ID == s.ID {
		return true, nil
	}
	if !s.IncludeDescendants {
		return false, nil
	}
	return service.IsDescendant(ctx, s.ID, node.ID)
}

// Authorization remains in the primary database; observatory receives only IDs.
func (s HealthContextScope) allowedIDs(ctx context.Context, namespace string, related bool) ([]uuid.UUID, error) {
	if s.service == nil {
		return nil, errors.New("health context resolver is unavailable")
	}
	aliases, err := s.service.ListAliasesInScope(ctx, s.ID, s.IncludeDescendants)
	if err != nil {
		return nil, err
	}
	if related {
		ancestors, err := s.service.ListAncestors(ctx, s.ID)
		if err != nil {
			return nil, err
		}
		for _, node := range ancestors {
			items, err := s.service.ListAliases(ctx, node.ID)
			if err != nil {
				return nil, err
			}
			aliases = append(aliases, items...)
		}
	}
	ids := make([]uuid.UUID, 0)
	seen := map[uuid.UUID]bool{}
	for _, alias := range aliases {
		if alias.Namespace != namespace {
			continue
		}
		id, err := uuid.Parse(alias.ExternalID)
		if err != nil {
			return nil, err
		}
		if !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	return ids, nil
}
func (s HealthContextScope) isNational(ctx context.Context) (bool, error) {
	if s.service == nil {
		return false, errors.New("health context resolver is unavailable")
	}
	node, err := s.service.GetNode(ctx, s.ID)
	return node.Code == "UG", err
}
