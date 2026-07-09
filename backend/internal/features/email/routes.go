package email

import (
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
	"github.com/moh-sso-dashboard/internal/ratelimit"
)

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler, limiter *ratelimit.Limiter) {
	email := protected.Group("/emails")
	sendLimit := ratelimit.MiddlewareForPolicy(limiter, ratelimit.EmailSendPolicy())
	manageLimit := ratelimit.MiddlewareForPolicy(limiter, ratelimit.EmailManagePolicy())
	{
		email.POST("/send", middleware.RequirePermission(authz.PermissionEmailSend), sendLimit, handler.Send)
		email.POST("/queue", middleware.RequirePermission(authz.PermissionEmailSend), sendLimit, handler.Queue)
		email.POST("/recipient-preview", middleware.RequirePermission(authz.PermissionEmailSend), sendLimit, handler.PreviewRecipients)
		email.GET("", middleware.RequirePermission(authz.PermissionEmailRead), handler.List)
		email.GET("/status/:status", middleware.RequirePermission(authz.PermissionEmailRead), handler.ListByStatus)
		email.GET("/:id", middleware.RequirePermission(authz.PermissionEmailRead), handler.GetByID)
		email.POST("/:id/retry", middleware.RequirePermission(authz.PermissionEmailManage), manageLimit, handler.Retry)
		email.DELETE("/:id", middleware.RequirePermission(authz.PermissionEmailManage), manageLimit, handler.Delete)
	}
}
