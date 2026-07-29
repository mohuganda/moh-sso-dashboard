package health_context

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/observability"
	"github.com/moh-sso-dashboard/internal/service"
	"github.com/moh-sso-dashboard/internal/utils"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	service *Service
	audit   *service.AuditService
}

func NewHandler(service *Service, audit *service.AuditService) *Handler {
	return &Handler{service: service, audit: audit}
}

// RequireAccessibleSelection validates the request health-context header. Users
// with contextual assignments must select a context; unassigned legacy users
// remain compatible while resource migrations are completed.
func (h *Handler) RequireAccessibleSelection() gin.HandlerFunc {
	return func(c *gin.Context) {
		value := strings.TrimSpace(c.GetHeader("X-Health-Context-ID"))
		if value == "" {
			contexts, err := h.service.ListEffectiveContexts(
				c.Request.Context(),
				c.GetString("user_id"),
			)
			if err != nil {
				response.Fail(
					c,
					http.StatusServiceUnavailable,
					"HEALTH_CONTEXT_UNAVAILABLE",
					"health context access could not be verified",
				)
				c.Abort()
				return
			}
			if len(contexts) > 0 {
				log.Warn().
					Str("request_id", observability.RequestIDFromContext(c.Request.Context())).
					Str("feature", featureFromRequest(c)).
					Str("reason", "active_context_required").
					Msg("health context authorization denied")
				response.Fail(
					c,
					http.StatusBadRequest,
					"HEALTH_CONTEXT_REQUIRED",
					"select a health context for this request",
				)
				c.Abort()
				return
			}
			c.Next()
			return
		}
		contextID, err := uuid.Parse(value)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "INVALID_HEALTH_CONTEXT", "invalid health context")
			c.Abort()
			return
		}
		explanation, err := h.service.ExplainAccess(
			c.Request.Context(),
			c.GetString("user_id"),
			contextID,
		)
		if err != nil || !explanation.Allowed {
			log.Warn().
				Str("request_id", observability.RequestIDFromContext(c.Request.Context())).
				Str("feature", featureFromRequest(c)).
				Str("context_type", string(explanation.RequestedContext.ContextType)).
				Str("reason", explanation.Reason).
				Msg("health context authorization denied")
			response.Fail(c, http.StatusForbidden, "HEALTH_CONTEXT_FORBIDDEN", "health context access denied")
			c.Abort()
			return
		}
		c.Set("health_context_id", contextID)
		c.Set("health_context", explanation.RequestedContext)
		c.Set("health_context_scope_mode", explanation.ScopeMode)
		c.Next()
	}
}

func featureFromRequest(c *gin.Context) string {
	path := strings.TrimPrefix(c.Request.URL.Path, "/api/v1/")
	if path == "" {
		return "unknown"
	}
	if separator := strings.IndexByte(path, '/'); separator >= 0 {
		return path[:separator]
	}
	return path
}

func (h *Handler) CreateNode(c *gin.Context) {
	var request createNodeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", "invalid health context")
		return
	}
	enabled := true
	if request.Enabled != nil {
		enabled = *request.Enabled
	}
	node, err := h.service.CreateNode(c.Request.Context(), CreateNodeInput{
		Code: request.Code, Name: request.Name, ContextType: request.ContextType,
		ParentID: request.ParentID, Source: request.Source, Metadata: request.Metadata,
		Enabled: enabled,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	h.auditAction(c, "health_context.created", map[string]any{
		"context_id": node.ID, "code": node.Code, "context_type": node.ContextType,
	})
	response.OK(c, http.StatusCreated, toNodeResponse(node))
}

func (h *Handler) CreateChild(c *gin.Context) {
	parentID, ok := parseID(c, "contextId")
	if !ok {
		return
	}
	var request createNodeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", "invalid health context")
		return
	}
	request.ParentID = &parentID
	enabled := true
	if request.Enabled != nil {
		enabled = *request.Enabled
	}
	node, err := h.service.CreateNode(c.Request.Context(), CreateNodeInput{
		Code: request.Code, Name: request.Name, ContextType: request.ContextType,
		ParentID: request.ParentID, Source: request.Source, Metadata: request.Metadata,
		Enabled: enabled,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	h.auditAction(c, "health_context.created", map[string]any{
		"context_id": node.ID, "parent_id": parentID, "code": node.Code,
		"context_type": node.ContextType,
	})
	response.OK(c, http.StatusCreated, toNodeResponse(node))
}

func (h *Handler) UpdateNode(c *gin.Context) {
	id, ok := parseID(c, "contextId")
	if !ok {
		return
	}
	var request updateNodeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", "invalid health context")
		return
	}
	previous, err := h.service.GetNode(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	node, err := h.service.UpdateNode(c.Request.Context(), UpdateNodeInput{
		ID: id, Code: request.Code, Name: request.Name, ContextType: request.ContextType,
		ParentID: request.ParentID, Source: request.Source, Metadata: request.Metadata,
		Enabled: request.Enabled, Version: request.Version,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	h.auditAction(c, "health_context.updated", map[string]any{
		"context_id": id,
		"previous":   toNodeResponse(previous),
		"current":    toNodeResponse(node),
	})
	response.OK(c, http.StatusOK, toNodeResponse(node))
}

func (h *Handler) DeleteNode(c *gin.Context) {
	id, ok := parseID(c, "contextId")
	if !ok {
		return
	}
	node, err := h.service.GetNode(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	if err := h.service.DeleteNode(c.Request.Context(), id); err != nil {
		h.handleError(c, err)
		return
	}
	h.auditAction(c, "health_context.deleted", map[string]any{
		"context_id": id, "code": node.Code, "context_type": node.ContextType,
	})
	response.OK(c, http.StatusOK, updateResponse{Updated: true})
}

func (h *Handler) ListNodes(c *gin.Context) {
	var parentID *uuid.UUID
	value := strings.TrimSpace(c.Query("parentId"))
	if value == "" {
		value = strings.TrimSpace(c.Param("contextId"))
	}
	if value != "" {
		parsed, err := uuid.Parse(value)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", "invalid parentId")
			return
		}
		parentID = &parsed
	}
	nodes, err := h.service.ListNodes(
		c.Request.Context(), parentID, ContextType(c.Query("type")),
		c.Query("includeDisabled") == "true",
	)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, toNodeResponses(nodes))
}

func (h *Handler) GetNode(c *gin.Context) {
	id, ok := parseID(c, "contextId")
	if !ok {
		return
	}
	node, err := h.service.GetNode(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, toNodeResponse(node))
}

func (h *Handler) ListDescendants(c *gin.Context) {
	id, ok := parseID(c, "contextId")
	if !ok {
		return
	}
	nodes, err := h.service.ListDescendants(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, toNodeResponses(nodes))
}

func (h *Handler) UpsertAlias(c *gin.Context) {
	contextID, ok := parseID(c, "contextId")
	if !ok {
		return
	}
	var request upsertAliasRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", "invalid health context alias")
		return
	}
	alias, err := h.service.UpsertAlias(c.Request.Context(), UpsertAliasInput{
		ContextNodeID: contextID,
		Namespace:     request.Namespace,
		ExternalID:    request.ExternalID,
		Metadata:      request.Metadata,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	h.auditAction(c, "health_context.alias_upserted", map[string]any{
		"context_id": contextID, "alias_id": alias.ID, "namespace": alias.Namespace,
	})
	response.OK(c, http.StatusOK, toAliasResponse(alias))
}

func (h *Handler) ListAliases(c *gin.Context) {
	contextID, ok := parseID(c, "contextId")
	if !ok {
		return
	}
	aliases, err := h.service.ListAliases(c.Request.Context(), contextID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, toAliasResponses(aliases))
}

func (h *Handler) DeleteAlias(c *gin.Context) {
	contextID, ok := parseID(c, "contextId")
	if !ok {
		return
	}
	aliasID, ok := parseID(c, "aliasId")
	if !ok {
		return
	}
	if err := h.service.DeleteAlias(c.Request.Context(), contextID, aliasID); err != nil {
		h.handleError(c, err)
		return
	}
	h.auditAction(c, "health_context.alias_deleted", map[string]any{
		"context_id": contextID, "alias_id": aliasID,
	})
	response.OK(c, http.StatusOK, updateResponse{Updated: true})
}

func (h *Handler) MyContexts(c *gin.Context) {
	items, err := h.service.ListEffectiveContexts(c.Request.Context(), c.GetString("user_id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, myContextsResponse{HealthContexts: toEffectiveResponses(items)})
}

func (h *Handler) GetMyContext(c *gin.Context) {
	id, ok := parseID(c, "contextId")
	if !ok {
		return
	}
	item, err := h.service.ExplainAccess(c.Request.Context(), c.GetString("user_id"), id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	if !item.Allowed {
		h.handleError(c, ErrContextForbidden)
		return
	}
	response.OK(c, http.StatusOK, toExplanationResponse(item))
}

func (h *Handler) ExplainMyAccess(c *gin.Context) {
	id, ok := parseID(c, "contextId")
	if !ok {
		return
	}
	item, err := h.service.ExplainAccess(c.Request.Context(), c.GetString("user_id"), id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, toExplanationResponse(item))
}

func (h *Handler) SelectMyContext(c *gin.Context) {
	var request selectContextRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", "invalid context selection")
		return
	}
	if err := h.service.SelectActiveContext(c.Request.Context(), c.GetString("user_id"), request.ContextID); err != nil {
		h.handleError(c, err)
		return
	}
	h.auditAction(c, "health_context.selected", map[string]any{"context_id": request.ContextID})
	response.OK(c, http.StatusOK, selectedContextResponse{ContextID: request.ContextID})
}

func (h *Handler) ListUserAssignments(c *gin.Context) {
	items, err := h.service.ListUserAssignments(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, toAssignmentResponses(items))
}

func (h *Handler) ReplaceUserAssignments(c *gin.Context) {
	var request replaceAssignmentsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", "invalid assignments")
		return
	}
	inputs := assignmentInputs(request.Assignments)
	if err := h.service.ReplaceUserAssignments(
		c.Request.Context(), c.Param("id"), c.GetString("user_id"), inputs,
	); err != nil {
		h.handleError(c, err)
		return
	}
	h.auditAction(c, "health_context.user_assignments_replaced", map[string]any{
		"target_user_id": c.Param("id"), "assignment_count": len(inputs),
	})
	response.OK(c, http.StatusOK, updateResponse{Updated: true})
}

func (h *Handler) ListGroupAssignments(c *gin.Context) {
	id, ok := parseID(c, "groupId")
	if !ok {
		return
	}
	items, err := h.service.ListGroupAssignments(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, toAssignmentResponses(items))
}

func (h *Handler) ReplaceGroupAssignments(c *gin.Context) {
	id, ok := parseID(c, "groupId")
	if !ok {
		return
	}
	var request replaceAssignmentsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", "invalid assignments")
		return
	}
	inputs := assignmentInputs(request.Assignments)
	if err := h.service.ReplaceGroupAssignments(
		c.Request.Context(), id, c.GetString("user_id"), inputs,
	); err != nil {
		h.handleError(c, err)
		return
	}
	h.auditAction(c, "health_context.group_assignments_replaced", map[string]any{
		"group_id": id, "assignment_count": len(inputs),
	})
	response.OK(c, http.StatusOK, updateResponse{Updated: true})
}

func (h *Handler) GetDrift(c *gin.Context) {
	items, err := h.service.GetGroupMappingDrift(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, toDriftReportResponse(items))
}

func (h *Handler) PreviewSync(c *gin.Context) {
	var request healthContextSyncRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", "invalid sync preview")
		return
	}
	mappings := syncMappings(request.Mappings)
	items, err := h.service.PreviewGroupContextSync(c.Request.Context(), mappings)
	if err != nil {
		h.handleError(c, err)
		return
	}
	groups := make(map[uuid.UUID]struct{}, len(mappings))
	for _, mapping := range mappings {
		groups[mapping.GroupID] = struct{}{}
	}
	h.auditAction(c, "health_context.sync_previewed", map[string]any{
		"mapping_count": len(mappings), "affected_groups": len(groups),
		"replace_existing": request.ReplaceExisting,
	})
	response.OK(c, http.StatusOK, healthContextSyncPreviewResponse{
		Valid: true, MappingCount: len(mappings), AffectedGroups: len(groups),
		ReplaceExisting: request.ReplaceExisting, Drift: toDriftReportResponse(items),
	})
}

func (h *Handler) ApplySync(c *gin.Context) {
	var request healthContextSyncRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", "invalid sync request")
		return
	}
	result, err := h.service.ApplyGroupContextSync(
		c.Request.Context(),
		c.GetString("user_id"),
		syncMappings(request.Mappings),
		request.ReplaceExisting,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}
	h.auditAction(c, "health_context.sync_applied", map[string]any{
		"mapping_count": result.AppliedMappings, "affected_groups": result.AffectedGroups,
		"replace_existing": request.ReplaceExisting,
	})
	response.OK(c, http.StatusOK, healthContextSyncApplyResponse{
		AppliedMappings: result.AppliedMappings,
		AffectedGroups:  result.AffectedGroups,
	})
}

func syncMappings(requests []groupContextSyncMappingRequest) []GroupContextSyncMapping {
	result := make([]GroupContextSyncMapping, 0, len(requests))
	for _, request := range requests {
		result = append(result, GroupContextSyncMapping{
			GroupID: request.GroupID, ContextNodeID: request.ContextNodeID,
			ScopeMode: request.ScopeMode,
		})
	}
	return result
}

func assignmentInputs(requests []assignmentRequest) []AssignmentInput {
	result := make([]AssignmentInput, 0, len(requests))
	for _, request := range requests {
		result = append(result, AssignmentInput{
			ContextNodeID: request.ContextNodeID, ScopeMode: request.ScopeMode,
			IsDefault: request.IsDefault, ValidFrom: request.ValidFrom,
			ValidUntil: request.ValidUntil, Source: request.Source,
			SourceReference: request.SourceReference,
		})
	}
	return result
}

func parseID(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", "invalid identifier")
		return uuid.Nil, false
	}
	return id, true
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "health context not found")
	case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrInvalidAssignment):
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", "invalid health context input")
	case errors.Is(err, ErrCycle):
		response.Fail(c, http.StatusConflict, "HEALTH_CONTEXT_CYCLE", "health context hierarchy cycle")
	case errors.Is(err, ErrContextInUse):
		response.Fail(c, http.StatusConflict, "HEALTH_CONTEXT_IN_USE", "health context has children or assignments")
	case errors.Is(err, ErrVersionConflict):
		response.Fail(c, http.StatusConflict, "VERSION_CONFLICT", "health context was updated by another request")
	case IsAccessDenied(err):
		response.Fail(c, http.StatusForbidden, "FORBIDDEN", "health context access denied")
	default:
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "health context request failed")
	}
}

func (h *Handler) auditAction(c *gin.Context, action string, details map[string]any) {
	if h.audit == nil {
		return
	}
	_ = h.audit.Log(c.Request.Context(), utils.ToNullUUID(c.GetString("user_id")), action, details)
}
