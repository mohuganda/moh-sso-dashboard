package notifications

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/service"
)

type Handler struct {
	NotificationsSvc service.NotificationsService
}

func NewHandler(
	svc service.NotificationsService,
) *Handler {
	return &Handler{
		NotificationsSvc: svc,
	}
}

/* =========================================================
 * Create notification
 * ========================================================= */

func (h *Handler) Notify(c *gin.Context) {
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

	response.OK(c, http.StatusCreated, toNotificationResponse(*n))
}

/* =========================================================
 * List notifications
 * ========================================================= */

func (h *Handler) ListNotifications(c *gin.Context) {
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

	response.OK(c, http.StatusOK, toNotificationResponses(notifications))
}

/* =========================================================
 * Get notification by ID
 * ========================================================= */

func (h *Handler) GetNotificationByID(c *gin.Context) {
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

	response.OK(c, http.StatusOK, toNotificationResponse(*n))
}

/* =========================================================
 * Mark notification as read
 * ========================================================= */

func (h *Handler) MarkNotificationAsRead(c *gin.Context) {
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

func (h *Handler) DeleteNotification(c *gin.Context) {
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

func (h *Handler) DeleteOldNotifications(c *gin.Context) {
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

func (h *Handler) CountNotifications(c *gin.Context) {
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

	response.OK(c, http.StatusOK, CountResponse{Count: count})
}

func (h *Handler) CountUnreadNotificationsCount(c *gin.Context) {
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

	response.OK(c, http.StatusOK, CountResponse{Count: count})
}

/* =========================================================
 * Delivery history
 * ========================================================= */

func (h *Handler) ListNotificationDeliveries(c *gin.Context) {
	notificationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Valid notification ID is required",
		)
		return
	}

	deliveries, err := h.NotificationsSvc.ListNotificationDeliveries(
		c.Request.Context(),
		notificationID,
	)
	if err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to list notification deliveries",
		)
		return
	}

	response.OK(c, http.StatusOK, toNotificationDeliveryResponses(deliveries))
}

func (h *Handler) ListAllNotificationDeliveries(c *gin.Context) {
	limit := int32(20)
	if v := c.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = int32(parsed)
		}
	}

	offset := int32(0)
	if v := c.Query("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = int32(parsed)
		}
	}

	deliveries, total, err := h.NotificationsSvc.ListAllNotificationDeliveries(
		c.Request.Context(),
		service.NotificationDeliveryListFilter{
			Channel: c.Query("channel"),
			Status:  c.Query("status"),
			Limit:   limit,
			Offset:  offset,
		},
	)
	if err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to list notification deliveries",
		)
		return
	}

	response.OK(c, http.StatusOK, NotificationDeliveryListResponse{
		Items:  toNotificationDeliveryResponses(deliveries),
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

func (h *Handler) GetNotificationDelivery(c *gin.Context) {
	deliveryID, err := uuid.Parse(c.Param("deliveryID"))
	if err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Valid notification delivery ID is required",
		)
		return
	}

	delivery, err := h.NotificationsSvc.GetNotificationDelivery(c.Request.Context(), deliveryID)
	if err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to fetch notification delivery",
		)
		return
	}

	response.OK(c, http.StatusOK, toNotificationDeliveryResponse(delivery))
}

func (h *Handler) RetryNotificationDelivery(c *gin.Context) {
	deliveryID, err := uuid.Parse(c.Param("deliveryID"))
	if err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Valid notification delivery ID is required",
		)
		return
	}

	if err := h.NotificationsSvc.RetryNotificationDelivery(
		c.Request.Context(),
		deliveryID,
	); err != nil {
		status := http.StatusInternalServerError
		code := "INTERNAL_ERROR"
		message := "Failed to retry notification delivery"

		if err.Error() == "sent notification deliveries cannot be retried" {
			status = http.StatusConflict
			code = "INVALID_DELIVERY_STATE"
			message = "Sent notification deliveries cannot be retried"
		}

		response.Fail(c, status, code, message)
		return
	}

	response.OK(
		c,
		http.StatusOK,
		DeliveryRetryResponse{
			ID:     deliveryID.String(),
			Status: "RETRY",
		},
	)
}

func (h *Handler) CancelNotificationDelivery(c *gin.Context) {
	deliveryID, err := uuid.Parse(c.Param("deliveryID"))
	if err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Valid notification delivery ID is required",
		)
		return
	}

	if err := h.NotificationsSvc.CancelNotificationDelivery(c.Request.Context(), deliveryID); err != nil {
		status := http.StatusInternalServerError
		code := "INTERNAL_ERROR"
		message := "Failed to cancel notification delivery"

		if err.Error() == "sent notification deliveries cannot be cancelled" {
			status = http.StatusConflict
			code = "INVALID_DELIVERY_STATE"
			message = "Sent notification deliveries cannot be cancelled"
		}

		response.Fail(c, status, code, message)
		return
	}

	response.OK(c, http.StatusOK, DeliveryActionResponse{
		ID:     deliveryID.String(),
		Status: "CANCELLED",
	})
}

func (h *Handler) TestSMS(c *gin.Context) {
	var input TestSMSRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"SMS recipient and message are required",
		)
		return
	}

	notificationID, err := h.NotificationsSvc.QueueTestSMS(
		c.Request.Context(),
		input.To,
		input.Message,
	)
	if err != nil {
		status := http.StatusInternalServerError
		code := "INTERNAL_ERROR"
		message := "Failed to queue test SMS"
		if err.Error() == "sms delivery is disabled" {
			status = http.StatusServiceUnavailable
			code = "SMS_DISABLED"
			message = "SMS delivery is disabled"
		}

		response.Fail(c, status, code, message)
		return
	}

	response.OK(c, http.StatusAccepted, TestSMSResponse{
		NotificationID: notificationID.String(),
		Status:         "PENDING",
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
