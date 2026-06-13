package data_quality

import (
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
)

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler) {
	issues := protected.Group("/issues")
	{
		issues.POST("", middleware.RequirePermission(authz.PermissionDataQualityWrite), handler.CreateIssue)
		issues.GET("", middleware.RequirePermission(authz.PermissionDataQualityRead), handler.ListIssues)
		issues.PUT("/:issueCode", middleware.RequirePermission(authz.PermissionDataQualityWrite), handler.UpdateIssue)
		issues.POST("/:issueCode/resolveIssue", middleware.RequirePermission(authz.PermissionDataQualityResolve), handler.ResolveIssue)
		issues.GET("/:issueCode/transactions", middleware.RequirePermission(authz.PermissionDataQualityRead), handler.ListIssueResolutionTransactions)
	}
}
