package notifications

import "github.com/gin-gonic/gin"

func RegisterAdminRoutes(admin *gin.RouterGroup, handler *Handler) {
	notifications := admin.Group("/notifications")
	{
		notifications.POST("", handler.Notify)
		notifications.GET("", handler.ListNotifications)
		notifications.GET("/:id", handler.GetNotificationByID)
		notifications.PATCH("/:id/read", handler.MarkNotificationAsRead)
		notifications.DELETE("/:id", handler.DeleteNotification)
		notifications.GET("/count", handler.CountNotifications)
		notifications.GET("/count/unread", handler.CountUnreadNotificationsCount)
		notifications.DELETE("/cleanup", handler.DeleteOldNotifications)
	}
}
