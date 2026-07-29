package health_context

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/authz"
)

type fakeRepository struct {
	nodes                      map[uuid.UUID]Node
	aliases                    map[[2]string]uuid.UUID
	effective                  []EffectiveContext
	descendants                map[[2]uuid.UUID]bool
	active                     *uuid.UUID
	drift                      []GroupMappingDrift
	applied                    []GroupContextSyncMapping
	audienceUserIDs            []string
	audienceContextIDs         []uuid.UUID
	audienceIncludeDescendants bool
}

func (f *fakeRepository) ListAudienceUserIDs(
	_ context.Context,
	contextIDs []uuid.UUID,
	includeDescendants bool,
) ([]string, error) {
	f.audienceContextIDs = append([]uuid.UUID(nil), contextIDs...)
	f.audienceIncludeDescendants = includeDescendants
	return append([]string(nil), f.audienceUserIDs...), nil
}

func (f *fakeRepository) ListGroupMappingDrift(
	context.Context,
) ([]GroupMappingDrift, error) {
	return f.drift, nil
}

func (f *fakeRepository) ApplyGroupContextMappings(
	_ context.Context,
	_ string,
	mappings []GroupContextSyncMapping,
	_ bool,
) (GroupContextSyncResult, error) {
	f.applied = append([]GroupContextSyncMapping(nil), mappings...)
	groups := make(map[uuid.UUID]struct{}, len(mappings))
	for _, mapping := range mappings {
		groups[mapping.GroupID] = struct{}{}
	}
	return GroupContextSyncResult{
		AppliedMappings: len(mappings),
		AffectedGroups:  len(groups),
	}, nil
}

func (f *fakeRepository) CreateNode(_ context.Context, input CreateNodeInput) (Node, error) {
	node := Node{
		ID: uuid.New(), Code: input.Code, Name: input.Name, ContextType: input.ContextType,
		ParentID: input.ParentID, Source: input.Source, Metadata: input.Metadata, Enabled: input.Enabled,
	}
	f.nodes[node.ID] = node
	return node, nil
}
func (f *fakeRepository) UpdateNode(_ context.Context, input UpdateNodeInput) (Node, error) {
	node := f.nodes[input.ID]
	node.Code = input.Code
	node.Name = input.Name
	node.ContextType = input.ContextType
	node.ParentID = input.ParentID
	node.Source = input.Source
	node.Metadata = input.Metadata
	node.Enabled = input.Enabled
	node.Version = input.Version + 1
	f.nodes[node.ID] = node
	return node, nil
}
func (f *fakeRepository) DeleteNode(_ context.Context, id uuid.UUID) error {
	delete(f.nodes, id)
	return nil
}
func (f *fakeRepository) GetNode(_ context.Context, id uuid.UUID) (Node, error) {
	node, ok := f.nodes[id]
	if !ok {
		return Node{}, ErrNotFound
	}
	return node, nil
}
func (f *fakeRepository) ListNodes(context.Context, *uuid.UUID, ContextType, bool) ([]Node, error) {
	return nil, nil
}
func (f *fakeRepository) ListAncestors(context.Context, uuid.UUID) ([]Node, error) {
	return nil, nil
}
func (f *fakeRepository) ListDescendants(context.Context, uuid.UUID) ([]Node, error) {
	return nil, nil
}
func (f *fakeRepository) IsDescendant(_ context.Context, ancestor, descendant uuid.UUID) (bool, error) {
	return f.descendants[[2]uuid.UUID{ancestor, descendant}], nil
}
func (f *fakeRepository) UpsertAlias(_ context.Context, input UpsertAliasInput) (Alias, error) {
	return Alias{
		ID: uuid.New(), ContextNodeID: input.ContextNodeID, Namespace: input.Namespace,
		ExternalID: input.ExternalID, Metadata: input.Metadata,
	}, nil
}
func (f *fakeRepository) GetNodeByAlias(_ context.Context, namespace, externalID string) (Node, error) {
	id, ok := f.aliases[[2]string{namespace, externalID}]
	if !ok {
		return Node{}, ErrNotFound
	}
	return f.GetNode(context.Background(), id)
}
func (f *fakeRepository) ListAliases(context.Context, uuid.UUID) ([]Alias, error) {
	return nil, nil
}
func (f *fakeRepository) ListAliasesInScope(
	_ context.Context,
	contextID uuid.UUID,
	includeDescendants bool,
) ([]Alias, error) {
	result := make([]Alias, 0)
	for key, nodeID := range f.aliases {
		if nodeID != contextID &&
			(!includeDescendants || !f.descendants[[2]uuid.UUID{contextID, nodeID}]) {
			continue
		}
		result = append(result, Alias{
			ID:            uuid.New(),
			ContextNodeID: nodeID,
			Namespace:     key[0],
			ExternalID:    key[1],
		})
	}
	return result, nil
}
func (f *fakeRepository) DeleteAlias(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}
func (f *fakeRepository) ListUserAssignments(context.Context, string) ([]Assignment, error) {
	return nil, nil
}
func (f *fakeRepository) ReplaceUserAssignments(context.Context, string, string, []AssignmentInput) error {
	return nil
}
func (f *fakeRepository) ListGroupAssignments(context.Context, uuid.UUID) ([]Assignment, error) {
	return nil, nil
}
func (f *fakeRepository) ReplaceGroupAssignments(context.Context, uuid.UUID, string, []AssignmentInput) error {
	return nil
}
func (f *fakeRepository) ListEffectiveContexts(context.Context, string) ([]EffectiveContext, error) {
	return f.effective, nil
}
func (f *fakeRepository) GetActiveContext(context.Context, string) (*uuid.UUID, error) {
	return f.active, nil
}
func (f *fakeRepository) SetActiveContext(_ context.Context, _ string, id uuid.UUID) error {
	f.active = &id
	return nil
}

func TestExplainAccessAllowsInheritedGroupDescendant(t *testing.T) {
	districtID := uuid.New()
	facilityID := uuid.New()
	groupID := uuid.New()
	groupPath := "/MOH/Data and Statistics/District Team"
	repository := &fakeRepository{
		nodes: map[uuid.UUID]Node{
			districtID: {ID: districtID, Code: "KLA", Name: "Kampala", ContextType: ContextDistrict, Enabled: true},
			facilityID: {ID: facilityID, Code: "FAC-1", Name: "Facility One", ContextType: ContextFacility, Enabled: true},
		},
		effective: []EffectiveContext{{
			Node:      Node{ID: districtID, Code: "KLA", Name: "Kampala", ContextType: ContextDistrict, Enabled: true},
			ScopeMode: ScopeNodeAndDescendants, AssignmentType: "GROUP",
			SourceGroupID: &groupID, SourceGroupPath: &groupPath,
		}},
		descendants: map[[2]uuid.UUID]bool{{districtID, facilityID}: true},
	}

	explanation, err := NewService(repository).ExplainAccess(context.Background(), "user-1", facilityID)
	if err != nil {
		t.Fatal(err)
	}
	if !explanation.Allowed || explanation.AssignmentType != "GROUP" {
		t.Fatalf("expected inherited group access, got %#v", explanation)
	}
}

func TestPolicyRequiresFunctionalPermissionAndContext(t *testing.T) {
	contextID := uuid.New()
	repository := &fakeRepository{
		nodes: map[uuid.UUID]Node{
			contextID: {ID: contextID, Code: "UG", Name: "Uganda", ContextType: ContextNational, Enabled: true},
		},
		effective: []EffectiveContext{{
			Node:      Node{ID: contextID, Code: "UG", Name: "Uganda", ContextType: ContextNational, Enabled: true},
			ScopeMode: ScopeNodeOnly, AssignmentType: "DIRECT",
		}},
		descendants: map[[2]uuid.UUID]bool{},
	}
	policy := NewPolicy(NewService(repository))

	_, err := policy.Authorize(context.Background(), AuthorizationInput{
		UserID: "user-1", ContextID: contextID,
		RequiredPermission: authz.PermissionSurveillanceRead,
	})
	if err == nil {
		t.Fatal("expected missing functional permission to deny access")
	}

	explanation, err := policy.Authorize(context.Background(), AuthorizationInput{
		UserID: "user-1", ContextID: contextID,
		Permissions:        []authz.Permission{authz.PermissionSurveillanceRead},
		RequiredPermission: authz.PermissionSurveillanceRead,
	})
	if err != nil || !explanation.Allowed {
		t.Fatalf("expected combined permission and context access, got %#v, %v", explanation, err)
	}

	explanation, err = policy.Authorize(context.Background(), AuthorizationInput{
		UserID: "user-1", ContextID: contextID,
		Permissions:        []authz.Permission{authz.PermissionSurveillanceRead},
		RequiredPermission: authz.PermissionSurveillanceRead,
		Systems:            []string{authz.SystemDashboardWeb},
		RequiredSystem:     authz.SystemDataStatistics,
	})
	if err != ErrContextForbidden {
		t.Fatalf("expected missing system access to deny, got %#v, %v", explanation, err)
	}
	if explanation.Reason != "required system access is missing" {
		t.Fatalf("unexpected denial reason: %q", explanation.Reason)
	}

	explanation, err = policy.Authorize(context.Background(), AuthorizationInput{
		UserID: "user-1", ContextID: contextID,
		Permissions:        []authz.Permission{authz.PermissionSurveillanceRead},
		RequiredPermission: authz.PermissionSurveillanceRead,
		Systems:            []string{authz.SystemDataStatistics},
		RequiredSystem:     authz.SystemDataStatistics,
	})
	if err != nil || !explanation.Allowed {
		t.Fatalf("expected permission, system, and context access, got %#v, %v", explanation, err)
	}
}

func TestSelectActiveContextRejectsUnassignedContext(t *testing.T) {
	contextID := uuid.New()
	repository := &fakeRepository{
		nodes: map[uuid.UUID]Node{
			contextID: {ID: contextID, Code: "UG", Name: "Uganda", ContextType: ContextNational, Enabled: true},
		},
		descendants: map[[2]uuid.UUID]bool{},
	}
	err := NewService(repository).SelectActiveContext(context.Background(), "user-1", contextID)
	if err != ErrContextForbidden {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestUpsertAliasNormalizesNamespace(t *testing.T) {
	contextID := uuid.New()
	repository := &fakeRepository{
		nodes: map[uuid.UUID]Node{
			contextID: {
				ID: contextID, Code: "FAC-1", Name: "Facility One",
				ContextType: ContextFacility, Enabled: true,
			},
		},
	}

	alias, err := NewService(repository).UpsertAlias(context.Background(), UpsertAliasInput{
		ContextNodeID: contextID,
		Namespace:     " DHIS2 ",
		ExternalID:    " abc123 ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if alias.Namespace != "dhis2" || alias.ExternalID != "abc123" {
		t.Fatalf("alias was not normalized: %#v", alias)
	}
}

func TestResolveAliasNormalizesExternalIdentifier(t *testing.T) {
	contextID := uuid.New()
	repository := &fakeRepository{
		nodes: map[uuid.UUID]Node{
			contextID: {
				ID: contextID, Code: "FAC-1", Name: "Facility One",
				ContextType: ContextFacility, Enabled: true,
			},
		},
		aliases: map[[2]string]uuid.UUID{{"dhis2", "abc123"}: contextID},
	}

	node, err := NewService(repository).ResolveAlias(context.Background(), " DHIS2 ", " abc123 ")
	if err != nil {
		t.Fatal(err)
	}
	if node.ID != contextID {
		t.Fatalf("expected context %s, got %s", contextID, node.ID)
	}
}

func TestPolicyAuthorizesExternalAlias(t *testing.T) {
	contextID := uuid.New()
	repository := &fakeRepository{
		nodes: map[uuid.UUID]Node{
			contextID: {
				ID: contextID, Code: "DIST-1", Name: "District One",
				ContextType: ContextDistrict, Enabled: true,
			},
		},
		aliases: map[[2]string]uuid.UUID{{"surveillance", "district-1"}: contextID},
		effective: []EffectiveContext{{
			Node: Node{
				ID: contextID, Code: "DIST-1", Name: "District One",
				ContextType: ContextDistrict, Enabled: true,
			},
			ScopeMode: ScopeNodeOnly, AssignmentType: "DIRECT",
		}},
		descendants: map[[2]uuid.UUID]bool{},
	}

	explanation, err := NewPolicy(NewService(repository)).AuthorizeAlias(
		context.Background(),
		AliasAuthorizationInput{
			UserID:             "user-1",
			Permissions:        []authz.Permission{authz.PermissionSurveillanceRead},
			RequiredPermission: authz.PermissionSurveillanceRead,
			Namespace:          "surveillance",
			ExternalID:         "district-1",
		},
	)
	if err != nil || !explanation.Allowed {
		t.Fatalf("expected alias authorization, got %#v, %v", explanation, err)
	}
}

func TestUpdateNodeRejectsDescendantAsParent(t *testing.T) {
	parentID := uuid.New()
	childID := uuid.New()
	repository := &fakeRepository{
		nodes: map[uuid.UUID]Node{
			parentID: {
				ID: parentID, Code: "PARENT", Name: "Parent", ContextType: ContextDistrict,
				Enabled: true, Version: 1,
			},
			childID: {
				ID: childID, Code: "CHILD", Name: "Child", ContextType: ContextFacility,
				ParentID: &parentID, Enabled: true, Version: 1,
			},
		},
		descendants: map[[2]uuid.UUID]bool{{parentID, childID}: true},
	}

	_, err := NewService(repository).UpdateNode(context.Background(), UpdateNodeInput{
		ID: parentID, Code: "PARENT", Name: "Parent", ContextType: ContextDistrict,
		ParentID: &childID, Enabled: true, Version: 1,
	})
	if err != ErrCycle {
		t.Fatalf("expected hierarchy cycle rejection, got %v", err)
	}
}

func TestApplyGroupContextSyncRequiresExplicitValidMappings(t *testing.T) {
	contextID := uuid.New()
	groupID := uuid.New()
	repository := &fakeRepository{
		nodes: map[uuid.UUID]Node{
			contextID: {
				ID: contextID, Code: "UG-KLA", Name: "Kampala",
				ContextType: ContextDistrict, Enabled: true,
			},
		},
	}

	result, err := NewService(repository).ApplyGroupContextSync(
		context.Background(),
		"admin-user",
		[]GroupContextSyncMapping{{
			GroupID: groupID, ContextNodeID: contextID,
			ScopeMode: ScopeNodeAndDescendants,
		}},
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.AppliedMappings != 1 || result.AffectedGroups != 1 {
		t.Fatalf("unexpected sync result: %#v", result)
	}
	if len(repository.applied) != 1 || repository.applied[0].ContextNodeID != contextID {
		t.Fatalf("mapping was not applied: %#v", repository.applied)
	}
}

func TestApplyGroupContextSyncRejectsInactiveContext(t *testing.T) {
	contextID := uuid.New()
	repository := &fakeRepository{
		nodes: map[uuid.UUID]Node{
			contextID: {
				ID: contextID, Code: "OLD", Name: "Inactive",
				ContextType: ContextDistrict, Enabled: false,
			},
		},
	}

	_, err := NewService(repository).ApplyGroupContextSync(
		context.Background(),
		"admin-user",
		[]GroupContextSyncMapping{{
			GroupID: uuid.New(), ContextNodeID: contextID, ScopeMode: ScopeNodeOnly,
		}},
		false,
	)
	if err != ErrInactiveContext {
		t.Fatalf("expected inactive context rejection, got %v", err)
	}
}

func TestResolveAudienceUserIDsNormalizesContextsAndPreservesDescendantMode(t *testing.T) {
	contextID := uuid.New()
	repository := &fakeRepository{
		nodes: map[uuid.UUID]Node{
			contextID: {
				ID: contextID, Code: "UG-KLA", Name: "Kampala",
				ContextType: ContextDistrict, Enabled: true,
			},
		},
		effective: []EffectiveContext{{
			Node: Node{
				ID: contextID, Code: "UG-KLA", Name: "Kampala",
				ContextType: ContextDistrict, Enabled: true,
			},
			ScopeMode: ScopeNodeAndDescendants, AssignmentType: "DIRECT",
		}},
		audienceUserIDs: []string{"user-1", "user-2"},
	}

	userIDs, err := NewService(repository).ResolveAudienceUserIDs(
		context.Background(),
		"actor-1",
		[]uuid.UUID{contextID, contextID},
		true,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(userIDs) != 2 {
		t.Fatalf("expected resolved audience, got %#v", userIDs)
	}
	if len(repository.audienceContextIDs) != 1 ||
		repository.audienceContextIDs[0] != contextID {
		t.Fatalf("expected duplicate contexts to be removed, got %#v", repository.audienceContextIDs)
	}
	if !repository.audienceIncludeDescendants {
		t.Fatal("expected descendant audience mode to reach the repository")
	}
}

func TestResolveAudienceUserIDsRejectsContextOutsideActorScope(t *testing.T) {
	contextID := uuid.New()
	repository := &fakeRepository{
		nodes: map[uuid.UUID]Node{
			contextID: {
				ID: contextID, Code: "UG-GUL", Name: "Gulu",
				ContextType: ContextDistrict, Enabled: true,
			},
		},
	}

	_, err := NewService(repository).ResolveAudienceUserIDs(
		context.Background(),
		"actor-1",
		[]uuid.UUID{contextID},
		false,
	)
	if err != ErrContextForbidden {
		t.Fatalf("expected forbidden audience context, got %v", err)
	}
	if len(repository.audienceContextIDs) != 0 {
		t.Fatal("repository must not resolve recipients for an unauthorized context")
	}
}

func TestListAudienceUserIDsRejectsDisabledContext(t *testing.T) {
	contextID := uuid.New()
	repository := &fakeRepository{
		nodes: map[uuid.UUID]Node{
			contextID: {
				ID: contextID, Code: "OLD-FACILITY", Name: "Closed Facility",
				ContextType: ContextFacility, Enabled: false,
			},
		},
	}

	_, err := NewService(repository).ListAudienceUserIDs(
		context.Background(),
		[]uuid.UUID{contextID},
		false,
	)
	if err != ErrInactiveContext {
		t.Fatalf("expected inactive context rejection, got %v", err)
	}
}
