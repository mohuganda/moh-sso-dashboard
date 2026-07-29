package email

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/model"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func recipientValidationMessage(req SendEmailRequest, directRecipientCount int) string {
	if len(req.ToGroups) == 0 && len(req.ToGroupPaths) == 0 &&
		len(req.ToHealthContexts) == 0 && directRecipientCount == 0 {
		return "Select at least one email address, recipient group, or health context"
	}

	if len(req.ToGroups) > 0 || len(req.ToGroupPaths) > 0 {
		return "The selected groups contain no users with valid email addresses"
	}

	return "At least one valid recipient is required"
}

func (h *Handler) Send(c *gin.Context) {
	var req SendEmailRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Invalid email payload",
		)
		return
	}

	msg, err := mapSendEmailRequest(req)
	if err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"INVALID_EMAIL_PAYLOAD",
			err.Error(),
		)
		return
	}

	directRecipientCount := len(msg.To)

	msg, expandedRecipientCount, err := h.service.ExpandGroupRecipients(
		c.Request.Context(),
		msg,
		req.ToGroups,
		req.ToGroupPaths,
	)
	if err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"INVALID_RECIPIENT_GROUPS",
			err.Error(),
		)
		return
	}
	contextRecipients, err := h.service.ResolveHealthContextEmailRecipients(
		c.Request.Context(),
		c.GetString("user_id"),
		req.ToHealthContexts,
		req.IncludeHealthContextDescendants,
	)
	if err != nil {
		response.Fail(c, http.StatusForbidden, "HEALTH_CONTEXT_FORBIDDEN", "One or more selected health contexts are not accessible")
		return
	}
	msg = addUniqueAddresses(msg, contextRecipients)

	if len(msg.To) == 0 {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			recipientValidationMessage(req, directRecipientCount),
		)
		return
	}

	_ = expandedRecipientCount

	if err := h.service.Send(c.Request.Context(), msg); err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to send email",
		)
		return
	}

	response.OK(
		c,
		http.StatusOK,
		MessageResponse{Message: "email sent successfully"},
	)
}

func (h *Handler) Queue(c *gin.Context) {
	var req SendEmailRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Invalid email payload",
		)
		return
	}

	msg, err := mapSendEmailRequest(req)
	if err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"INVALID_EMAIL_PAYLOAD",
			"invalid email payload",
		)
		return
	}

	directRecipientCount := len(msg.To)

	msg, _, err = h.service.ExpandGroupRecipients(
		c.Request.Context(),
		msg,
		req.ToGroups,
		req.ToGroupPaths,
	)
	if err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"INVALID_RECIPIENT_GROUPS",
			"Failed to resolve recipient groups",
		)
		return
	}
	contextRecipients, err := h.service.ResolveHealthContextEmailRecipients(
		c.Request.Context(),
		c.GetString("user_id"),
		req.ToHealthContexts,
		req.IncludeHealthContextDescendants,
	)
	if err != nil {
		response.Fail(c, http.StatusForbidden, "HEALTH_CONTEXT_FORBIDDEN", "One or more selected health contexts are not accessible")
		return
	}
	msg = addUniqueAddresses(msg, contextRecipients)

	if len(msg.To) == 0 {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			recipientValidationMessage(req, directRecipientCount),
		)
		return
	}

	if err := h.service.Queue(c.Request.Context(), msg); err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to queue email",
		)
		return
	}

	response.OK(c, http.StatusAccepted, MessageResponse{Message: "email queued successfully"})
}

func (h *Handler) PreviewRecipients(c *gin.Context) {
	var req EmailRecipientPreviewRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Invalid recipient preview payload",
		)
		return
	}

	recipients, err := h.service.ResolveGroupEmailRecipients(
		c.Request.Context(),
		req.ToGroups,
		req.ToGroupPaths,
	)
	if err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"INVALID_RECIPIENT_GROUPS",
			"Failed to resolve recipient groups",
		)
		return
	}
	contextRecipients, err := h.service.ResolveHealthContextEmailRecipients(
		c.Request.Context(),
		c.GetString("user_id"),
		req.ToHealthContexts,
		req.IncludeHealthContextDescendants,
	)
	if err != nil {
		response.Fail(c, http.StatusForbidden, "HEALTH_CONTEXT_FORBIDDEN", "One or more selected health contexts are not accessible")
		return
	}
	previewMessage := addUniqueAddresses(model.Message{To: recipients}, contextRecipients)

	response.OK(c, http.StatusOK, EmailRecipientPreviewResponse{
		RecipientCount: len(previewMessage.To),
		Recipients:     toAddressResponses(previewMessage.To),
	})
}

func (h *Handler) List(c *gin.Context) {
	limit, offset := parsePagination(c)

	items, err := h.service.List(c.Request.Context(), limit, offset)
	if err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to list emails",
		)
		return
	}

	response.OK(c, http.StatusOK, toOutboxMessageResponses(items))
}

func (h *Handler) GetByID(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Email id is required",
		)
		return
	}

	item, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(
			c,
			http.StatusNotFound,
			"NOT_FOUND",
			"Email not found",
		)
		return
	}

	response.OK(c, http.StatusOK, toOutboxMessageResponse(*item))
}

func (h *Handler) ListByStatus(c *gin.Context) {
	status := strings.TrimSpace(strings.ToUpper(c.Param("status")))
	if status == "" {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Status is required",
		)
		return
	}

	limit, offset := parsePagination(c)

	items, err := h.service.ListByStatus(c.Request.Context(), status, limit, offset)
	if err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to list emails by status",
		)
		return
	}

	response.OK(c, http.StatusOK, toOutboxMessageResponses(items))
}

func (h *Handler) Retry(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Email id is required",
		)
		return
	}

	item, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(
			c,
			http.StatusNotFound,
			"NOT_FOUND",
			"Email not found",
		)
		return
	}

	msg := item.Message
	msg.ID = ""

	if err := h.service.Queue(c.Request.Context(), msg); err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to retry email",
		)
		return
	}

	response.OK(c, http.StatusAccepted, MessageResponse{Message: "email re-queued successfully"})
}

func (h *Handler) Delete(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Email id is required",
		)
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to delete email",
		)
		return
	}

	response.OK(c, http.StatusOK, MessageResponse{Message: "email deleted successfully"})
}

func parsePagination(c *gin.Context) (int32, int32) {
	limit := int32(20)
	offset := int32(0)

	if v := strings.TrimSpace(c.Query("limit")); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 32); err == nil && parsed > 0 {
			limit = int32(parsed)
		}
	}

	if v := strings.TrimSpace(c.Query("offset")); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 32); err == nil && parsed >= 0 {
			offset = int32(parsed)
		}
	}

	return limit, offset
}
