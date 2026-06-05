package email

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/http/response"
	repository "github.com/moh-sso-dashboard/internal/repository/email"
	"github.com/moh-sso-dashboard/internal/service"
)

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
