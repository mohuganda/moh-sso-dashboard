package health_context

import (
	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
)

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler) {
	me := protected.Group("/me/health-contexts")
	me.Use(middleware.RequirePermission(authz.PermissionHealthContextsRead))
	{
		me.GET("", handler.MyContexts)
		me.GET("/:contextId", handler.GetMyContext)
		me.GET("/:contextId/explanation", handler.ExplainMyAccess)
		me.POST("/select", handler.SelectMyContext)
	}
}

func RegisterAdminRoutes(admin *gin.RouterGroup, handler *Handler) {
	contexts := admin.Group("/health-contexts")
	{
		contexts.GET("", middleware.RequirePermission(authz.PermissionHealthContextsRead), handler.ListNodes)
		contexts.GET("/drift", middleware.RequirePermission(authz.PermissionHealthContextsSync), handler.GetDrift)
		contexts.POST("/sync/preview", middleware.RequirePermission(authz.PermissionHealthContextsSync), handler.PreviewSync)
		contexts.POST("/sync/apply", middleware.RequirePermission(authz.PermissionHealthContextsSync), handler.ApplySync)
		contexts.POST("", middleware.RequirePermission(authz.PermissionHealthContextsWrite), handler.CreateNode)
		contexts.GET("/:contextId", middleware.RequirePermission(authz.PermissionHealthContextsRead), handler.GetNode)
		contexts.PUT("/:contextId", middleware.RequirePermission(authz.PermissionHealthContextsWrite), handler.UpdateNode)
		contexts.DELETE("/:contextId", middleware.RequirePermission(authz.PermissionHealthContextsWrite), handler.DeleteNode)
		contexts.GET("/:contextId/children", middleware.RequirePermission(authz.PermissionHealthContextsRead), handler.ListNodes)
		contexts.POST("/:contextId/children", middleware.RequirePermission(authz.PermissionHealthContextsWrite), handler.CreateChild)
		contexts.GET("/:contextId/descendants", middleware.RequirePermission(authz.PermissionHealthContextsRead), handler.ListDescendants)
		contexts.GET("/:contextId/aliases", middleware.RequirePermission(authz.PermissionHealthContextsRead), handler.ListAliases)
		contexts.PUT("/:contextId/aliases", middleware.RequirePermission(authz.PermissionHealthContextsSync), handler.UpsertAlias)
		contexts.DELETE("/:contextId/aliases/:aliasId", middleware.RequirePermission(authz.PermissionHealthContextsSync), handler.DeleteAlias)
	}

	users := admin.Group("/users/:id/health-contexts")
	users.Use(middleware.RequirePermission(authz.PermissionHealthContextsAssign))
	{
		users.GET("", handler.ListUserAssignments)
		users.PUT("", handler.ReplaceUserAssignments)
	}

	groups := admin.Group("/rbac/groups/:groupId/health-contexts")
	groups.Use(middleware.RequirePermission(authz.PermissionHealthContextsAssign))
	{
		groups.GET("", handler.ListGroupAssignments)
		groups.PUT("", handler.ReplaceGroupAssignments)
	}
}
