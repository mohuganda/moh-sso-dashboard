package rbac

import (
	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
	"github.com/moh-sso-dashboard/internal/ratelimit"
)

func RegisterAdminRoutes(admin *gin.RouterGroup, handler *Handler, limiter *ratelimit.Limiter) {
	rbac := admin.Group("/rbac")
	sensitiveWriteLimit := ratelimit.MiddlewareForPolicy(limiter, ratelimit.RBACSensitiveWritePolicy())
	bulkWriteLimit := ratelimit.MiddlewareForPolicy(limiter, ratelimit.RBACBulkWritePolicy())
	{
		rbac.GET("/systems", middleware.RequirePermission(authz.PermissionRBACRead), handler.ListSystems)
		rbac.GET("/systems/:clientId", middleware.RequirePermission(authz.PermissionRBACRead), handler.GetSystem)
		rbac.PUT("/systems/:clientId", middleware.RequirePermission(authz.PermissionRBACWrite), sensitiveWriteLimit, handler.UpdateSystem)
		rbac.GET("/permissions", middleware.RequirePermission(authz.PermissionRBACRead), handler.ListPermissions)
		rbac.PUT("/permissions/:permissionKey", middleware.RequirePermission(authz.PermissionRBACPermissionsWrite), sensitiveWriteLimit, handler.UpdatePermissionMetadata)
		rbac.GET("/drift", middleware.RequirePermission(authz.PermissionRBACRead), handler.GetDrift)
		rbac.POST("/drift/realm-export", middleware.RequirePermission(authz.PermissionRBACRead), handler.DriftFromRealmExport)
		rbac.POST("/sync/preview", middleware.RequirePermission(authz.PermissionRBACRead), handler.PreviewSync)
		rbac.POST("/sync/apply", middleware.RequirePermission(authz.PermissionRBACWrite), sensitiveWriteLimit, handler.ApplySync)
		rbac.GET("/effective-access/users/:userId", middleware.RequirePermission(authz.PermissionRBACRead), handler.GetEffectiveAccess)
		rbac.POST("/changes/preview", middleware.RequirePermission(authz.PermissionRBACRead), handler.PreviewChange)
		rbac.GET("/export", middleware.RequirePermission(authz.PermissionRBACRead), handler.ExportSeed)
		rbac.POST("/import/preview", middleware.RequirePermission(authz.PermissionRBACRead), handler.PreviewImport)
		rbac.POST("/import/apply", middleware.RequirePermission(authz.PermissionRBACWrite), sensitiveWriteLimit, handler.ApplyImport)
		rbac.GET("/audit", middleware.RequireAnyPermission(authz.PermissionRBACRead, authz.PermissionAuditRead), handler.ListAuditEvents)
		rbac.GET("/role-templates", middleware.RequirePermission(authz.PermissionRBACRead), handler.ListRoleTemplates)
		rbac.POST("/bulk/assign-permission", middleware.RequirePermission(authz.PermissionRBACPermissionsWrite), bulkWriteLimit, handler.BulkAssignPermission)
		rbac.POST("/bulk/remove-permission", middleware.RequirePermission(authz.PermissionRBACPermissionsWrite), bulkWriteLimit, handler.BulkRemovePermission)
		rbac.GET("/access-requests", middleware.RequirePermission(authz.PermissionRBACRolesWrite), handler.ListAccessRequests)
		rbac.POST("/access-requests/:id/:decision", middleware.RequirePermission(authz.PermissionRBACRolesWrite), sensitiveWriteLimit, handler.DecideAccessRequest)
		rbac.GET("/change-requests", middleware.RequirePermission(authz.PermissionRBACRead), handler.ListChangeRequests)
		rbac.POST("/change-requests", middleware.RequirePermission(authz.PermissionRBACWrite), sensitiveWriteLimit, handler.CreateChangeRequest)
		rbac.POST("/change-requests/:id/:decision", middleware.RequirePermission(authz.PermissionRBACWrite), sensitiveWriteLimit, handler.DecideChangeRequest)
		rbac.POST("/simulate", middleware.RequirePermission(authz.PermissionRBACRead), handler.Simulate)

		rbac.GET("/systems/:clientId/roles", middleware.RequirePermission(authz.PermissionRBACRead), handler.ListSystemRoles)
		rbac.POST("/systems/:clientId/roles", middleware.RequirePermission(authz.PermissionRBACRolesWrite), sensitiveWriteLimit, handler.CreateSystemRole)
		rbac.POST("/systems/:clientId/roles/from-template", middleware.RequirePermission(authz.PermissionRBACRolesWrite), sensitiveWriteLimit, handler.CreateRoleFromTemplate)
		rbac.PUT("/roles/:roleId", middleware.RequirePermission(authz.PermissionRBACRolesWrite), sensitiveWriteLimit, handler.UpdateSystemRole)
		rbac.DELETE("/roles/:roleId", middleware.RequirePermission(authz.PermissionRBACRolesWrite), sensitiveWriteLimit, handler.DeleteSystemRole)
		rbac.GET("/roles/:roleId/usage", middleware.RequirePermission(authz.PermissionRBACRead), handler.GetRoleUsage)
		rbac.POST("/roles/:roleId/copy-permissions", middleware.RequirePermission(authz.PermissionRBACPermissionsWrite), bulkWriteLimit, handler.CopyPermissions)

		rbac.POST("/roles/:roleId/permissions", middleware.RequirePermission(authz.PermissionRBACPermissionsWrite), sensitiveWriteLimit, handler.AssignSystemRolePermission)
		rbac.DELETE("/roles/:roleId/permissions/:permissionKey", middleware.RequirePermission(authz.PermissionRBACPermissionsWrite), sensitiveWriteLimit, handler.RemoveSystemRolePermission)

		rbac.GET("/realm-roles", middleware.RequirePermission(authz.PermissionRBACRead), handler.ListRealmRolePermissions)
		rbac.GET("/realm-roles/:realmRole/usage", middleware.RequirePermission(authz.PermissionRBACRead), handler.GetRealmRoleUsage)
		rbac.POST("/realm-roles/:realmRole/permissions", middleware.RequirePermission(authz.PermissionRBACPermissionsWrite), sensitiveWriteLimit, handler.AssignRealmRolePermission)
		rbac.DELETE("/realm-roles/:realmRole/permissions/:permissionKey", middleware.RequirePermission(authz.PermissionRBACPermissionsWrite), sensitiveWriteLimit, handler.RemoveRealmRolePermission)

		rbac.POST("/systems/:clientId/access-roles", middleware.RequirePermission(authz.PermissionRBACRolesWrite), sensitiveWriteLimit, handler.AddSystemAccessRole)
		rbac.DELETE("/systems/:clientId/access-roles/:roleName", middleware.RequirePermission(authz.PermissionRBACRolesWrite), sensitiveWriteLimit, handler.RemoveSystemAccessRole)
	}
}

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler) {
	if handler == nil {
		return
	}
	protected.POST("/access-requests", middleware.RequirePermission(authz.PermissionPortalAccess), handler.CreateAccessRequest)
}
