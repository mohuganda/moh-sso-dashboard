package health_context

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateNode(context.Context, CreateNodeInput) (Node, error)
	UpdateNode(context.Context, UpdateNodeInput) (Node, error)
	DeleteNode(context.Context, uuid.UUID) error
	GetNode(context.Context, uuid.UUID) (Node, error)
	ListNodes(context.Context, *uuid.UUID, ContextType, bool) ([]Node, error)
	ListAncestors(context.Context, uuid.UUID) ([]Node, error)
	ListDescendants(context.Context, uuid.UUID) ([]Node, error)
	IsDescendant(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	UpsertAlias(context.Context, UpsertAliasInput) (Alias, error)
	GetNodeByAlias(context.Context, string, string) (Node, error)
	ListAliases(context.Context, uuid.UUID) ([]Alias, error)
	ListAliasesInScope(context.Context, uuid.UUID, bool) ([]Alias, error)
	DeleteAlias(context.Context, uuid.UUID, uuid.UUID) error

	ListUserAssignments(context.Context, string) ([]Assignment, error)
	ReplaceUserAssignments(context.Context, string, string, []AssignmentInput) error
	ListGroupAssignments(context.Context, uuid.UUID) ([]Assignment, error)
	ReplaceGroupAssignments(context.Context, uuid.UUID, string, []AssignmentInput) error
	ListEffectiveContexts(context.Context, string) ([]EffectiveContext, error)
	ListAudienceUserIDs(context.Context, []uuid.UUID, bool) ([]string, error)
	ListGroupMappingDrift(context.Context) ([]GroupMappingDrift, error)
	ApplyGroupContextMappings(context.Context, string, []GroupContextSyncMapping, bool) (GroupContextSyncResult, error)

	GetActiveContext(context.Context, string) (*uuid.UUID, error)
	SetActiveContext(context.Context, string, uuid.UUID) error
}
