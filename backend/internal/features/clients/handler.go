package clients

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/cache"
	"github.com/moh-sso-dashboard/internal/http/apierror"
	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/service"
	"github.com/moh-sso-dashboard/internal/utils"
)

type Handler struct {
	service      *service.ClientService
	auditService *service.AuditService
	cache        *cache.RedisCache
}

func NewHandler(
	s *service.ClientService,
	audit *service.AuditService,
	cache *cache.RedisCache,
) *Handler {
	return &Handler{
		service:      s,
		auditService: audit,
		cache:        cache,
	}
}

/* =========================================================
 * Create Client
 * ========================================================= */
func (h *Handler) CreateClient(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("user_id"))

	var req service.CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.audit(c, "client.create_failed", map[string]any{
			"reason": "invalid_body",
		})

		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Invalid request payload",
		)
		return
	}

	client, err := h.service.CreateClient(
		c.Request.Context(),
		req,
		userID,
	)
	if err != nil {
		h.audit(c, "client.create_failed", map[string]any{
			"client_id": req.ClientID,
			"reason":    err.Error(),
		})

		var apiErr *apierror.APIError
		if errors.As(err, &apiErr) {
			response.Fail(c, apiErr.HTTPStatus, apiErr.Code, apiErr.Message)
			return
		}

		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to create client",
		)
		return
	}

	h.audit(c, "client.create_success", map[string]any{
		"client_id": client.ClientID,
	})

	response.OK(c, http.StatusCreated, client)
}

/* =========================================================
 * Get Client
 * ========================================================= */
func (h *Handler) GetClient(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.audit(
			c,
			"client.get_failed",
			map[string]interface{}{
				"reason": "missing_id",
			},
		)

		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Client ID is required",
		)
		return
	}

	uid, _ := uuid.Parse(id)

	client, err := h.service.GetClient(uid)
	if err != nil {
		h.audit(
			c,
			"client.get_failed",
			map[string]interface{}{
				"client_id": id,
			},
		)

		response.Fail(
			c,
			http.StatusNotFound,
			"CLIENT_NOT_FOUND",
			"Client not found",
		)
		return
	}

	h.audit(c, "client.get_success", map[string]any{
		"client_id": id,
	})

	response.OK(c, http.StatusOK, client)
}

/* =========================================================
 * List Clients
 * ========================================================= */
func (h *Handler) ListClients(c *gin.Context) {

	var (
		clients   []model.Client
		fromCache bool
	)

	cacheKey := listClientsCacheKey()

	if h.cache != nil {
		if ok, _ := h.cache.Get(c.Request.Context(), cacheKey, &clients); ok {
			fromCache = true
		}
	}

	if !fromCache {
		var err error
		clients, err = h.service.ListClients()
		if err != nil {
			response.Fail(
				c,
				http.StatusInternalServerError,
				"INTERNAL_ERROR",
				"Failed to list clients",
			)
			return
		}

		// store in Redis
		if h.cache != nil {
			_ = h.cache.Set(
				c.Request.Context(),
				cacheKey,
				clients,
				60*time.Second,
			)
		}
	}

	clientRoles := c.MustGet("client_roles").(map[string][]string)
	isAdmin := c.GetBool("is_admin")
	filtered := filterAccessibleClients(clients, clientRoles, isAdmin)

	h.audit(c, "client.list", nil)
	response.OK(c, http.StatusOK, filtered)
}

/* =========================================================
 * Delete Client
 * ========================================================= */
func (h *Handler) DeleteClient(c *gin.Context) {
	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid client ID")
		return
	}

	adminID, _ := uuid.Parse(c.GetString("user_id"))

	if err := h.service.DeleteClient(
		c.Request.Context(),
		clientID,
		adminID,
	); err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to delete client",
		)
		return
	}

	h.audit(c, "client.delete_success", map[string]any{
		"client_id": clientID.String(),
	})

	c.Status(http.StatusNoContent)
}

/* =========================================================
 * Toggle Client Enabled
 * ========================================================= */
func (h *Handler) ToggleClientEnabled(c *gin.Context) {
	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid client ID")
		return
	}

	var body toggleClientEnabledRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Enabled flag is required",
		)
		return
	}

	adminID, _ := uuid.Parse(c.GetString("user_id"))

	if err := h.service.ToggleClientEnabled(
		c.Request.Context(),
		clientID,
		body.Enabled,
		adminID,
	); err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to update client status",
		)
		return
	}

	h.audit(c, "client.toggled", map[string]any{
		"client_id": clientID.String(),
		"enabled":   body.Enabled,
	})

	c.Status(http.StatusNoContent)
}

/* =========================================================
 * Client Roles
 * ========================================================= */

// POST /clients/:id/roles
func (h *Handler) CreateClientRole(c *gin.Context) {
	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid client ID")
		return
	}

	var body *model.CreateClientRoleRequest
	if err := c.ShouldBindJSON(&body); err != nil || body.Role == "" {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Role is required",
		)
		return
	}

	adminID, _ := uuid.Parse(c.GetString("user_id"))

	if err := h.service.CreateClientRole(
		c.Request.Context(),
		clientID,
		body,
		adminID,
	); err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to create client role",
		)
		return
	}

	h.audit(c, "client.role_created", map[string]any{
		"client_id": clientID.String(),
		"role":      body.Role,
	})

	c.Status(http.StatusCreated)
}

// GET /clients/:id/roles
func (h *Handler) ListClientRoles(c *gin.Context) {
	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid client ID")
		return
	}

	roles, err := h.service.ListClientRoles(
		c.Request.Context(),
		clientID,
	)
	if err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to list client roles",
		)
		return
	}

	h.audit(c, "client.roles_listed", map[string]any{
		"client_id": clientID.String(),
	})

	response.OK(c, http.StatusOK, roles)
}

// DELETE /clients/:id/roles/:role
func (h *Handler) DeleteClientRole(c *gin.Context) {
	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid client ID")
		return
	}

	role := c.Param("role")
	if role == "" {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Role is required",
		)
		return
	}

	adminID, _ := uuid.Parse(c.GetString("user_id"))

	if err := h.service.DeleteClientRole(
		c.Request.Context(),
		clientID,
		role,
		adminID,
	); err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to delete client role",
		)
		return
	}

	h.audit(c, "client.role_deleted", map[string]any{
		"client_id": clientID.String(),
		"role":      role,
	})

	c.Status(http.StatusNoContent)
}

/* =========================================================
 * Audit helper
 * ========================================================= */
func (h *Handler) audit(
	c *gin.Context,
	action string,
	meta map[string]any,
) {
	if meta == nil {
		meta = map[string]any{}
	}

	meta["ip"] = c.ClientIP()
	meta["user_agent"] = c.Request.UserAgent()

	_ = h.auditService.Log(
		c.Request.Context(),
		utils.ToNullUUID(c.GetString("user_id")),
		action,
		meta,
	)
}

func listClientsCacheKey() string {
	return "clients:all"
}
