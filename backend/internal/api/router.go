package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/api/handler"
	"github.com/moh-sso-dashboard/internal/keycloak"
	"github.com/moh-sso-dashboard/internal/middleware"
	"github.com/moh-sso-dashboard/internal/ratelimit"
	service "github.com/moh-sso-dashboard/internal/service"
)

func SetupRouter(
	keycloakClient *keycloak.Client,
	limiter *ratelimit.Limiter,
	authHandler *handler.AuthHandler,
	clientHandler *handler.ClientHandler,
	userHandler *handler.UserHandler,
	metricsHandler *handler.MetricsHandler,
	auditService *service.AuditService,
	auditHandler *handler.AuditHandler,
	notificationsHandler *handler.NotificationsHandler,
	documentHandler *handler.DocumentHandler,
	storageLocationHandler *handler.StorageLocationHandler,
	sessionHandler *handler.SessionHandler,
	announcementHandler *handler.AnnouncementHandler,
) *gin.Engine {

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// --------------------------------------------------
	// CORS
	// --------------------------------------------------
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
		},
		AllowMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
		},
		ExposeHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// --------------------------------------------------
	// API
	// --------------------------------------------------
	api := r.Group("/api/v1")

	// --------------------------------------------------
	// Auth (PUBLIC + RATE LIMITED)
	// --------------------------------------------------
	auth := api.Group("/auth")
	{
		auth.GET(
			"/login",
			ratelimit.Middleware(
				limiter,
				ratelimit.ByIP,
				20,
				time.Minute,
			),
			authHandler.HandleAuthLogin,
		)

		auth.GET(
			"/callback",
			ratelimit.Middleware(
				limiter,
				ratelimit.ByIP,
				30,
				time.Minute,
			),
			authHandler.HandleAuthCallback,
		)

		auth.POST(
			"/refresh",
			ratelimit.Middleware(
				limiter,
				ratelimit.ByIP,
				10,
				time.Minute,
			),
			authHandler.HandleAuthRefreshToken,
		)

		auth.GET("/logout", authHandler.HandleAuthLogout)
	}

	// -------------------------------------
	// announcements
	// -----------------------------------------
	// -------------------------------------
	// user / published announcements
	// -------------------------------------
	userAnnouncements := api.Group("/announcements")
	{
		userAnnouncements.GET("/me", announcementHandler.ListMyAnnouncements)
		userAnnouncements.GET("/active", announcementHandler.ListActivePublishedAnnouncements)
		userAnnouncements.GET("/user", announcementHandler.ListAnnouncementsForUser)
		userAnnouncements.GET("/role/:role_name", announcementHandler.ListAnnouncementsForRole)
		userAnnouncements.GET("/client/:client_id", announcementHandler.ListAnnouncementsForClient)
	}

	// --------------------------------------------------
	// Protected (AUTH REQUIRED)
	// --------------------------------------------------
	protected := api.Group("")
	protected.Use(middleware.ExtractAuthContext(keycloakClient))
	protected.Use(middleware.RequireAuth())
	protected.Use(middleware.AuditMiddleware(auditService))

	// moderate user-based rate limit
	protected.Use(
		ratelimit.Middleware(
			limiter,
			ratelimit.ByUser,
			120,
			time.Minute,
		),
	)

	{
		protected.GET("/auth/me", authHandler.HandleAuthGetMe)

		// ------------------
		// Clients
		// ------------------
		clients := protected.Group("/clients")
		{
			clients.GET("", clientHandler.ListClients)
			clients.GET("/:id", clientHandler.GetClient)
			clients.POST("", clientHandler.CreateClient)
			clients.PATCH("/:id/toggle", clientHandler.ToggleClientEnabled)
			clients.DELETE("/:id", clientHandler.DeleteClient)
			clients.GET("/:id/roles", clientHandler.ListClientRoles)
		}

		// ------------------
		// Users
		// ------------------
		users := protected.Group("/users")
		{
			users.GET("", userHandler.ListUsers)
			users.GET("/:id", userHandler.GetUser)
			users.POST("", userHandler.CreateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
			users.PATCH("/:id/toggle", userHandler.SetUserEnabled)
		}

		// -----------------------
		// Document Management
		// -----------------------
		documents := protected.Group("/documents")
		{
			documents.GET("", documentHandler.ListDocuments)

			documents.GET("/:id/download", documentHandler.DownloadDocument)
			documents.GET("/:id/processes", documentHandler.ListDocumentProcesses)
			documents.POST("/:id/reprocess", documentHandler.ReprocessDocument)
			documents.GET("/:id", documentHandler.GetDocument)

			documents.POST("", documentHandler.CreateDocument)
			documents.PUT("/:id", documentHandler.EditDocument)
			documents.DELETE("/:id", documentHandler.DeleteDocument)
		}

		// --------------------------
		// Storage Locations Management
		// --------------------------
		storageLocation := protected.Group("/storage-locations")
		{
			storageLocation.POST("", storageLocationHandler.Create)
			storageLocation.GET("", storageLocationHandler.ListActive)
			storageLocation.GET("/:id", storageLocationHandler.GetByID)
			storageLocation.PUT("/:id", storageLocationHandler.Update)
			storageLocation.DELETE("/:id", storageLocationHandler.Delete)
		}

		// ------------------------------
		// Session Management
		// --------------------------------
		sessions := protected.Group("/sessions")
		{
			sessions.GET("", sessionHandler.GetUserSessions)
			sessions.DELETE("/:id", sessionHandler.LogoutSession)
		}

		// --------------------------------------------------
		// Admin (ADMIN ONLY + STRICTER LIMITS)
		// --------------------------------------------------
		admin := protected.Group("/admin")
		admin.Use(middleware.RequireAdmin())

		// stricter admin rate limit
		admin.Use(
			ratelimit.Middleware(
				limiter,
				ratelimit.ByUser,
				60,
				time.Minute,
			),
		)

		{
			// -------- Users --------
			admin.GET("/users", userHandler.ListUsers)
			admin.GET("/users/:id", userHandler.GetUser)
			admin.POST("/users", userHandler.CreateUser)
			admin.DELETE("/users/:id", userHandler.DeleteUser)

			admin.GET("/users/:id/client-roles", userHandler.GetUserClientRoles)
			admin.PUT("/users/:id/client-roles", userHandler.UpdateUserClientRoles)
			admin.POST("/users/:id/reset-password", userHandler.ResetUserPassword)

			// -------- Client Roles --------
			admin.POST("/clients/:id/roles", clientHandler.CreateClientRole)
			admin.DELETE("/clients/:id/roles/:role", clientHandler.DeleteClientRole)

			// -------- Metrics --------
			metrics := admin.Group("/metrics")
			{
				metrics.GET("/overview", metricsHandler.Overview)
				metrics.GET("/system/count-users", metricsHandler.CountUsers)
				metrics.GET("/system/count-disabled-users", metricsHandler.CountDisabledUsers)
				metrics.GET("/system/active-today", metricsHandler.ActiveUsersToday)
				metrics.GET("/system/active-this-week", metricsHandler.ActiveUsersThisWeek)
				metrics.GET("/system/login-trend", metricsHandler.LoginTrend)
				metrics.GET("/system/login-trend-range", metricsHandler.LoginTrendByDay)

				metrics.GET("/security/failed-logins", metricsHandler.CountFailedLogins)
				metrics.GET("/security/failed-logins-range", metricsHandler.CountFailedLoginsInRange)
				metrics.GET("/security/suspicious-logins", metricsHandler.SuspiciousLogins)

				metrics.GET("/clients/count", metricsHandler.CountClients)
				metrics.GET("/clients/most-accessed", metricsHandler.MostAccessedClients)
				metrics.GET("/clients/login-count", metricsHandler.LoginCountForClient)
				metrics.GET("/clients/active-today", metricsHandler.ActiveUsersPerClientToday)

				metrics.GET("/users/new-range", metricsHandler.NewUsersInRange)
				metrics.GET("/users/new-trend", metricsHandler.NewUsersTrend)
				metrics.GET("/users/never-logged-in", metricsHandler.NeverLoggedInUsers)
				metrics.GET("/users/last-login/:userID", metricsHandler.LastLoginForUser)
				metrics.GET("/users/client-usage/:userID", metricsHandler.UserClientUsage)
			}

			// -------- Audit Logs (rate-limit tighter) --------
			audit := admin.Group("/audit-logs")
			audit.Use(
				ratelimit.Middleware(
					limiter,
					ratelimit.ByUser,
					30,
					time.Minute,
				),
			)
			{
				audit.GET("", auditHandler.ListAuditLogs)
				audit.GET("/actions", auditHandler.ListAuditActions)
				audit.GET("/:id", auditHandler.GetAuditLog)

				audit.GET("/metrics/overview", auditHandler.AuditMetricsOverview)
				audit.GET("/metrics/failed-logins-by-day", auditHandler.FailedLoginsByDay)
				audit.GET("/metrics/top-failure-ips", auditHandler.TopFailureIPs)

				audit.GET("/export", auditHandler.ExportAuditLogs)
			}

			// -------- Notifications --------
			notifications := admin.Group("/notifications")
			{
				notifications.POST("", notificationsHandler.Notify)
				notifications.GET("", notificationsHandler.ListNotifications)
				notifications.GET("/:id", notificationsHandler.GetNotificationByID)
				notifications.PATCH("/:id/read", notificationsHandler.MarkNotificationAsRead)
				notifications.DELETE("/:id", notificationsHandler.DeleteNotification)

				notifications.GET("/count", notificationsHandler.CountNotifications)
				notifications.GET("/count/unread", notificationsHandler.CountUnreadNotificationsCount)

				notifications.DELETE("/cleanup", notificationsHandler.DeleteOldNotifications)
			}

			//  -------- announements --------------------
			// -------------------------------------
			// admin announcements
			// -------------------------------------
			announcements := admin.Group("/announcements")
			{
				announcements.GET("", announcementHandler.ListAnnouncementsAdmin)
				announcements.GET("/stats", announcementHandler.GetAnnouncementStats)
				announcements.GET("/:id", announcementHandler.GetAnnouncementByID)

				announcements.POST("", announcementHandler.CreateAnnouncement)
				announcements.PUT("/:id", announcementHandler.UpdateAnnouncement)
				announcements.DELETE("/:id", announcementHandler.DeleteAnnouncement)
				announcements.POST("/:id/restore", announcementHandler.RestoreAnnouncement)

				announcements.POST("/:id/publish", announcementHandler.PublishAnnouncementNow)
				announcements.POST("/:id/schedule", announcementHandler.ScheduleAnnouncement)
				announcements.POST("/:id/archive", announcementHandler.ArchiveAnnouncement)

				announcements.PATCH("/:id/pin", announcementHandler.SetAnnouncementPinned)
				announcements.PATCH("/:id/priority", announcementHandler.SetAnnouncementPriority)
			}
		}
	}
	return r
}
