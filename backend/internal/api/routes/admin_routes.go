package routes

import (
	"time"

	"github.com/gin-gonic/gin"

	announcementfeature "github.com/moh-sso-dashboard/internal/features/announcements"
	auditfeature "github.com/moh-sso-dashboard/internal/features/audit"
	clientfeature "github.com/moh-sso-dashboard/internal/features/clients"
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
	metrics := admin.Group("/metrics")
	{
		metrics.GET("/overview", deps.Metrics.Overview)
		metrics.GET("/system/count-users", deps.Metrics.CountUsers)
		metrics.GET("/system/count-disabled-users", deps.Metrics.CountDisabledUsers)
		metrics.GET("/system/active-today", deps.Metrics.ActiveUsersToday)
		metrics.GET("/system/active-this-week", deps.Metrics.ActiveUsersThisWeek)
		metrics.GET("/system/login-trend", deps.Metrics.LoginTrend)
		metrics.GET("/system/login-trend-range", deps.Metrics.LoginTrendByDay)
		metrics.GET("/security/failed-logins", deps.Metrics.CountFailedLogins)
		metrics.GET("/security/failed-logins-range", deps.Metrics.CountFailedLoginsInRange)
		metrics.GET("/security/suspicious-logins", deps.Metrics.SuspiciousLogins)
		metrics.GET("/clients/count", deps.Metrics.CountClients)
		metrics.GET("/clients/most-accessed", deps.Metrics.MostAccessedClients)
		metrics.GET("/clients/login-count", deps.Metrics.LoginCountForClient)
		metrics.GET("/clients/active-today", deps.Metrics.ActiveUsersPerClientToday)
		metrics.GET("/users/new-range", deps.Metrics.NewUsersInRange)
		metrics.GET("/users/new-trend", deps.Metrics.NewUsersTrend)
		metrics.GET("/users/never-logged-in", deps.Metrics.NeverLoggedInUsers)
		metrics.GET("/users/last-login/:userID", deps.Metrics.LastLoginForUser)
		metrics.GET("/users/client-usage/:userID", deps.Metrics.UserClientUsage)
	}
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
	notifications := admin.Group("/notifications")
	{
		notifications.POST("", deps.Notifications.Notify)
		notifications.GET("", deps.Notifications.ListNotifications)
		notifications.GET("/:id", deps.Notifications.GetNotificationByID)
		notifications.PATCH("/:id/read", deps.Notifications.MarkNotificationAsRead)
		notifications.DELETE("/:id", deps.Notifications.DeleteNotification)
		notifications.GET("/count", deps.Notifications.CountNotifications)
		notifications.GET("/count/unread", deps.Notifications.CountUnreadNotificationsCount)
		notifications.DELETE("/cleanup", deps.Notifications.DeleteOldNotifications)
	}
}

func registerAdminAnnouncementRoutes(admin *gin.RouterGroup, deps Dependencies) {
	announcementfeature.RegisterAdminRoutes(admin, deps.Announcements)
}
