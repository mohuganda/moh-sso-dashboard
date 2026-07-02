package announcements

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/http/response"
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

func getAnnouncementAttachmentID(c *gin.Context) (uuid.UUID, bool) {
	idParam := c.Param("attachmentId")
	if idParam == "" {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "attachment id is required")
		return uuid.Nil, false
	}

	attachmentID, err := uuid.Parse(idParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "invalid attachment id")
		return uuid.Nil, false
	}

	return attachmentID, true
}

func getCurrentUserID(c *gin.Context) (uuid.UUID, bool) {
	userID := utils.ToNullUUID(c.GetString("user_id"))
	if !userID.Valid {
		response.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "user not authenticated")
		return uuid.Nil, false
	}

	return userID.UUID, true
}

func (h *Handler) announcementResponsesWithAttachments(
	ctx *gin.Context,
	items []announcement,
	adminLinks bool,
) []AnnouncementResponse {
	res := make([]AnnouncementResponse, len(items))
	for i, item := range items {
		attachments, err := h.announcementService.ListAttachments(ctx.Request.Context(), item.ID)
		if err != nil {
			res[i] = toAnnouncementResponse(item)
			continue
		}
		if adminLinks {
			res[i] = toAnnouncementResponseWithAttachments(item, attachments)
			h.attachAnnouncementAudience(ctx, item.ID, &res[i])
		} else {
			res[i] = toUserAnnouncementResponseWithAttachments(item, attachments)
		}
	}
	return res
}

func (h *Handler) attachAnnouncementAudience(
	c *gin.Context,
	announcementID uuid.UUID,
	res *AnnouncementResponse,
) {
	if h == nil || h.announcementService == nil || res == nil || announcementID == uuid.Nil {
		return
	}

	clientIDs, err := h.announcementService.ListClientAudience(c.Request.Context(), announcementID)
	if err != nil {
		return
	}

	roleNames, err := h.announcementService.ListRoleAudience(c.Request.Context(), announcementID)
	if err != nil {
		return
	}

	userIDs, err := h.announcementService.ListUserAudience(c.Request.Context(), announcementID)
	if err != nil {
		return
	}

	*res = withAnnouncementAudience(*res, clientIDs, roleNames, userIDs)
}

func (h *Handler) ListAnnouncementsAdmin(c *gin.Context) {
	limit := getPageLimit(c, 20)
	offset := getPageOffset(c)

	items, err := h.announcementService.ListAnnouncementsAdminPage(
		c.Request.Context(),
		limit,
		offset,
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch announcements")
		return
	}

	response.OK(c, http.StatusOK, h.announcementResponsesWithAttachments(c, items, true))
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

	attachments, _ := h.announcementService.ListAttachments(c.Request.Context(), item.ID)
	res := toAnnouncementResponseWithAttachments(item, attachments)
	h.attachAnnouncementAudience(c, item.ID, &res)
	response.OK(c, http.StatusOK, res)
}

func (h *Handler) ListPublicAnnouncements(c *gin.Context) {
	ctx := c.Request.Context()

	limit, err := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 32)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "request failed")
		return
	}

	offset, err := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 32)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "request failed")
		return
	}

	announcements, err := h.announcementService.ListPublicAnnouncements(
		ctx,
		int32(limit),
		int32(offset),
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load public announcements")
		return
	}

	response.OK(c, http.StatusOK, h.announcementResponsesWithAttachments(c, announcements, false))
}

func (h *Handler) CreateAnnouncement(c *gin.Context) {
	var req createAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "request failed")
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

	input := CreateAnnouncementInput{
		Title:         strings.TrimSpace(req.Title),
		Message:       strings.TrimSpace(req.Message),
		Summary:       nullableString(req.Summary),
		Level:         req.Level,
		Tag:           nullableString(req.Tag),
		LinkURL:       nullableString(req.LinkURL),
		LinkLabel:     nullableString(req.LinkLabel),
		Priority:      req.Priority,
		IsPinned:      req.IsPinned,
		Status:        req.Status,
		PublishAt:     publishAt,
		ExpiresAt:     expiresAt,
		AudienceType:  req.AudienceType,
		NotifyByEmail: req.NotifyByEmail,
		NotifyBySMS:   req.NotifyBySMS,
		SMSMessage:    nullableString(req.SMSMessage),
		CreatedBy:     userID,
	}

	item, err := h.announcementService.CreateAnnouncementFromInput(c.Request.Context(), input)
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
				"notify_by_email": item.NotifyByEmail,
				"notify_by_sms":   req.NotifyBySMS,
			},
		)
	}

	res := toAnnouncementResponse(item)
	h.attachAnnouncementAudience(c, item.ID, &res)
	response.OK(c, http.StatusCreated, res)
}

func (h *Handler) UpdateAnnouncement(c *gin.Context) {
	announcementID, ok := getAnnouncementID(c)
	if !ok {
		return
	}

	var req updateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "request failed")
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

	input := UpdateAnnouncementInput{
		ID:            announcementID,
		Title:         strings.TrimSpace(req.Title),
		Message:       strings.TrimSpace(req.Message),
		Summary:       nullableString(req.Summary),
		Level:         req.Level,
		Tag:           nullableString(req.Tag),
		LinkURL:       nullableString(req.LinkURL),
		LinkLabel:     nullableString(req.LinkLabel),
		Priority:      req.Priority,
		IsPinned:      req.IsPinned,
		PublishAt:     publishAt,
		ExpiresAt:     expiresAt,
		AudienceType:  req.AudienceType,
		NotifyByEmail: req.NotifyByEmail,
		NotifyBySMS:   req.NotifyBySMS,
		SMSMessage:    nullableString(req.SMSMessage),
		UpdatedBy:     userID,
	}

	item, err := h.announcementService.UpdateAnnouncementFromInput(c.Request.Context(), input)
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
				"notify_by_email": item.NotifyByEmail,
				"notify_by_sms":   req.NotifyBySMS,
			},
		)
	}

	res := toAnnouncementResponse(item)
	h.attachAnnouncementAudience(c, item.ID, &res)
	response.OK(c, http.StatusOK, res)
}

func (h *Handler) PublishAnnouncementNow(c *gin.Context) {
	announcementID, ok := getAnnouncementID(c)
	if !ok {
		return
	}

	var req publishAnnouncementRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "request failed")
			return
		}
	}

	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	item, err := h.announcementService.PublishAnnouncementNow(
		c.Request.Context(),
		announcementID,
		userID,
		announcementEmailOptionsFromPublishRequest(req),
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to publish announcement"+"request failed")
		return
	}

	if h.auditService != nil {
		_ = h.auditService.Log(
			c.Request.Context(),
			uuid.NullUUID{UUID: userID, Valid: true},
			"ANNOUNCEMENT_PUBLISHED",
			map[string]any{
				"announcement_id":                    item.ID.String(),
				"title":                              item.Title,
				"notify_by_email":                    item.NotifyByEmail,
				"email_notification_sent_at_present": item.EmailNotificationSentAt.Valid,
				"notify_by_sms":                      item.NotifyBySms,
				"sms_notification_queued_at_present": item.SmsNotificationQueuedAt.Valid,
			},
		)
	}

	res := toAnnouncementResponse(item)
	response.OK(c, http.StatusOK, res)
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

	item, err := h.announcementService.MoveAnnouncementToDraftByUser(
		c.Request.Context(),
		announcementID,
		userID,
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

	res := toAnnouncementResponse(item)
	response.OK(c, http.StatusOK, res)
}

func (h *Handler) ScheduleAnnouncement(c *gin.Context) {
	announcementID, ok := getAnnouncementID(c)
	if !ok {
		return
	}

	var req scheduleAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "request failed")
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

	item, err := h.announcementService.ScheduleAnnouncementByUser(
		c.Request.Context(),
		announcementID,
		publishAt,
		userID,
		announcementEmailOptionsFromScheduleRequest(req, publishAt),
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
				"notify_by_sms":   item.NotifyBySms,
			},
		)
	}

	response.OK(c, http.StatusOK, toAnnouncementResponse(item))
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

	response.OK(c, http.StatusOK, toAnnouncementResponse(item))
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

	response.OK(c, http.StatusOK, toAnnouncementResponse(item))
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

	response.OK(c, http.StatusOK, MessageResponse{Message: "announcement deleted successfully"})
}

func (h *Handler) UploadAnnouncementAttachment(c *gin.Context) {
	announcementID, ok := getAnnouncementID(c)
	if !ok {
		return
	}

	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	var input UploadAnnouncementAttachmentInput
	includeInEmail := true

	contentType := c.GetHeader("Content-Type")
	if strings.HasPrefix(strings.ToLower(contentType), "multipart/form-data") {
		file, err := c.FormFile("file")
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "file is required")
			return
		}

		reader, err := file.Open()
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "failed to open uploaded file")
			return
		}
		defer reader.Close()

		if raw := strings.TrimSpace(c.PostForm("include_in_email")); raw != "" {
			if parsed, err := strconv.ParseBool(raw); err == nil {
				includeInEmail = parsed
			}
		}

		sortOrder := int32(0)
		if raw := strings.TrimSpace(c.PostForm("sort_order")); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil {
				sortOrder = int32(parsed)
			}
		}

		input = UploadAnnouncementAttachmentInput{
			FileName:       file.Filename,
			ContentType:    file.Header.Get("Content-Type"),
			Reader:         reader,
			Size:           file.Size,
			IncludeInEmail: includeInEmail,
			Inline:         strings.EqualFold(c.PostForm("inline"), "true"),
			ContentID:      c.PostForm("content_id"),
			SortOrder:      sortOrder,
			UploadedBy:     userID,
		}
	} else {
		var req createAnnouncementAttachmentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "request failed")
			return
		}
		if req.IncludeInEmail != nil {
			includeInEmail = *req.IncludeInEmail
		}
		input = UploadAnnouncementAttachmentInput{
			FileName:       req.FileName,
			ContentType:    req.ContentType,
			DataBase64:     req.DataBase64,
			IncludeInEmail: includeInEmail,
			Inline:         req.Inline,
			ContentID:      req.ContentID,
			SortOrder:      req.SortOrder,
			UploadedBy:     userID,
		}
	}

	attachment, err := h.announcementService.UploadAttachment(c.Request.Context(), announcementID, input)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", "request failed")
		return
	}

	if h.auditService != nil {
		_ = h.auditService.Log(
			c.Request.Context(),
			uuid.NullUUID{UUID: userID, Valid: true},
			"ANNOUNCEMENT_ATTACHMENT_UPLOADED",
			map[string]any{
				"announcement_id": announcementID.String(),
				"attachment_id":   attachment.ID.String(),
				"file_name":       attachment.OriginalFileName,
			},
		)
	}

	response.OK(c, http.StatusCreated, toAnnouncementAttachmentResponse(announcementID, attachment))
}

func (h *Handler) ListAnnouncementAttachments(c *gin.Context) {
	announcementID, ok := getAnnouncementID(c)
	if !ok {
		return
	}

	attachments, err := h.announcementService.ListAttachments(c.Request.Context(), announcementID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list announcement attachments")
		return
	}

	response.OK(c, http.StatusOK, toAnnouncementAttachmentResponses(announcementID, attachments))
}

func (h *Handler) DownloadAnnouncementAttachment(c *gin.Context) {
	announcementID, ok := getAnnouncementID(c)
	if !ok {
		return
	}

	attachmentID, ok := getAnnouncementAttachmentID(c)
	if !ok {
		return
	}

	download, err := h.announcementService.OpenAttachmentDownload(c.Request.Context(), announcementID, attachmentID)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "announcement attachment not found")
		return
	}
	defer download.Reader.Close()

	contentType := nullStringValue(download.Attachment.ContentType)
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, download.Attachment.OriginalFileName))
	c.Header("Content-Type", contentType)
	_, _ = io.Copy(c.Writer, download.Reader)
}

func (h *Handler) UpdateAnnouncementAttachment(c *gin.Context) {
	announcementID, ok := getAnnouncementID(c)
	if !ok {
		return
	}
	attachmentID, ok := getAnnouncementAttachmentID(c)
	if !ok {
		return
	}

	var req updateAnnouncementAttachmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "request failed")
		return
	}

	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	attachment, err := h.announcementService.UpdateAttachment(
		c.Request.Context(),
		announcementID,
		attachmentID,
		UpdateAnnouncementAttachmentInput{
			IncludeInEmail: req.IncludeInEmail,
			Inline:         req.Inline,
			ContentID:      req.ContentID,
			SortOrder:      req.SortOrder,
			UpdatedBy:      userID,
		},
	)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toAnnouncementAttachmentResponse(announcementID, attachment))
}

func (h *Handler) DeleteAnnouncementAttachment(c *gin.Context) {
	announcementID, ok := getAnnouncementID(c)
	if !ok {
		return
	}
	attachmentID, ok := getAnnouncementAttachmentID(c)
	if !ok {
		return
	}

	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	attachment, err := h.announcementService.DeleteAttachment(c.Request.Context(), announcementID, attachmentID, userID)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", "request failed")
		return
	}

	if h.auditService != nil {
		_ = h.auditService.Log(
			c.Request.Context(),
			uuid.NullUUID{UUID: userID, Valid: true},
			"ANNOUNCEMENT_ATTACHMENT_DELETED",
			map[string]any{
				"announcement_id": announcementID.String(),
				"attachment_id":   attachment.ID.String(),
				"file_name":       attachment.OriginalFileName,
			},
		)
	}

	response.OK(c, http.StatusOK, MessageResponse{Message: "announcement attachment deleted successfully"})
}

func (h *Handler) SetAnnouncementPinned(c *gin.Context) {
	announcementID, ok := getAnnouncementID(c)
	if !ok {
		return
	}

	var req setPinnedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "request failed")
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

	response.OK(c, http.StatusOK, toAnnouncementResponse(item))
}

func (h *Handler) SetAnnouncementPriority(c *gin.Context) {
	announcementID, ok := getAnnouncementID(c)
	if !ok {
		return
	}

	var req setPriorityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "request failed")
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

	response.OK(c, http.StatusOK, toAnnouncementResponse(item))
}

func (h *Handler) GetAnnouncementStats(c *gin.Context) {
	stats, err := h.announcementService.GetAnnouncementStats(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch announcement stats")
		return
	}

	response.OK(c, http.StatusOK, toAnnouncementStatsResponse(stats))
}

func (h *Handler) ListActivePublishedAnnouncements(c *gin.Context) {
	limit := getPageLimit(c, 20)
	offset := getPageOffset(c)

	items, err := h.announcementService.ListActivePublishedAnnouncementsPage(
		c.Request.Context(),
		limit,
		offset,
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch active announcements")
		return
	}

	response.OK(c, http.StatusOK, h.announcementResponsesWithAttachments(c, items, false))
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

	items, err := h.announcementService.ListAnnouncementsForClientPage(
		c.Request.Context(),
		clientID,
		limit,
		offset,
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch announcements for client")
		return
	}

	response.OK(c, http.StatusOK, h.announcementResponsesWithAttachments(c, items, false))
}

func (h *Handler) ListAnnouncementsForRole(c *gin.Context) {
	roleName := strings.TrimSpace(c.Param("role_name"))
	if roleName == "" {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "role name is required")
		return
	}

	limit := getPageLimit(c, 20)
	offset := getPageOffset(c)

	items, err := h.announcementService.ListAnnouncementsForRolePage(
		c.Request.Context(),
		roleName,
		limit,
		offset,
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch announcements for role")
		return
	}

	response.OK(c, http.StatusOK, h.announcementResponsesWithAttachments(c, items, false))
}

func (h *Handler) ListAnnouncementsForUser(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	limit := getPageLimit(c, 20)
	offset := getPageOffset(c)

	items, err := h.announcementService.ListAnnouncementsForUserPage(
		c.Request.Context(),
		userID,
		limit,
		offset,
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch announcements for user")
		return
	}

	response.OK(c, http.StatusOK, h.announcementResponsesWithAttachments(c, items, false))
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

	response.OK(c, http.StatusOK, h.announcementResponsesWithAttachments(c, items, false))
}
