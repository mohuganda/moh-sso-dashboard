package audit

import (
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
	"github.com/moh-sso-dashboard/internal/ratelimit"
)

func RegisterAdminRoutes(
	admin *gin.RouterGroup,
	handler *Handler,
	limiter *ratelimit.Limiter,
) {
	audit := admin.Group("/audit-logs")
	audit.Use(middleware.RequirePermission(authz.PermissionAuditRead))
	{
		audit.GET("", handler.ListAuditLogs)
		audit.GET("/actions", handler.ListAuditActions)
		audit.GET("/metrics/overview", handler.AuditMetricsOverview)
		audit.GET("/metrics/failed-logins-by-day", handler.FailedLoginsByDay)
		audit.GET("/metrics/top-failure-ips", handler.TopFailureIPs)
		audit.GET("/export", ratelimit.MiddlewareForPolicy(limiter, ratelimit.AuditExportPolicy()), handler.ExportAuditLogs)
		audit.GET("/:id", handler.GetAuditLog)
	}
}
