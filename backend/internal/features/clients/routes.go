package clients

import (
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
)

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler) {
	clients := protected.Group("/clients")
	{
		clients.GET("", middleware.RequireAnyPermission(authz.PermissionSystemsRead, authz.PermissionClientsRead), handler.ListClients)
		clients.GET("/:id", middleware.RequirePermission(authz.PermissionClientsRead), handler.GetClient)
		clients.POST("", middleware.RequirePermission(authz.PermissionClientsWrite), handler.CreateClient)
		clients.PUT("/:id", middleware.RequirePermission(authz.PermissionClientsWrite), handler.UpdateClient)
		clients.PATCH("/:id/toggle", middleware.RequirePermission(authz.PermissionClientsWrite), handler.ToggleClientEnabled)
		clients.DELETE("/:id", middleware.RequirePermission(authz.PermissionClientsWrite), handler.DeleteClient)
		clients.GET("/:id/roles", middleware.RequirePermission(authz.PermissionClientsRead), handler.ListClientRoles)
	}
}

func RegisterAdminRoutes(admin *gin.RouterGroup, handler *Handler) {
	admin.POST("/clients/:id/roles", middleware.RequirePermission(authz.PermissionClientsRolesWrite), handler.CreateClientRole)
	admin.DELETE("/clients/:id/roles/:role", middleware.RequirePermission(authz.PermissionClientsRolesWrite), handler.DeleteClientRole)
}
