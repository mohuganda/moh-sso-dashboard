package announcements

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/model"
	sharedservice "github.com/moh-sso-dashboard/internal/service"
	"github.com/moh-sso-dashboard/internal/utils"
)

type Handler struct {
	announcementService *Service
	auditService        *sharedservice.AuditService
}

func NewHandler(
	announcementService *Service,
	auditService *sharedservice.AuditService,
) *Handler {
	return &Handler{
		announcementService: announcementService,
		auditService:        auditService,
	}
}

func getPageLimit(c *gin.Context, defaultValue int32) int32 {
	limit := defaultValue
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = int32(parsed)
		}
	}
	return limit
}

func getPageOffset(c *gin.Context) int32 {
	if raw := c.Query("offset"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 0 {
			return int32(parsed)
		}
	}
	return 0
}

func getAnnouncementID(c *gin.Context) (uuid.UUID, bool) {
	idParam := c.Param("id")
	if idParam == "" {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "announcement id is required")
		return uuid.Nil, false
	}

	announcementID, err := uuid.Parse(idParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "invalid announcement id")
		return uuid.Nil, false
	}

	return announcementID, true
}

func getCurrentUserID(c *gin.Context) (uuid.UUID, bool) {
	userID := utils.ToNullUUID(c.GetString("user_id"))
	if !userID.Valid {
		response.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "user not authenticated")
		return uuid.Nil, false
	}

	return userID.UUID, true
}

func (h *Handler) ListAnnouncementsAdmin(c *gin.Context) {
	limit := getPageLimit(c, 20)
	offset := getPageOffset(c)

	items, err := h.announcementService.ListAnnouncementsAdmin(
		c.Request.Context(),
		db.ListAnnouncementsAdminParams{
			Limit:  limit,
			Offset: offset,
		},
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch announcements")
		return
	}

	res := make([]AnnouncementResponse, len(items))
	for i, item := range items {
		res[i] = toAnnouncementResponse(item)
	}

	response.OK(c, http.StatusOK, res)
}

func (h *Handler) GetAnnouncementByID(c *gin.Context) {
	announcementID, ok := getAnnouncementID(c)
	if !ok {
		return
	}

	item, err := h.announcementService.GetAnnouncementByID(c.Request.Context(), announcementID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch announcement")
		return
	}

	response.OK(c, http.StatusOK, toAnnouncementResponse(item))
}

func (h *Handler) ListPublicAnnouncements(c *gin.Context) {
	ctx := c.Request.Context()

	limit, err := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 32)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "", err.Error())
		return
	}

	offset, err := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 32)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "", err.Error())
		return
	}

	announcements, err := h.announcementService.ListPublicAnnouncements(
		ctx,
		int32(limit),
		int32(offset),
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to load public announcements", "")
		return
	}
	res := make([]AnnouncementResponse, len(announcements))
	for i, item := range announcements {
		res[i] = toAnnouncementResponse(item)
	}

	response.OK(c, http.StatusOK, res)
}

func (h *Handler) CreateAnnouncement(c *gin.Context) {
	var req createAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	publishAt, err := nullableTime(req.PublishAt)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "publish_at must be a valid RFC3339 datetime")
		return
	}

	expiresAt, err := nullableTime(req.ExpiresAt)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "expires_at must be a valid RFC3339 datetime")
		return
	}

	params := db.CreateAnnouncementParams{
		Title:        strings.TrimSpace(req.Title),
		Message:      strings.TrimSpace(req.Message),
		Summary:      nullableString(req.Summary),
		Level:        model.AnnouncementLevel(req.Level),
		Tag:          nullableString(req.Tag),
		LinkUrl:      nullableString(req.LinkURL),
		Priority:     req.Priority,
		IsPinned:     req.IsPinned,
		Status:       model.AnnouncementStatus(req.Status),
		PublishAt:    publishAt,
		ExpiresAt:    expiresAt,
		AudienceType: model.AnnouncementAudienceType(req.AudienceType),
		CreatedBy:    userID,
	}

	if strings.TrimSpace(req.Status) == "" {
		params.Status = model.AnnouncementStatusDRAFT
	}

	if strings.TrimSpace(req.AudienceType) == "" {
		params.AudienceType = model.AnnouncementAudienceTypeALLUSERS
	}

	item, err := h.announcementService.CreateAnnouncement(c.Request.Context(), params)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create announcement")
		return
	}

	clientIDs, err := parseUUIDList(req.ClientIDs)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "one or more client_ids are invalid")
		return
	}

	userIDs, err := parseUUIDList(req.UserIDs)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "one or more user_ids are invalid")
		return
	}

	if len(clientIDs) > 0 {
		if err := h.announcementService.ReplaceClientAudience(c.Request.Context(), item.ID, clientIDs); err != nil {
			response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to save client audience")
			return
		}
	}

	if len(req.RoleNames) > 0 {
		if err := h.announcementService.ReplaceRoleAudience(c.Request.Context(), item.ID, req.RoleNames); err != nil {
			response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to save role audience")
			return
		}
	}

	if len(userIDs) > 0 {
		if err := h.announcementService.ReplaceUserAudience(c.Request.Context(), item.ID, userIDs); err != nil {
			response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to save user audience")
			return
		}
	}

	if h.auditService != nil {
		_ = h.auditService.Log(
			c.Request.Context(),
			uuid.NullUUID{UUID: userID, Valid: true},
			"ANNOUNCEMENT_CREATED",
			map[string]any{
				"announcement_id": item.ID.String(),
				"title":           item.Title,
				"level":           item.Level,
				"priority":        item.Priority,
				"status":          item.Status,
				"is_pinned":       item.IsPinned,
			},
		)
	}

	response.OK(c, http.StatusCreated, item)
}

func (h *Handler) UpdateAnnouncement(c *gin.Context) {
	announcementID, ok := getAnnouncementID(c)
	if !ok {
		return
	}

	var req updateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	publishAt, err := nullableTime(req.PublishAt)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "publish_at must be a valid RFC3339 datetime")
		return
	}

	expiresAt, err := nullableTime(req.ExpiresAt)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "expires_at must be a valid RFC3339 datetime")
		return
	}

	params := db.UpdateAnnouncementParams{
		ID:           announcementID,
		Title:        strings.TrimSpace(req.Title),
		Message:      strings.TrimSpace(req.Message),
		Summary:      nullableString(req.Summary),
		Level:        model.AnnouncementLevel(req.Level),
		Tag:          nullableString(req.Tag),
		LinkUrl:      nullableString(req.LinkURL),
		Priority:     req.Priority,
		IsPinned:     req.IsPinned,
		PublishAt:    publishAt,
		ExpiresAt:    expiresAt,
		AudienceType: model.AnnouncementAudienceType(req.AudienceType),
		UpdatedBy: uuid.NullUUID{
			UUID:  userID,
			Valid: true,
		},
	}

	if strings.TrimSpace(req.AudienceType) == "" {
		params.AudienceType = model.AnnouncementAudienceTypeALLUSERS
	}

	item, err := h.announcementService.UpdateAnnouncement(c.Request.Context(), params)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update announcement")
		return
	}

	clientIDs, err := parseUUIDList(req.ClientIDs)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "one or more client_ids are invalid")
		return
	}

	userIDs, err := parseUUIDList(req.UserIDs)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "one or more user_ids are invalid")
		return
	}

	if err := h.announcementService.ReplaceClientAudience(c.Request.Context(), item.ID, clientIDs); err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update client audience")
		return
	}

	if err := h.announcementService.ReplaceRoleAudience(c.Request.Context(), item.ID, req.RoleNames); err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update role audience")
		return
	}

	if err := h.announcementService.ReplaceUserAudience(c.Request.Context(), item.ID, userIDs); err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update user audience")
		return
	}

	if h.auditService != nil {
		_ = h.auditService.Log(
			c.Request.Context(),
			uuid.NullUUID{UUID: userID, Valid: true},
			"ANNOUNCEMENT_UPDATED",
			map[string]any{
				"announcement_id": item.ID.String(),
				"title":           item.Title,
				"level":           item.Level,
				"priority":        item.Priority,
				"status":          item.Status,
				"is_pinned":       item.IsPinned,
			},
		)
	}

	response.OK(c, http.StatusOK, item)
}

func (h *Handler) PublishAnnouncementNow(c *gin.Context) {
	announcementID, ok := getAnnouncementID(c)
	if !ok {
		return
	}

	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	item, err := h.announcementService.PublishAnnouncementNow(c.Request.Context(), announcementID, userID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to publish announcement")
		return
	}

	if h.auditService != nil {
		_ = h.auditService.Log(
			c.Request.Context(),
			uuid.NullUUID{UUID: userID, Valid: true},
			"ANNOUNCEMENT_PUBLISHED",
			map[string]any{
				"announcement_id": item.ID.String(),
				"title":           item.Title,
			},
		)
	}

	response.OK(c, http.StatusOK, item)
}
func (h *Handler) MoveAnnouncementToDraft(c *gin.Context) {
	announcementID, ok := getAnnouncementID(c)
	if !ok {
		return
	}

	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	item, err := h.announcementService.MoveAnnouncementToDraft(
		c.Request.Context(),
		db.DraftAnnouncementParams{
			ID: announcementID,
			UpdatedBy: uuid.NullUUID{
				UUID:  userID,
				Valid: true,
			},
		},
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to draft announcement")
		return
	}

	if h.auditService != nil {
		_ = h.auditService.Log(
			c.Request.Context(),
			uuid.NullUUID{UUID: userID, Valid: true},
			"ANNOUNCEMENT_DRAFTED",
			map[string]any{
				"announcement_id": item.ID.String(),
			},
		)
	}

	response.OK(c, http.StatusOK, item)
}

func (h *Handler) ScheduleAnnouncement(c *gin.Context) {
	announcementID, ok := getAnnouncementID(c)
	if !ok {
		return
	}

	var req scheduleAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	publishAt, err := time.Parse(time.RFC3339, strings.TrimSpace(req.PublishAt))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "publish_at must be a valid RFC3339 datetime")
		return
	}

	item, err := h.announcementService.ScheduleAnnouncement(
		c.Request.Context(),
		db.ScheduleAnnouncementParams{
			ID: announcementID,
			PublishAt: sql.NullTime{
				Time:  publishAt,
				Valid: true,
			},
			UpdatedBy: uuid.NullUUID{
				UUID:  userID,
				Valid: true,
			},
		},
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to schedule announcement")
		return
	}

	if h.auditService != nil {
		_ = h.auditService.Log(
			c.Request.Context(),
			uuid.NullUUID{UUID: userID, Valid: true},
			"ANNOUNCEMENT_SCHEDULED",
			map[string]any{
				"announcement_id": item.ID.String(),
				"title":           item.Title,
				"publish_at":      item.PublishAt,
			},
		)
	}

	response.OK(c, http.StatusOK, item)
}

func (h *Handler) ArchiveAnnouncement(c *gin.Context) {
	announcementID, ok := getAnnouncementID(c)
	if !ok {
		return
	}

	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	item, err := h.announcementService.ArchiveAnnouncement(c.Request.Context(), announcementID, userID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to archive announcement")
		return
	}

	if h.auditService != nil {
		_ = h.auditService.Log(
			c.Request.Context(),
			uuid.NullUUID{UUID: userID, Valid: true},
			"ANNOUNCEMENT_ARCHIVED",
			map[string]any{
				"announcement_id": item.ID.String(),
				"title":           item.Title,
			},
		)
	}

	response.OK(c, http.StatusOK, item)
}

func (h *Handler) RestoreAnnouncement(c *gin.Context) {
	announcementID, ok := getAnnouncementID(c)
	if !ok {
		return
	}

	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	item, err := h.announcementService.RestoreAnnouncement(c.Request.Context(), announcementID, userID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to restore announcement")
		return
	}

	if h.auditService != nil {
		_ = h.auditService.Log(
			c.Request.Context(),
			uuid.NullUUID{UUID: userID, Valid: true},
			"ANNOUNCEMENT_RESTORED",
			map[string]any{
				"announcement_id": item.ID.String(),
				"title":           item.Title,
			},
		)
	}

	response.OK(c, http.StatusOK, item)
}

func (h *Handler) DeleteAnnouncement(c *gin.Context) {
	announcementID, ok := getAnnouncementID(c)
	if !ok {
		return
	}

	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	if err := h.announcementService.DeleteAnnouncement(c.Request.Context(), announcementID, userID); err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete announcement")
		return
	}

	if h.auditService != nil {
		_ = h.auditService.Log(
			c.Request.Context(),
			uuid.NullUUID{UUID: userID, Valid: true},
			"ANNOUNCEMENT_DELETED",
			map[string]any{
				"announcement_id": announcementID.String(),
			},
		)
	}

	response.OK(c, http.StatusOK, gin.H{
		"message": "announcement deleted successfully",
	})
}

func (h *Handler) SetAnnouncementPinned(c *gin.Context) {
	announcementID, ok := getAnnouncementID(c)
	if !ok {
		return
	}

	var req setPinnedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	item, err := h.announcementService.SetAnnouncementPinned(
		c.Request.Context(),
		announcementID,
		req.IsPinned,
		userID,
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update pinned state")
		return
	}

	if h.auditService != nil {
		_ = h.auditService.Log(
			c.Request.Context(),
			uuid.NullUUID{UUID: userID, Valid: true},
			"ANNOUNCEMENT_PIN_UPDATED",
			map[string]any{
				"announcement_id": item.ID.String(),
				"is_pinned":       item.IsPinned,
			},
		)
	}

	response.OK(c, http.StatusOK, item)
}

func (h *Handler) SetAnnouncementPriority(c *gin.Context) {
	announcementID, ok := getAnnouncementID(c)
	if !ok {
		return
	}

	var req setPriorityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	item, err := h.announcementService.SetAnnouncementPriority(
		c.Request.Context(),
		announcementID,
		req.Priority,
		userID,
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update priority")
		return
	}

	if h.auditService != nil {
		_ = h.auditService.Log(
			c.Request.Context(),
			uuid.NullUUID{UUID: userID, Valid: true},
			"ANNOUNCEMENT_PRIORITY_UPDATED",
			map[string]any{
				"announcement_id": item.ID.String(),
				"priority":        item.Priority,
			},
		)
	}

	response.OK(c, http.StatusOK, item)
}

func (h *Handler) GetAnnouncementStats(c *gin.Context) {
	stats, err := h.announcementService.GetAnnouncementStats(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch announcement stats")
		return
	}

	response.OK(c, http.StatusOK, stats)
}

func (h *Handler) ListActivePublishedAnnouncements(c *gin.Context) {
	limit := getPageLimit(c, 20)
	offset := getPageOffset(c)

	items, err := h.announcementService.ListActivePublishedAnnouncements(
		c.Request.Context(),
		db.ListActivePublishedAnnouncementsParams{
			Limit:  limit,
			Offset: offset,
		},
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch active announcements")
		return
	}

	response.OK(c, http.StatusOK, items)
}

func (h *Handler) ListAnnouncementsForClient(c *gin.Context) {
	clientIDParam := c.Param("client_id")
	if clientIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "client id is required")
		return
	}

	clientID, err := uuid.Parse(clientIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "invalid client id")
		return
	}

	limit := getPageLimit(c, 20)
	offset := getPageOffset(c)

	items, err := h.announcementService.ListAnnouncementsForClient(
		c.Request.Context(),
		db.ListAnnouncementsForClientParams{
			ClientID: clientID,
			Limit:    limit,
			Offset:   offset,
		},
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch announcements for client")
		return
	}

	response.OK(c, http.StatusOK, items)
}

func (h *Handler) ListAnnouncementsForRole(c *gin.Context) {
	roleName := strings.TrimSpace(c.Param("role_name"))
	if roleName == "" {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "role name is required")
		return
	}

	limit := getPageLimit(c, 20)
	offset := getPageOffset(c)

	items, err := h.announcementService.ListAnnouncementsForRole(
		c.Request.Context(),
		db.ListAnnouncementsForRoleParams{
			RoleName:   roleName,
			PageLimit:  limit,
			PageOffset: offset,
		},
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch announcements for role")
		return
	}

	response.OK(c, http.StatusOK, items)
}

func (h *Handler) ListAnnouncementsForUser(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	limit := getPageLimit(c, 20)
	offset := getPageOffset(c)

	items, err := h.announcementService.ListAnnouncementsForUser(
		c.Request.Context(),
		db.ListAnnouncementsForUserParams{
			UserID: userID,
			Limit:  limit,
			Offset: offset,
		},
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch announcements for user")
		return
	}

	response.OK(c, http.StatusOK, items)
}

func (h *Handler) ListMyAnnouncements(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	roleName := strings.TrimSpace(c.Query("role"))
	clientIDParam := strings.TrimSpace(c.Query("client_id"))

	var clientID uuid.UUID
	if clientIDParam != "" {
		parsed, err := uuid.Parse(clientIDParam)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "invalid client_id")
			return
		}
		clientID = parsed
	}

	limit := getPageLimit(c, 20)
	offset := getPageOffset(c)

	items, err := h.announcementService.ListMyAnnouncements(
		c.Request.Context(),
		userID,
		roleName,
		clientID,
		limit,
		offset,
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch my announcements")
		return
	}

	response.OK(c, http.StatusOK, items)
}
