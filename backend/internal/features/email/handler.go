package email

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/model"
	repository "github.com/moh-sso-dashboard/internal/repository/email"
	"github.com/moh-sso-dashboard/internal/service"
)

type EmailAddressRequest struct {
	Name  string `json:"name"`
	Email string `json:"email" binding:"required,email"`
}

type EmailAttachmentRequest struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type,omitempty"`
	Path        string `json:"path,omitempty"`
	ContentID   string `json:"content_id,omitempty"`
	Inline      bool   `json:"inline,omitempty"`
}

type SendEmailRequest struct {
	ID           string                   `json:"id,omitempty"`
	From         *EmailAddressRequest     `json:"from,omitempty"`
	To           []EmailAddressRequest    `json:"to" binding:"required,min=1,dive"`
	Cc           []EmailAddressRequest    `json:"cc,omitempty"`
	Bcc          []EmailAddressRequest    `json:"bcc,omitempty"`
	ReplyTo      []EmailAddressRequest    `json:"reply_to,omitempty"`
	Subject      string                   `json:"subject" binding:"required"`
	TextBody     string                   `json:"text_body,omitempty"`
	HTMLBody     string                   `json:"html_body,omitempty"`
	TemplateName string                   `json:"template_name,omitempty"`
	TemplateData map[string]any           `json:"template_data,omitempty"`
	Attachments  []EmailAttachmentRequest `json:"attachments,omitempty"`
	Headers      map[string]string        `json:"headers,omitempty"`
	Metadata     map[string]string        `json:"metadata,omitempty"`
	ScheduledAt  *string                  `json:"scheduled_at,omitempty"`
}

type Handler struct {
	service service.EmailService
	repo    repository.EmailRepository
}

func NewHandler(service service.EmailService, repo repository.EmailRepository) *Handler {
	return &Handler{
		service: service,
		repo:    repo,
	}
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

	if err := h.service.Send(c.Request.Context(), msg); err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to send email",
		)
		return
	}

	response.OK(c, http.StatusOK, gin.H{
		"message": "email sent successfully",
	})
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
			err.Error(),
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

	response.OK(c, http.StatusAccepted, gin.H{
		"message": "email queued successfully",
	})
}

func (h *Handler) List(c *gin.Context) {
	limit, offset := parsePagination(c)

	items, err := h.repo.List(c.Request.Context(), limit, offset)
	if err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to list emails",
		)
		return
	}

	response.OK(c, http.StatusOK, items)
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

	item, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(
			c,
			http.StatusNotFound,
			"NOT_FOUND",
			"Email not found",
		)
		return
	}

	response.OK(c, http.StatusOK, item)
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

	items, err := h.repo.ListByStatus(c.Request.Context(), status, limit, offset)
	if err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to list emails by status",
		)
		return
	}

	response.OK(c, http.StatusOK, items)
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

	item, err := h.repo.GetByID(c.Request.Context(), id)
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

	response.OK(c, http.StatusAccepted, gin.H{
		"message": "email re-queued successfully",
	})
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

	if err := h.repo.DeleteByID(c.Request.Context(), id); err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to delete email",
		)
		return
	}

	response.OK(c, http.StatusOK, gin.H{
		"message": "email deleted successfully",
	})
}

func mapSendEmailRequest(req SendEmailRequest) (model.Message, error) {
	var scheduledAt *time.Time

	if req.ScheduledAt != nil && strings.TrimSpace(*req.ScheduledAt) != "" {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.ScheduledAt))
		if err != nil {
			return model.Message{}, err
		}

		scheduledAt = &t
	}

	msg := model.Message{
		ID:           strings.TrimSpace(req.ID),
		To:           mapAddresses(req.To),
		Cc:           mapAddresses(req.Cc),
		Bcc:          mapAddresses(req.Bcc),
		ReplyTo:      mapAddresses(req.ReplyTo),
		Subject:      strings.TrimSpace(req.Subject),
		TextBody:     req.TextBody,
		HTMLBody:     req.HTMLBody,
		TemplateName: strings.TrimSpace(req.TemplateName),
		TemplateData: req.TemplateData,
		Attachments:  mapAttachments(req.Attachments),
		Headers:      req.Headers,
		Metadata:     req.Metadata,
		ScheduledAt:  scheduledAt,
	}

	if req.From != nil {
		msg.From = &model.Address{
			Name:  strings.TrimSpace(req.From.Name),
			Email: strings.TrimSpace(req.From.Email),
		}
	}

	return msg, nil
}

func mapAddresses(in []EmailAddressRequest) []model.Address {
	out := make([]model.Address, 0, len(in))

	for _, a := range in {
		out = append(out, model.Address{
			Name:  strings.TrimSpace(a.Name),
			Email: strings.TrimSpace(a.Email),
		})
	}

	return out
}

func mapAttachments(in []EmailAttachmentRequest) []model.Attachment {
	out := make([]model.Attachment, 0, len(in))

	for _, a := range in {
		out = append(out, model.Attachment{
			FileName:    strings.TrimSpace(a.FileName),
			ContentType: strings.TrimSpace(a.ContentType),
			Path:        strings.TrimSpace(a.Path),
			ContentID:   strings.TrimSpace(a.ContentID),
			Inline:      a.Inline,
		})
	}

	return out
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
