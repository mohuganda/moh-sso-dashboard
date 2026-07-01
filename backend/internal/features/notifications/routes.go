package notifications

import (
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
	"github.com/moh-sso-dashboard/internal/ratelimit"
)

func RegisterAdminRoutes(admin *gin.RouterGroup, handler *Handler, limiter *ratelimit.Limiter) {
	notifications := admin.Group("/notifications")
	writeLimit := ratelimit.MiddlewareForPolicy(limiter, ratelimit.NotificationsWritePolicy())
	{
		notifications.POST("", middleware.RequirePermission(authz.PermissionNotificationsWrite), writeLimit, handler.Notify)
		notifications.GET("", middleware.RequirePermission(authz.PermissionNotificationsRead), handler.ListNotifications)
		notifications.GET("/count", middleware.RequirePermission(authz.PermissionNotificationsRead), handler.CountNotifications)
		notifications.GET("/count/unread", middleware.RequirePermission(authz.PermissionNotificationsRead), handler.CountUnreadNotificationsCount)
		notifications.DELETE("/cleanup", middleware.RequirePermission(authz.PermissionNotificationsWrite), writeLimit, handler.DeleteOldNotifications)
		notifications.GET("/deliveries", middleware.RequirePermission(authz.PermissionNotificationsRead), handler.ListAllNotificationDeliveries)
		notifications.GET("/deliveries/:deliveryID", middleware.RequirePermission(authz.PermissionNotificationsRead), handler.GetNotificationDelivery)
		notifications.POST("/deliveries/:deliveryID/retry", middleware.RequirePermission(authz.PermissionNotificationsWrite), writeLimit, handler.RetryNotificationDelivery)
		notifications.POST("/deliveries/:deliveryID/cancel", middleware.RequirePermission(authz.PermissionNotificationsWrite), writeLimit, handler.CancelNotificationDelivery)
		notifications.POST("/test-sms", middleware.RequirePermission(authz.PermissionNotificationsWrite), writeLimit, handler.TestSMS)
		notifications.GET("/:id/deliveries", middleware.RequirePermission(authz.PermissionNotificationsRead), handler.ListNotificationDeliveries)
		notifications.GET("/:id", middleware.RequirePermission(authz.PermissionNotificationsRead), handler.GetNotificationByID)
		notifications.PATCH("/:id/read", middleware.RequirePermission(authz.PermissionNotificationsRead), handler.MarkNotificationAsRead)
		notifications.DELETE("/:id", middleware.RequirePermission(authz.PermissionNotificationsWrite), writeLimit, handler.DeleteNotification)
	}
}
