package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/service"
	"github.com/moh-sso-dashboard/internal/utils"
)

type AnnouncementHandler struct {
	announcementService *service.AnnouncementService
	auditService        *service.AuditService
}

func NewAnnouncementHandler(
	announcementService *service.AnnouncementService,
	auditService *service.AuditService,
) *AnnouncementHandler {
	return &AnnouncementHandler{
		announcementService: announcementService,
		auditService:        auditService,
	}
}

type createAnnouncementRequest struct {
	Title    string  `json:"title" binding:"required"`
	Message  string  `json:"message" binding:"required"`
	Tag      string  `json:"tag" binding:"required"`
	Priority int32   `json:"priority"`
	LinkURL  *string `json:"link_url"`
}

func (h *AnnouncementHandler) ListAnnouncements(c *gin.Context) {

	// default limit
	var limit int32 = 20

	// allow override via query param
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = int32(parsed)
		}
	}

	items, err := h.announcementService.ListAnnouncements(
		c.Request.Context(),
		limit,
	)

	if err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"failed to fetch announcements",
		)
		return
	}

	response.OK(c, http.StatusOK, items)
}

func (h *AnnouncementHandler) CreateAnnouncement(c *gin.Context) {
	var req createAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	userID := utils.ToNullUUID(c.GetString("user_id"))

	params := db.CreateAnnouncementParams{
		Title:   req.Title,
		Message: req.Message,
		Tag:     req.Tag,
		Priority: sql.NullInt32{
			Int32: req.Priority,
			Valid: true,
		},
		LinkUrl: sql.NullString{
			String: *req.LinkURL,
			Valid:  true,
		},
	}

	if userID.Valid {
		params.CreatedBy = userID
	}

	item, err := h.announcementService.CreateAnnouncement(c.Request.Context(), params)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to create announcement", err.Error())
		return
	}

	// Optional audit logging
	if h.auditService != nil {
		_ = h.auditService.Log(
			c.Request.Context(),
			userID,
			"ANNOUNCEMENT_CREATED",
			map[string]any{
				"announcement_id": item.ID.String(),
				"title":           item.Title,
				"tag":             item.Tag,
				"priority":        item.Priority,
			},
		)
	}

	response.OK(c, http.StatusCreated, item)
}

func (h *AnnouncementHandler) DeleteAnnouncement(c *gin.Context) {
	idParam := c.Param("id")
	if idParam == "" {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "announcement id is required")
		return
	}

	announcementID, err := uuid.Parse(idParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "invalid announcement id")
		return
	}

	err = h.announcementService.DeleteAnnouncement(c.Request.Context(), announcementID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete announcement")
		return
	}

	userID := utils.ToNullUUID(c.GetString("user_id"))

	// Optional audit logging
	if h.auditService != nil {
		_ = h.auditService.Log(
			c.Request.Context(),
			userID,
			"ANNOUNCEMENT_DELETED",
			map[string]any{
				"announcement_id": announcementID,
			},
		)
	}

	response.OK(c, http.StatusOK, "Successfully deleted announcement")
}
