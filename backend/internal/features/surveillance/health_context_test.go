package surveillance

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	healthcontext "github.com/moh-sso-dashboard/internal/features/health_context"
)

type scopeRepository struct {
	healthcontext.Repository
	aliases       []healthcontext.Alias
	parent        healthcontext.Node
	parentAliases []healthcontext.Alias
	failure       error
}

func (r *scopeRepository) GetNode(_ context.Context, id uuid.UUID) (healthcontext.Node, error) {
	return healthcontext.Node{ID: id, Enabled: true}, nil
}
func (r *scopeRepository) ListAliasesInScope(context.Context, uuid.UUID, bool) ([]healthcontext.Alias, error) {
	return r.aliases, r.failure
}
func (r *scopeRepository) ListAncestors(context.Context, uuid.UUID) ([]healthcontext.Node, error) {
	return []healthcontext.Node{r.parent}, nil
}
func (r *scopeRepository) ListAliases(context.Context, uuid.UUID) ([]healthcontext.Alias, error) {
	return r.parentAliases, nil
}

func TestScopeAuthorizationIDsSeparateHierarchyFromDataAccess(t *testing.T) {
	facility, parentRegion := uuid.New(), uuid.New()
	repo := &scopeRepository{
		aliases: []healthcontext.Alias{
			{Namespace: healthContextNamespaceFacility, ExternalID: facility.String()},
			{Namespace: healthContextNamespaceFacility, ExternalID: facility.String()},
		},
		parent:        healthcontext.Node{ID: uuid.New()},
		parentAliases: []healthcontext.Alias{{Namespace: healthContextNamespaceRegion, ExternalID: parentRegion.String()}},
	}
	scope := HealthContextScope{ID: uuid.New(), service: healthcontext.NewService(repo)}
	ids, err := scope.allowedIDs(context.Background(), healthContextNamespaceFacility, false)
	if err != nil || len(ids) != 1 || ids[0] != facility {
		t.Fatalf("facility scope: %v %v", ids, err)
	}
	ids, err = scope.allowedIDs(context.Background(), healthContextNamespaceRegion, false)
	if err != nil || len(ids) != 0 {
		t.Fatalf("metric scope included an ancestor: %v %v", ids, err)
	}
	ids, err = scope.allowedIDs(context.Background(), healthContextNamespaceRegion, true)
	if err != nil || len(ids) != 1 || ids[0] != parentRegion {
		t.Fatalf("hierarchy selector lost ancestor: %v %v", ids, err)
	}
	repo.failure = errors.New("primary database unavailable")
	if _, err := scope.allowedIDs(context.Background(), healthContextNamespaceFacility, false); err == nil {
		t.Fatal("resolver failure must fail closed")
	}
	if _, err := (HealthContextScope{}).allowedIDs(context.Background(), healthContextNamespaceFacility, false); err == nil {
		t.Fatal("missing resolver must fail closed")
	}
}
