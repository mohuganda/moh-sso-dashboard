package email

import (
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
)

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler) {
	email := protected.Group("/emails")
	{
		email.POST("/send", middleware.RequirePermission(authz.PermissionEmailSend), handler.Send)
		email.POST("/queue", middleware.RequirePermission(authz.PermissionEmailSend), handler.Queue)
		email.GET("", middleware.RequirePermission(authz.PermissionEmailRead), handler.List)
		email.GET("/status/:status", middleware.RequirePermission(authz.PermissionEmailRead), handler.ListByStatus)
		email.GET("/:id", middleware.RequirePermission(authz.PermissionEmailRead), handler.GetByID)
		email.POST("/:id/retry", middleware.RequirePermission(authz.PermissionEmailManage), handler.Retry)
		email.DELETE("/:id", middleware.RequirePermission(authz.PermissionEmailManage), handler.Delete)
	}
}
