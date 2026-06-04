package routes

import (
	"time"

	"github.com/gin-gonic/gin"

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
	users := admin.Group("/users")
	{
		users.GET("", deps.Users.ListUsers)
		users.GET("/:id", deps.Users.GetUser)
		users.POST("", deps.Users.CreateUser)
		users.PUT("/:id", deps.Users.UpdateUser)
		users.DELETE("/:id", deps.Users.DeleteUser)
		users.PATCH("/:id/enabled", deps.Users.SetUserEnabled)
		users.PATCH("/:id/toggle", deps.Users.SetUserEnabled)
		users.POST("/:id/onboarding-email", deps.Users.SendUserOnboardingEmail)
		users.POST("/:id/verification-email", deps.Users.SendUserVerificationEmail)
		users.POST("/:id/password-reset", deps.Users.SendUserPasswordResetEmail)
		users.POST("/:id/reset-password", deps.Users.ResetUserPassword)
		users.GET("/:id/client-roles", deps.Users.GetUserClientRoles)
		users.GET("/:id/clients/:clientID/roles", deps.Users.GetUserClientRolesForClient)
		users.POST("/:id/clients/:clientID/roles", deps.Users.AddUserClientRoles)
		users.DELETE("/:id/clients/:clientID/roles", deps.Users.RemoveUserClientRoles)
		users.PUT("/:id/client-roles", deps.Users.UpdateUserClientRoles)
	}
}

func registerAdminClientRoleRoutes(admin *gin.RouterGroup, deps Dependencies) {
	admin.POST("/clients/:id/roles", deps.Clients.CreateClientRole)
	admin.DELETE("/clients/:id/roles/:role", deps.Clients.DeleteClientRole)
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
	audit := admin.Group("/audit-logs")
	audit.Use(ratelimit.Middleware(deps.Limiter, ratelimit.ByUser, deps.AuditLogRateLimitPerMin, time.Minute))
	{
		audit.GET("", deps.Audit.ListAuditLogs)
		audit.GET("/actions", deps.Audit.ListAuditActions)
		audit.GET("/:id", deps.Audit.GetAuditLog)
		audit.GET("/metrics/overview", deps.Audit.AuditMetricsOverview)
		audit.GET("/metrics/failed-logins-by-day", deps.Audit.FailedLoginsByDay)
		audit.GET("/metrics/top-failure-ips", deps.Audit.TopFailureIPs)
		audit.GET("/export", deps.Audit.ExportAuditLogs)
	}
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
	announcements := admin.Group("/announcements")
	{
		announcements.GET("", deps.Announcements.ListAnnouncementsAdmin)
		announcements.GET("/stats", deps.Announcements.GetAnnouncementStats)
		announcements.GET("/:id", deps.Announcements.GetAnnouncementByID)
		announcements.POST("", deps.Announcements.CreateAnnouncement)
		announcements.PUT("/:id", deps.Announcements.UpdateAnnouncement)
		announcements.DELETE("/:id", deps.Announcements.DeleteAnnouncement)
		announcements.POST("/:id/restore", deps.Announcements.RestoreAnnouncement)
		announcements.POST("/:id/publish", deps.Announcements.PublishAnnouncementNow)
		announcements.POST("/:id/draft", deps.Announcements.MoveAnnouncementToDraft)
		announcements.POST("/:id/schedule", deps.Announcements.ScheduleAnnouncement)
		announcements.POST("/:id/archive", deps.Announcements.ArchiveAnnouncement)
		announcements.PATCH("/:id/pin", deps.Announcements.SetAnnouncementPinned)
		announcements.PATCH("/:id/priority", deps.Announcements.SetAnnouncementPriority)
	}
}
