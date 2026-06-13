package rbac

import (
	"database/sql"
	"errors"
	"net/http"

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
