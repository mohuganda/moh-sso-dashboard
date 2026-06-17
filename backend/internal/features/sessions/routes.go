package sessions

import (
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
)

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler) {
	sessions := protected.Group("/sessions")
	{
		sessions.GET("", middleware.RequirePermission(authz.PermissionPortalAccess), handler.GetUserSessions)
		sessions.DELETE("/:id", middleware.RequirePermission(authz.PermissionPortalAccess), handler.LogoutSession)
	}
}
