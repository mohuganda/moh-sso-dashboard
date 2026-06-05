package routes

import (
	"time"

	"github.com/gin-gonic/gin"

	announcementfeature "github.com/moh-sso-dashboard/internal/features/announcements"
	auditfeature "github.com/moh-sso-dashboard/internal/features/audit"
	clientfeature "github.com/moh-sso-dashboard/internal/features/clients"
	metricsfeature "github.com/moh-sso-dashboard/internal/features/metrics"
	notificationsfeature "github.com/moh-sso-dashboard/internal/features/notifications"
	userfeature "github.com/moh-sso-dashboard/internal/features/users"
	"github.com/moh-sso-dashboard/internal/middleware"
	"github.com/moh-sso-dashboard/internal/ratelimit"
)

func RegisterAdminRoutes(protected *gin.RouterGroup, deps Dependencies) {
	admin := protected.Group("/admin")
	admin.Use(middleware.RequireAdmin())
	admin.Use(ratelimit.Middleware(deps.Limiter, ratelimit.ByUser, deps.AdminRateLimitPerMin, time.Minute))

	registerAdminUserRoutes(admin, deps)
	registerAdminClientRoleRoutes(admin, deps)
	registerAdminMetricRoutes(admin, deps)
	registerAdminAuditRoutes(admin, deps)
	registerAdminNotificationRoutes(admin, deps)
	registerAdminAnnouncementRoutes(admin, deps)
}

func registerAdminUserRoutes(admin *gin.RouterGroup, deps Dependencies) {
	userfeature.RegisterAdminRoutes(admin, deps.Users)
}

func registerAdminClientRoleRoutes(admin *gin.RouterGroup, deps Dependencies) {
	clientfeature.RegisterAdminRoutes(admin, deps.Clients)
}

func registerAdminMetricRoutes(admin *gin.RouterGroup, deps Dependencies) {
	metricsfeature.RegisterAdminRoutes(admin, deps.Metrics)
}

func registerAdminAuditRoutes(admin *gin.RouterGroup, deps Dependencies) {
	auditfeature.RegisterAdminRoutes(
		admin,
		deps.Audit,
		deps.Limiter,
		deps.AuditLogRateLimitPerMin,
	)
}

func registerAdminNotificationRoutes(admin *gin.RouterGroup, deps Dependencies) {
	notificationsfeature.RegisterAdminRoutes(admin, deps.Notifications)
}

func registerAdminAnnouncementRoutes(admin *gin.RouterGroup, deps Dependencies) {
	announcementfeature.RegisterAdminRoutes(admin, deps.Announcements)
}
