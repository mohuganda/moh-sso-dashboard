package storage_locations

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/http/response"
	sharedservice "github.com/moh-sso-dashboard/internal/service"
)

var (
	ErrInvalidID         = errors.New("invalid id")
	ErrStorageCodeExists = errors.New("storage location code already exists")
	ErrStorageNotFound   = errors.New("storage location not found")
)

type Handler struct {
	storageLocationService Service
	auditService           *sharedservice.AuditService
}

func NewHandler(
	storageLocationService Service,
	auditService *sharedservice.AuditService,
) *Handler {
	return &Handler{
		storageLocationService: storageLocationService,
		auditService:           auditService,
	}
}

func (h *Handler) Create(c *gin.Context) {
	var req struct {
		Code     string `json:"code" binding:"required"`
		Name     string `json:"name" binding:"required"`
		Provider string `json:"provider"`
		BaseUri  string `json:"base_uri"`
		IsActive bool   `json:"is_active" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}

	loc, err := h.storageLocationService.Create(c.Request.Context(), CreateStorageLocationInput{
		Code:     req.Code,
		Name:     req.Name,
		Provider: req.Provider,
		BaseUri:  req.BaseUri,
		IsActive: sql.NullBool{
			Bool: req.IsActive,
		},
	})
	if err != nil {
		if errors.Is(err, ErrStorageCodeExists) {
			response.Fail(c, http.StatusConflict, "CODE_EXISTS", err.Error())
			return
		}
		response.Fail(c, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
		return
	}

	response.OK(c, http.StatusCreated, loc)
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")

	loc, err := h.storageLocationService.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrInvalidID) {
			response.Fail(c, http.StatusBadRequest, "INVALID_ID", err.Error())
			return
		}
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	response.OK(c, http.StatusOK, loc)
}

func (h *Handler) ListActive(c *gin.Context) {
	locs, err := h.storageLocationService.ListActive(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}

	response.OK(c, http.StatusOK, locs)
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateStorageLocationInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}

	loc, err := h.storageLocationService.Update(c.Request.Context(), id, req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
		return
	}

	response.OK(c, http.StatusOK, loc)
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.storageLocationService.Delete(c.Request.Context(), id); err != nil {
		response.Fail(c, http.StatusInternalServerError, "DELETE_FAILED", err.Error())
		return
	}

	response.OK(c, http.StatusOK, gin.H{"message": "deleted"})
}
