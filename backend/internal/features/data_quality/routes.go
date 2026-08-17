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
	registerValidationRuleRoutes(protected.Group("/data-quality"), handler)
	registerValidationRuleRoutes(protected.Group("/data-validation"), handler)

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

		issues.GET(
			"/summary-by-program",
			middleware.RequirePermission(authz.PermissionIssueTrackerRead),
			handler.ListIssueSummaryByProgram,
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

		issues.GET(
			"/keycloak-groups",
			middleware.RequirePermission(authz.PermissionIssueTrackerRead),
			handler.ListKeycloakGroups,
		)

		issues.GET(
			"/keycloak-groups/:groupId/members",
			middleware.RequirePermission(authz.PermissionIssueTrackerRead),
			handler.ListKeycloakGroupMembers,
		)

		issues.POST(
			"/assign",
			middleware.RequireAnyPermission(
				authz.PermissionIssueTrackerAssign,
				authz.PermissionIssueTrackerManage,
				authz.PermissionIssueTrackerWrite,
			),
			handler.AssignIssues,
		)

		issues.POST(
			"/:issueCode/assign",
			middleware.RequireAnyPermission(
				authz.PermissionIssueTrackerAssign,
				authz.PermissionIssueTrackerManage,
				authz.PermissionIssueTrackerWrite,
			),
			handler.AssignIssues,
		)
	}
}

func registerValidationRuleRoutes(group *gin.RouterGroup, handler *Handler) {
	rules := group.Group("/rules")
	{
		rules.GET(
			"",
			middleware.RequirePermission(authz.PermissionDataQualityRead),
			handler.ListValidationRules,
		)

		rules.POST(
			"/import",
			middleware.RequirePermission(authz.PermissionDataQualityWrite),
			handler.ImportValidationRules,
		)
	}
}
