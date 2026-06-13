package rbac

import (
	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
)

func RegisterAdminRoutes(admin *gin.RouterGroup, handler *Handler) {
	rbac := admin.Group("/rbac")
	{
		rbac.GET("/systems", middleware.RequirePermission(authz.PermissionRBACRead), handler.ListSystems)
		rbac.GET("/systems/:clientId", middleware.RequirePermission(authz.PermissionRBACRead), handler.GetSystem)
		rbac.PUT("/systems/:clientId", middleware.RequirePermission(authz.PermissionRBACWrite), handler.UpdateSystem)
		rbac.GET("/permissions", middleware.RequirePermission(authz.PermissionRBACRead), handler.ListPermissions)

		rbac.GET("/systems/:clientId/roles", middleware.RequirePermission(authz.PermissionRBACRead), handler.ListSystemRoles)
		rbac.POST("/systems/:clientId/roles", middleware.RequirePermission(authz.PermissionRBACRolesWrite), handler.CreateSystemRole)
		rbac.PUT("/roles/:roleId", middleware.RequirePermission(authz.PermissionRBACRolesWrite), handler.UpdateSystemRole)
		rbac.DELETE("/roles/:roleId", middleware.RequirePermission(authz.PermissionRBACRolesWrite), handler.DeleteSystemRole)

		rbac.POST("/roles/:roleId/permissions", middleware.RequirePermission(authz.PermissionRBACPermissionsWrite), handler.AssignSystemRolePermission)
		rbac.DELETE("/roles/:roleId/permissions/:permissionKey", middleware.RequirePermission(authz.PermissionRBACPermissionsWrite), handler.RemoveSystemRolePermission)

		rbac.GET("/realm-roles", middleware.RequirePermission(authz.PermissionRBACRead), handler.ListRealmRolePermissions)
		rbac.POST("/realm-roles/:realmRole/permissions", middleware.RequirePermission(authz.PermissionRBACPermissionsWrite), handler.AssignRealmRolePermission)
		rbac.DELETE("/realm-roles/:realmRole/permissions/:permissionKey", middleware.RequirePermission(authz.PermissionRBACPermissionsWrite), handler.RemoveRealmRolePermission)

		rbac.POST("/systems/:clientId/access-roles", middleware.RequirePermission(authz.PermissionRBACRolesWrite), handler.AddSystemAccessRole)
		rbac.DELETE("/systems/:clientId/access-roles/:roleName", middleware.RequirePermission(authz.PermissionRBACRolesWrite), handler.RemoveSystemAccessRole)
	}
}
