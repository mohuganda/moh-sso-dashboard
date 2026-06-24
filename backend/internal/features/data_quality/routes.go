package data_quality

import (
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
)

func RegisterProtectedRoutes(
	protected *gin.RouterGroup,
	handler *Handler,
) {
	issues := protected.Group("/issues")
	{
		issues.POST(
			"",
			middleware.RequirePermission(authz.PermissionIssueTrackerWrite),
			handler.CreateIssue,
		)

		issues.GET(
			"",
			middleware.RequirePermission(authz.PermissionIssueTrackerRead),
			handler.ListIssues,
		)

		issues.PUT(
			"/:issueCode",
			middleware.RequirePermission(authz.PermissionIssueTrackerWrite),
			handler.UpdateIssue,
		)

		issues.POST(
			"/:issueCode/resolveIssue",
			middleware.RequireAnyPermission(
				authz.PermissionIssueTrackerComment,
				authz.PermissionIssueTrackerManage,
				authz.PermissionIssueTrackerClose,
			),
			handler.ResolveIssue,
		)

		issues.GET(
			"/:issueCode/transactions",
			middleware.RequirePermission(authz.PermissionIssueTrackerRead),
			handler.ListIssueResolutionTransactions,
		)
	}
}
