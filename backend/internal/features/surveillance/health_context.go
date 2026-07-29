package surveillance

import (
	"context"

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
