package rbac

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/http/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListSystems(c *gin.Context) {
	systems, err := h.service.ListSystems(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_SYSTEMS_FAILED", err.Error())
		return
	}
	response.OK(c, http.StatusOK, systems)
}

func (h *Handler) GetSystem(c *gin.Context) {
	system, err := h.service.GetSystem(c.Request.Context(), c.Param("clientId"))
	if err != nil {
		writeError(c, err, "GET_SYSTEM_FAILED")
		return
	}
	response.OK(c, http.StatusOK, system)
}

func (h *Handler) UpdateSystem(c *gin.Context) {
	var input UpsertSystemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	input.ClientID = c.Param("clientId")

	system, err := h.service.UpsertSystem(c.Request.Context(), input)
	if err != nil {
		writeError(c, err, "UPDATE_SYSTEM_FAILED")
		return
	}
	response.OK(c, http.StatusOK, system)
}

func (h *Handler) ListPermissions(c *gin.Context) {
	permissions, err := h.service.ListPermissions(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_PERMISSIONS_FAILED", err.Error())
		return
	}
	response.OK(c, http.StatusOK, permissions)
}

func (h *Handler) ListSystemRoles(c *gin.Context) {
	roles, err := h.service.ListSystemRoles(c.Request.Context(), c.Param("clientId"))
	if err != nil {
		writeError(c, err, "LIST_SYSTEM_ROLES_FAILED")
		return
	}
	response.OK(c, http.StatusOK, roles)
}

func (h *Handler) CreateSystemRole(c *gin.Context) {
	var input RoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	role, err := h.service.CreateSystemRole(c.Request.Context(), c.Param("clientId"), input)
	if err != nil {
		writeError(c, err, "CREATE_SYSTEM_ROLE_FAILED")
		return
	}
	response.OK(c, http.StatusCreated, role)
}

func (h *Handler) UpdateSystemRole(c *gin.Context) {
	var input RoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	role, err := h.service.UpdateSystemRole(c.Request.Context(), c.Param("roleId"), input)
	if err != nil {
		writeError(c, err, "UPDATE_SYSTEM_ROLE_FAILED")
		return
	}
	response.OK(c, http.StatusOK, role)
}

func (h *Handler) DeleteSystemRole(c *gin.Context) {
	if err := h.service.DeleteSystemRole(c.Request.Context(), c.Param("roleId")); err != nil {
		writeError(c, err, "DELETE_SYSTEM_ROLE_FAILED")
		return
	}
	response.OK(c, http.StatusOK, gin.H{"deleted": true})
}

func (h *Handler) AssignSystemRolePermission(c *gin.Context) {
	var input PermissionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	if err := h.service.AssignSystemRolePermission(c.Request.Context(), c.Param("roleId"), input.PermissionKey); err != nil {
		writeError(c, err, "ASSIGN_PERMISSION_FAILED")
		return
	}
	response.OK(c, http.StatusOK, gin.H{"assigned": true})
}

func (h *Handler) RemoveSystemRolePermission(c *gin.Context) {
	if err := h.service.RemoveSystemRolePermission(c.Request.Context(), c.Param("roleId"), c.Param("permissionKey")); err != nil {
		writeError(c, err, "REMOVE_PERMISSION_FAILED")
		return
	}
	response.OK(c, http.StatusOK, gin.H{"removed": true})
}

func (h *Handler) ListRealmRolePermissions(c *gin.Context) {
	groups, err := h.service.ListRealmRolePermissions(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_REALM_ROLES_FAILED", err.Error())
		return
	}
	response.OK(c, http.StatusOK, groups)
}

func (h *Handler) AssignRealmRolePermission(c *gin.Context) {
	var input PermissionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	if err := h.service.AssignRealmRolePermission(c.Request.Context(), c.Param("realmRole"), input.PermissionKey); err != nil {
		writeError(c, err, "ASSIGN_REALM_PERMISSION_FAILED")
		return
	}
	response.OK(c, http.StatusOK, gin.H{"assigned": true})
}

func (h *Handler) RemoveRealmRolePermission(c *gin.Context) {
	if err := h.service.RemoveRealmRolePermission(c.Request.Context(), c.Param("realmRole"), c.Param("permissionKey")); err != nil {
		writeError(c, err, "REMOVE_REALM_PERMISSION_FAILED")
		return
	}
	response.OK(c, http.StatusOK, gin.H{"removed": true})
}

func (h *Handler) AddSystemAccessRole(c *gin.Context) {
	var input AccessRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	if err := h.service.AddSystemAccessRole(c.Request.Context(), c.Param("clientId"), input.RoleName); err != nil {
		writeError(c, err, "ADD_ACCESS_ROLE_FAILED")
		return
	}
	response.OK(c, http.StatusOK, gin.H{"added": true})
}

func (h *Handler) RemoveSystemAccessRole(c *gin.Context) {
	force := c.Query("force") == "true"
	if err := h.service.RemoveSystemAccessRole(c.Request.Context(), c.Param("clientId"), c.Param("roleName"), force); err != nil {
		writeError(c, err, "REMOVE_ACCESS_ROLE_FAILED")
		return
	}
	response.OK(c, http.StatusOK, gin.H{"removed": true})
}

func (h *Handler) GetDrift(c *gin.Context) {
	report, err := h.service.DriftFromDefaultRealmExport(c.Request.Context())
	if err != nil {
		writeError(c, err, "DRIFT_FAILED")
		return
	}
	response.OK(c, http.StatusOK, report)
}

func (h *Handler) DriftFromRealmExport(c *gin.Context) {
	payload, ok := bindRealmExportPayload(c)
	if !ok {
		return
	}
	report, err := h.service.DriftFromRealmExport(c.Request.Context(), payload)
	if err != nil {
		writeError(c, err, "DRIFT_REALM_EXPORT_FAILED")
		return
	}
	response.OK(c, http.StatusOK, report)
}

func (h *Handler) PreviewSync(c *gin.Context) {
	payload, ok := bindRealmExportPayload(c)
	if !ok {
		return
	}
	preview, _, err := h.service.PreviewRealmExportSync(c.Request.Context(), payload)
	if err != nil {
		writeError(c, err, "SYNC_PREVIEW_FAILED")
		return
	}
	response.OK(c, http.StatusOK, preview)
}

func (h *Handler) ApplySync(c *gin.Context) {
	payload, ok := bindRealmExportPayload(c)
	if !ok {
		return
	}
	result, err := h.service.ApplyRealmExportSync(c.Request.Context(), payload)
	if err != nil {
		writeError(c, err, "SYNC_APPLY_FAILED")
		return
	}
	response.OK(c, http.StatusOK, result)
}

func (h *Handler) GetEffectiveAccess(c *gin.Context) {
	result, err := h.service.GetEffectiveAccess(
		c.Request.Context(),
		c.Param("userId"),
		c.Query("username"),
		c.Query("email"),
	)
	if err != nil {
		writeError(c, err, "EFFECTIVE_ACCESS_FAILED")
		return
	}
	response.OK(c, http.StatusOK, result)
}

func (h *Handler) UpdatePermissionMetadata(c *gin.Context) {
	var input PermissionMetadataInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	permission, err := h.service.UpdatePermissionMetadata(c.Request.Context(), c.Param("permissionKey"), input)
	if err != nil {
		writeError(c, err, "UPDATE_PERMISSION_FAILED")
		return
	}
	response.OK(c, http.StatusOK, permission)
}

func (h *Handler) GetRoleUsage(c *gin.Context) {
	usage, err := h.service.GetRoleUsage(c.Request.Context(), c.Param("roleId"))
	if err != nil {
		writeError(c, err, "ROLE_USAGE_FAILED")
		return
	}
	response.OK(c, http.StatusOK, usage)
}

func (h *Handler) GetRealmRoleUsage(c *gin.Context) {
	usage, err := h.service.GetRealmRoleUsage(c.Request.Context(), c.Param("realmRole"))
	if err != nil {
		writeError(c, err, "REALM_ROLE_USAGE_FAILED")
		return
	}
	response.OK(c, http.StatusOK, usage)
}

func (h *Handler) PreviewChange(c *gin.Context) {
	var input ChangePreviewRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	preview, err := h.service.PreviewChange(c.Request.Context(), input)
	if err != nil {
		writeError(c, err, "CHANGE_PREVIEW_FAILED")
		return
	}
	response.OK(c, http.StatusOK, preview)
}

func (h *Handler) ExportSeed(c *gin.Context) {
	seed, err := h.service.ExportSeed(c.Request.Context())
	if err != nil {
		writeError(c, err, "EXPORT_RBAC_FAILED")
		return
	}
	response.OK(c, http.StatusOK, seed)
}

func (h *Handler) PreviewImport(c *gin.Context) {
	var input ImportPreviewRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	preview, _, err := h.service.PreviewImport(c.Request.Context(), input)
	if err != nil {
		writeError(c, err, "IMPORT_PREVIEW_FAILED")
		return
	}
	response.OK(c, http.StatusOK, preview)
}

func (h *Handler) ApplyImport(c *gin.Context) {
	var input ImportPreviewRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	result, err := h.service.ApplyImport(c.Request.Context(), input)
	if err != nil {
		writeError(c, err, "IMPORT_APPLY_FAILED")
		return
	}
	response.OK(c, http.StatusOK, result)
}

func (h *Handler) ListAuditEvents(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	events, err := h.service.ListAuditEvents(c.Request.Context(), AuditFilter{
		ActorUserID:    c.Query("actor"),
		SystemClientID: c.Query("systemClientId"),
		RoleName:       c.Query("roleName"),
		PermissionKey:  c.Query("permissionKey"),
		Action:         c.Query("action"),
		From:           c.Query("from"),
		To:             c.Query("to"),
		Limit:          limit,
	})
	if err != nil {
		writeError(c, err, "LIST_RBAC_AUDIT_FAILED")
		return
	}
	response.OK(c, http.StatusOK, events)
}

func (h *Handler) ListRoleTemplates(c *gin.Context) {
	response.OK(c, http.StatusOK, h.service.RoleTemplates())
}

func (h *Handler) CreateRoleFromTemplate(c *gin.Context) {
	var input RoleFromTemplateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	role, err := h.service.CreateRoleFromTemplate(c.Request.Context(), c.Param("clientId"), input)
	if err != nil {
		writeError(c, err, "ROLE_TEMPLATE_FAILED")
		return
	}
	response.OK(c, http.StatusCreated, role)
}

func (h *Handler) CopyPermissions(c *gin.Context) {
	var input CopyPermissionsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	role, err := h.service.CopyPermissions(c.Request.Context(), c.Param("roleId"), input)
	if err != nil {
		writeError(c, err, "COPY_PERMISSIONS_FAILED")
		return
	}
	response.OK(c, http.StatusOK, role)
}

func (h *Handler) BulkAssignPermission(c *gin.Context) {
	var input BulkPermissionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	if err := h.service.BulkAssignPermission(c.Request.Context(), input); err != nil {
		writeError(c, err, "BULK_ASSIGN_PERMISSION_FAILED")
		return
	}
	response.OK(c, http.StatusOK, gin.H{"assigned": true})
}

func (h *Handler) BulkRemovePermission(c *gin.Context) {
	var input BulkPermissionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	if err := h.service.BulkRemovePermission(c.Request.Context(), input); err != nil {
		writeError(c, err, "BULK_REMOVE_PERMISSION_FAILED")
		return
	}
	response.OK(c, http.StatusOK, gin.H{"removed": true})
}

func (h *Handler) CreateAccessRequest(c *gin.Context) {
	var input AccessRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	request, err := h.service.CreateAccessRequest(c.Request.Context(), input)
	if err != nil {
		writeError(c, err, "CREATE_ACCESS_REQUEST_FAILED")
		return
	}
	response.OK(c, http.StatusCreated, request)
}

func (h *Handler) ListAccessRequests(c *gin.Context) {
	requests, err := h.service.ListAccessRequests(c.Request.Context())
	if err != nil {
		writeError(c, err, "LIST_ACCESS_REQUESTS_FAILED")
		return
	}
	response.OK(c, http.StatusOK, requests)
}

func (h *Handler) DecideAccessRequest(c *gin.Context) {
	var input AccessRequestDecisionInput
	_ = c.ShouldBindJSON(&input)
	request, err := h.service.DecideAccessRequest(c.Request.Context(), c.Param("id"), normalizeDecision(c.Param("decision")), input)
	if err != nil {
		writeError(c, err, "DECIDE_ACCESS_REQUEST_FAILED")
		return
	}
	response.OK(c, http.StatusOK, request)
}

func (h *Handler) CreateChangeRequest(c *gin.Context) {
	var input ChangeRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	request, err := h.service.CreateChangeRequest(c.Request.Context(), input)
	if err != nil {
		writeError(c, err, "CREATE_CHANGE_REQUEST_FAILED")
		return
	}
	response.OK(c, http.StatusCreated, request)
}

func (h *Handler) ListChangeRequests(c *gin.Context) {
	requests, err := h.service.ListChangeRequests(c.Request.Context())
	if err != nil {
		writeError(c, err, "LIST_CHANGE_REQUESTS_FAILED")
		return
	}
	response.OK(c, http.StatusOK, requests)
}

func (h *Handler) DecideChangeRequest(c *gin.Context) {
	var input AccessRequestDecisionInput
	_ = c.ShouldBindJSON(&input)
	request, err := h.service.DecideChangeRequest(c.Request.Context(), c.Param("id"), normalizeDecision(c.Param("decision")), input)
	if err != nil {
		writeError(c, err, "DECIDE_CHANGE_REQUEST_FAILED")
		return
	}
	response.OK(c, http.StatusOK, request)
}

func (h *Handler) Simulate(c *gin.Context) {
	var input SimulationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	result, err := h.service.Simulate(c.Request.Context(), input)
	if err != nil {
		writeError(c, err, "SIMULATE_RBAC_FAILED")
		return
	}
	response.OK(c, http.StatusOK, result)
}

func normalizeDecision(decision string) string {
	switch decision {
	case "approve":
		return "approved"
	case "reject":
		return "rejected"
	case "cancel":
		return "cancelled"
	case "apply":
		return "applied"
	default:
		return decision
	}
}

func bindRealmExportPayload(c *gin.Context) ([]byte, bool) {
	var input SyncPreviewRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return nil, false
	}
	if len(input.RealmExport) == 0 {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "realmExport is required")
		return nil, false
	}

	payload := input.RealmExport
	if !json.Valid(payload) {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "realmExport must be a JSON object")
		return nil, false
	}
	return payload, true
}

func writeError(c *gin.Context, err error, code string) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
	case errors.Is(err, ErrPermissionMissing):
		response.Fail(c, http.StatusBadRequest, "UNKNOWN_PERMISSION", err.Error())
	case errors.Is(err, ErrLastAccessRole):
		response.Fail(c, http.StatusConflict, "LAST_ACCESS_ROLE", err.Error())
	case errors.Is(err, sql.ErrNoRows):
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", err.Error())
	default:
		response.Fail(c, http.StatusInternalServerError, code, err.Error())
	}
}
