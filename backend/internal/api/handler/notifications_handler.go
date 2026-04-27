package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/service"
)

type NotificationsHandler struct {
	NotificationsSvc service.NotificationsService
}

func NewNotificationsHandler(
	svc service.NotificationsService,
) *NotificationsHandler {
	return &NotificationsHandler{
		NotificationsSvc: svc,
	}
}

/* =========================================================
 * Create notification
 * ========================================================= */

	func (h *NotificationsHandler) Notify(c *gin.Context) {
		var input model.Notification

		if err := c.ShouldBindJSON(&input); err != nil {
			response.Fail(
				c,
				http.StatusBadRequest,
				"VALIDATION_FAILED",
				"Invalid notification payload",
			)
			return
		}

		n, err := h.NotificationsSvc.Notify(c.Request.Context(), input)
		if err != nil {
			response.Fail(
				c,
				http.StatusInternalServerError,
				"INTERNAL_ERROR",
				"Failed to create notification",
			)
			return
		}

		response.OK(c, http.StatusCreated, n)
	}

/* =========================================================
 * List notifications
 * ========================================================= */

func (h *NotificationsHandler) ListNotifications(c *gin.Context) {
	role := getTargetRole(c)

	// unread (optional)
	var unread *bool
	if v := c.Query("unread"); v != "" {
		b := v == "true"
		unread = &b
	}

	// limit
	limit := int32(20)
	if v := c.Query("limit"); v != "" {
		if l, err := strconv.Atoi(v); err == nil && l > 0 {
			limit = int32(l)
		}
	}

	// offset
	offset := int32(0)
	if v := c.Query("offset"); v != "" {
		if o, err := strconv.Atoi(v); err == nil && o >= 0 {
			offset = int32(o)
		}
	}

	notifications, err := h.NotificationsSvc.ListNotifications(
		c.Request.Context(),
		role,
		unread,
		limit,
		offset,
	)
	if err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to list notifications",
		)
		return
	}

	response.OK(c, http.StatusOK, notifications)
}

/* =========================================================
 * Get notification by ID
 * ========================================================= */

func (h *NotificationsHandler) GetNotificationByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Notification ID is required",
		)
		return
	}

	n, err := h.NotificationsSvc.GetNotificationByID(
		c.Request.Context(),
		id,
	)
	if err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to fetch notification",
		)
		return
	}

	if n == nil {
		response.Fail(
			c,
			http.StatusNotFound,
			"NOTIFICATION_NOT_FOUND",
			"Notification not found",
		)
		return
	}

	response.OK(c, http.StatusOK, n)
}

/* =========================================================
 * Mark notification as read
 * ========================================================= */

func (h *NotificationsHandler) MarkNotificationAsRead(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Notification ID is required",
		)
		return
	}

	if err := h.NotificationsSvc.MarkNotificationAsRead(
		c.Request.Context(),
		id,
	); err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to mark notification as read",
		)
		return
	}

	c.Status(http.StatusNoContent)
}

/* =========================================================
 * Delete notification
 * ========================================================= */

func (h *NotificationsHandler) DeleteNotification(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Notification ID is required",
		)
		return
	}

	if err := h.NotificationsSvc.DeleteNotification(
		c.Request.Context(),
		id,
	); err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to delete notification",
		)
		return
	}

	c.Status(http.StatusNoContent)
}

/* =========================================================
 * Delete old notifications
 * ========================================================= */

func (h *NotificationsHandler) DeleteOldNotifications(c *gin.Context) {
	if err := h.NotificationsSvc.DeleteOldNotifications(
		c.Request.Context(),
	); err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to cleanup notifications",
		)
		return
	}

	c.Status(http.StatusNoContent)
}

/* =========================================================
 * Counts
 * ========================================================= */

func (h *NotificationsHandler) CountNotifications(c *gin.Context) {
	role := getTargetRole(c)

	count, err := h.NotificationsSvc.CountNotifications(
		c.Request.Context(),
		role,
	)
	if err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to count notifications",
		)
		return
	}

	response.OK(c, http.StatusOK, gin.H{
		"count": count,
	})
}

func (h *NotificationsHandler) CountUnreadNotificationsCount(c *gin.Context) {
	role := getTargetRole(c)

	count, err := h.NotificationsSvc.CountUnreadNotificationsCount(
		c.Request.Context(),
		role,
	)
	if err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to count unread notifications",
		)
		return
	}

	response.OK(c, http.StatusOK, gin.H{
		"count": count,
	})
}

/* =========================================================
 * Helpers
 * ========================================================= */

func getTargetRole(c *gin.Context) string {
	if isAdmin, ok := c.Get("is_admin"); ok && isAdmin == true {
		return "admin"
	}

	if roles, ok := c.Get("client_roles"); ok {
		if cr, ok := roles.(map[string][]string); ok && len(cr) > 0 {
			return "user"
		}
	}

	return "user"
}
