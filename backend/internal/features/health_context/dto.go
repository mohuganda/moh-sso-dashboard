package health_context

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type nodeResponse struct {
	ID          uuid.UUID       `json:"id"`
	Code        string          `json:"code"`
	Name        string          `json:"name"`
	ContextType ContextType     `json:"contextType"`
	ParentID    *uuid.UUID      `json:"parentId,omitempty"`
	Source      string          `json:"source"`
	Metadata    json.RawMessage `json:"metadata"`
	Enabled     bool            `json:"enabled"`
	Version     int             `json:"version"`
}

type aliasResponse struct {
	ID            uuid.UUID       `json:"id"`
	ContextNodeID uuid.UUID       `json:"contextNodeId"`
	Namespace     string          `json:"namespace"`
	ExternalID    string          `json:"externalId"`
	Metadata      json.RawMessage `json:"metadata"`
	CreatedAt     time.Time       `json:"createdAt"`
}

type upsertAliasRequest struct {
	Namespace  string          `json:"namespace" binding:"required"`
	ExternalID string          `json:"externalId" binding:"required"`
	Metadata   json.RawMessage `json:"metadata"`
}

type myContextsResponse struct {
	HealthContexts []effectiveContextResponse `json:"healthContexts"`
}

type selectedContextResponse struct {
	ContextID uuid.UUID `json:"contextId"`
}

type updateResponse struct {
	Updated bool `json:"updated"`
}

type effectiveContextResponse struct {
	nodeResponse
	ScopeMode       ScopeMode  `json:"scopeMode"`
	AssignmentType  string     `json:"assignmentType"`
	AssignmentID    uuid.UUID  `json:"assignmentId"`
	SourceGroupID   *uuid.UUID `json:"sourceGroupId,omitempty"`
	SourceGroupPath *string    `json:"sourceGroupPath,omitempty"`
	IsDefault       bool       `json:"isDefault"`
	IsActive        bool       `json:"isActive"`
}

type assignmentResponse struct {
	ID              uuid.UUID  `json:"id"`
	UserID          string     `json:"userId,omitempty"`
	GroupID         *uuid.UUID `json:"groupId,omitempty"`
	ContextNodeID   uuid.UUID  `json:"contextNodeId"`
	ScopeMode       ScopeMode  `json:"scopeMode"`
	IsDefault       bool       `json:"isDefault"`
	ValidFrom       *time.Time `json:"validFrom,omitempty"`
	ValidUntil      *time.Time `json:"validUntil,omitempty"`
	Source          string     `json:"source"`
	SourceReference *string    `json:"sourceReference,omitempty"`
}

type createNodeRequest struct {
	Code        string          `json:"code" binding:"required"`
	Name        string          `json:"name" binding:"required"`
	ContextType ContextType     `json:"contextType" binding:"required"`
	ParentID    *uuid.UUID      `json:"parentId"`
	Source      string          `json:"source"`
	Metadata    json.RawMessage `json:"metadata"`
	Enabled     *bool           `json:"enabled"`
}

type updateNodeRequest struct {
	Code        string          `json:"code" binding:"required"`
	Name        string          `json:"name" binding:"required"`
	ContextType ContextType     `json:"contextType" binding:"required"`
	ParentID    *uuid.UUID      `json:"parentId"`
	Source      string          `json:"source"`
	Metadata    json.RawMessage `json:"metadata"`
	Enabled     bool            `json:"enabled"`
	Version     int             `json:"version" binding:"required,min=1"`
}

type assignmentRequest struct {
	ContextNodeID   uuid.UUID  `json:"contextNodeId" binding:"required"`
	ScopeMode       ScopeMode  `json:"scopeMode" binding:"required"`
	IsDefault       bool       `json:"isDefault"`
	ValidFrom       *time.Time `json:"validFrom"`
	ValidUntil      *time.Time `json:"validUntil"`
	Source          string     `json:"source"`
	SourceReference *string    `json:"sourceReference"`
}

type replaceAssignmentsRequest struct {
	Assignments []assignmentRequest `json:"assignments" binding:"required"`
}

type selectContextRequest struct {
	ContextID uuid.UUID `json:"contextId" binding:"required"`
}

type accessExplanationResponse struct {
	Allowed           bool          `json:"allowed"`
	RequestedContext  nodeResponse  `json:"requestedContext"`
	AssignmentContext *nodeResponse `json:"assignmentContext,omitempty"`
	ScopeMode         ScopeMode     `json:"scopeMode,omitempty"`
	AssignmentType    string        `json:"assignmentType,omitempty"`
	SourceGroupID     *uuid.UUID    `json:"sourceGroupId,omitempty"`
	SourceGroupPath   *string       `json:"sourceGroupPath,omitempty"`
	Reason            string        `json:"reason"`
}

type groupMappingDriftResponse struct {
	GroupID           uuid.UUID  `json:"groupId"`
	GroupPath         string     `json:"groupPath"`
	KeycloakGroupID   string     `json:"keycloakGroupId,omitempty"`
	GroupEnabled      bool       `json:"groupEnabled"`
	ContextNodeID     *uuid.UUID `json:"contextNodeId,omitempty"`
	ContextCode       string     `json:"contextCode,omitempty"`
	ContextName       string     `json:"contextName,omitempty"`
	ContextEnabled    bool       `json:"contextEnabled"`
	ScopeMode         ScopeMode  `json:"scopeMode,omitempty"`
	Status            string     `json:"status"`
	RecommendedAction string     `json:"recommendedAction"`
}

type healthContextDriftReportResponse struct {
	GeneratedAt time.Time                   `json:"generatedAt"`
	Total       int                         `json:"total"`
	InSync      int                         `json:"inSync"`
	Drifted     int                         `json:"drifted"`
	Items       []groupMappingDriftResponse `json:"items"`
}

type groupContextSyncMappingRequest struct {
	GroupID       uuid.UUID `json:"groupId" binding:"required"`
	ContextNodeID uuid.UUID `json:"contextNodeId" binding:"required"`
	ScopeMode     ScopeMode `json:"scopeMode" binding:"required"`
}

type healthContextSyncRequest struct {
	Mappings        []groupContextSyncMappingRequest `json:"mappings"`
	ReplaceExisting bool                             `json:"replaceExisting"`
}

type healthContextSyncPreviewResponse struct {
	Valid           bool                             `json:"valid"`
	MappingCount    int                              `json:"mappingCount"`
	AffectedGroups  int                              `json:"affectedGroups"`
	ReplaceExisting bool                             `json:"replaceExisting"`
	Drift           healthContextDriftReportResponse `json:"drift"`
}

type healthContextSyncApplyResponse struct {
	AppliedMappings int `json:"appliedMappings"`
	AffectedGroups  int `json:"affectedGroups"`
}

func toNodeResponse(node Node) nodeResponse {
	return nodeResponse{
		ID: node.ID, Code: node.Code, Name: node.Name, ContextType: node.ContextType,
		ParentID: node.ParentID, Source: node.Source, Metadata: node.Metadata,
		Enabled: node.Enabled, Version: node.Version,
	}
}

func toNodeResponses(nodes []Node) []nodeResponse {
	result := make([]nodeResponse, 0, len(nodes))
	for _, node := range nodes {
		result = append(result, toNodeResponse(node))
	}
	return result
}

func toAliasResponse(alias Alias) aliasResponse {
	return aliasResponse{
		ID: alias.ID, ContextNodeID: alias.ContextNodeID,
		Namespace: alias.Namespace, ExternalID: alias.ExternalID,
		Metadata: alias.Metadata, CreatedAt: alias.CreatedAt,
	}
}

func toAliasResponses(aliases []Alias) []aliasResponse {
	result := make([]aliasResponse, 0, len(aliases))
	for _, alias := range aliases {
		result = append(result, toAliasResponse(alias))
	}
	return result
}

func toEffectiveResponses(contexts []EffectiveContext) []effectiveContextResponse {
	result := make([]effectiveContextResponse, 0, len(contexts))
	for _, item := range contexts {
		result = append(result, effectiveContextResponse{
			nodeResponse:    toNodeResponse(item.Node),
			ScopeMode:       item.ScopeMode,
			AssignmentType:  item.AssignmentType,
			AssignmentID:    item.AssignmentID,
			SourceGroupID:   item.SourceGroupID,
			SourceGroupPath: item.SourceGroupPath,
			IsDefault:       item.IsDefault,
			IsActive:        item.IsActive,
		})
	}
	return result
}

func toAssignmentResponses(assignments []Assignment) []assignmentResponse {
	result := make([]assignmentResponse, 0, len(assignments))
	for _, item := range assignments {
		result = append(result, assignmentResponse{
			ID: item.ID, UserID: item.UserID, GroupID: item.GroupID,
			ContextNodeID: item.ContextNodeID, ScopeMode: item.ScopeMode,
			IsDefault: item.IsDefault, ValidFrom: item.ValidFrom,
			ValidUntil: item.ValidUntil, Source: item.Source,
			SourceReference: item.SourceReference,
		})
	}
	return result
}

func toExplanationResponse(item AccessExplanation) accessExplanationResponse {
	response := accessExplanationResponse{
		Allowed: item.Allowed, RequestedContext: toNodeResponse(item.RequestedContext),
		ScopeMode: item.ScopeMode, AssignmentType: item.AssignmentType,
		SourceGroupID: item.SourceGroupID, SourceGroupPath: item.SourceGroupPath,
		Reason: item.Reason,
	}
	if item.AssignmentContext != nil {
		context := toNodeResponse(*item.AssignmentContext)
		response.AssignmentContext = &context
	}
	return response
}

func toDriftReportResponse(items []GroupMappingDrift) healthContextDriftReportResponse {
	response := healthContextDriftReportResponse{
		GeneratedAt: time.Now().UTC(),
		Total:       len(items),
		Items:       make([]groupMappingDriftResponse, 0, len(items)),
	}
	for _, item := range items {
		if item.Status == "IN_SYNC" {
			response.InSync++
		} else {
			response.Drifted++
		}
		response.Items = append(response.Items, groupMappingDriftResponse{
			GroupID: item.GroupID, GroupPath: item.GroupPath,
			KeycloakGroupID: item.KeycloakGroupID, GroupEnabled: item.GroupEnabled,
			ContextNodeID: item.ContextNodeID, ContextCode: item.ContextCode,
			ContextName: item.ContextName, ContextEnabled: item.ContextEnabled,
			ScopeMode: item.ScopeMode, Status: item.Status,
			RecommendedAction: item.RecommendedAction,
		})
	}
	return response
}
