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
	documentTemplateHandler *handler.DocumentTemplateHandler,
	documentTemplateSheetHandler *handler.DocumentTemplateSheetHandler,
	documentTemplateColumnHandler *handler.DocumentTemplateColumnHandler,
	storageLocationHandler *handler.StorageLocationHandler,
	sessionHandler *handler.SessionHandler,
	dataQualityHandler *handler.DataQualityHandler,
	announcementHandler *handler.AnnouncementHandler,
	adminunitsHandler *handler.AdminUnitsHandler,
	visualiserHandler *handler.VisualiserHandler,
	surveillanceHandler *handler.SurveillanceHandler,
	geojsonHandler *handler.GeoJSONHandler,
	emailHandler *handler.EmailHandler,
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
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
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

		auth.GET(
			"/me",
			ratelimit.Middleware(
				limiter,
				ratelimit.ByIP,
				60,
				time.Minute,
			),
			authHandler.HandleAuthGetMe,
		)

		auth.POST(
			"/refresh",
			ratelimit.Middleware(
				limiter,
				ratelimit.ByIP,
				60,
				time.Minute,
			),
			authHandler.HandleAuthRefreshToken,
		)

		auth.GET("/logout", authHandler.HandleAuthLogout)
	}

	// -------------------------------------
	// User / published announcements
	// -------------------------------------
	userAnnouncements := api.Group("/announcements")
	{
		userAnnouncements.GET("/public", announcementHandler.ListPublicAnnouncements)
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

	protected.Use(
		ratelimit.Middleware(
			limiter,
			ratelimit.ByUser,
			120,
			time.Minute,
		),
	)

	{
		// ------------------
		// Email
		// ------------------
		email := protected.Group("/emails")
		{
			email.POST("/send", emailHandler.Send)
			email.POST("/queue", emailHandler.Queue)
			email.GET("", emailHandler.List)
			email.GET("/status/:status", emailHandler.ListByStatus)
			email.GET("/:id", emailHandler.GetByID)
			email.POST("/:id/retry", emailHandler.Retry)
			email.DELETE("/:id", emailHandler.Delete)
		}

		// ------------------
		// GeoJSON
		// ------------------
		geojson := protected.Group("/geojson")
		{
			geojson.GET("/:name", geojsonHandler.GetGeoJSON)
		}

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
		// General authenticated user routes.
		// Keep sensitive actions under /admin/users below.
		// ------------------
		users := protected.Group("/users")
		{
			users.GET("", userHandler.ListUsers)
			users.GET("/:id", userHandler.GetUser)
		}

		// -----------------------
		// Document Management
		// -----------------------
		documents := protected.Group("/documents")
		{
			documents.GET("", documentHandler.ListDocuments)
			documents.POST("", documentHandler.CreateDocument)

			documents.GET("/:id", documentHandler.GetDocument)
			documents.PUT("/:id", documentHandler.EditDocument)
			documents.DELETE("/:id", documentHandler.DeleteDocument)

			documents.GET("/files/:id/view", documentHandler.ViewDocument)
			documents.GET("/files/:id/download", documentHandler.DownloadDocument)

			documents.GET("/:id/processes", documentHandler.ListDocumentProcesses)
			documents.POST("/:id/reprocess", documentHandler.ReprocessDocument)
		}

		documentTemplates := protected.Group("/document-templates")
		{
			documentTemplates.GET("", documentTemplateHandler.ListTemplates)
			documentTemplates.POST("", documentTemplateHandler.CreateTemplate)

			documentTemplates.POST("/structure", documentTemplateHandler.CreateTemplateWithStructure)

			documentTemplates.GET("/code/:code/structure", documentTemplateHandler.GetTemplateStructure)

			documentTemplates.GET("/:id", documentTemplateHandler.GetTemplate)
			documentTemplates.PUT("/:id", documentTemplateHandler.UpdateTemplate)
			documentTemplates.DELETE("/:id", documentTemplateHandler.DeleteTemplate)

			documentTemplates.POST("/:id/publish", documentTemplateHandler.PublishTemplate)
			documentTemplates.POST("/:id/archive", documentTemplateHandler.ArchiveTemplate)

			documentTemplates.GET("/:id/structure", documentTemplateHandler.GetTemplateStructure)

			documentTemplates.GET("/:id/sheets", documentTemplateHandler.ListSheets)
			documentTemplates.GET("/:id/sheets/:sheetId/columns", documentTemplateHandler.ListColumns)
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
		// ------------------------------
		sessions := protected.Group("/sessions")
		{
			sessions.GET("", sessionHandler.GetUserSessions)
			sessions.DELETE("/:id", sessionHandler.LogoutSession)
		}

		// ------------------------------
		// Data Quality Issues
		// ------------------------------
		issues := protected.Group("/issues")
		{
			issues.POST("", dataQualityHandler.CreateIssue)
			issues.GET("", dataQualityHandler.ListIssues)
			issues.PUT("/:issueCode", dataQualityHandler.UpdateIssue)
			issues.POST("/:issueCode/resolveIssue", dataQualityHandler.ResolveIssue)
			issues.GET("/:issueCode/transactions", dataQualityHandler.ListIssueResolutionTransactions)
		}

		// ----------------------------------
		// Visualiser
		// ----------------------------------
		visualiser := protected.Group("/visualizer")
		{
			visualiser.GET("/adminunits/orgunits", adminunitsHandler.GetOrgUnits)
			visualiser.GET("/adminunits/facilities", adminunitsHandler.GetFacilities)
			visualiser.GET("/adminunits/district", adminunitsHandler.GetDistricts)
			visualiser.POST("/adminunits/subcounties", adminunitsHandler.GetSubCounties)
			visualiser.POST("/adminunits/localgovt", adminunitsHandler.GetLocalGovt)
			visualiser.POST("/adminunits/districts", adminunitsHandler.GetDistrictsByRegion)
			visualiser.GET("/adminunits/region", adminunitsHandler.GetRegions)
			visualiser.GET("/adminunits/national", adminunitsHandler.GetNational)
			visualiser.GET("/adminunits/hierarchy", adminunitsHandler.GetHierarchy)

			visualiser.GET("/datasets", visualiserHandler.GetDatasets)
			visualiser.POST("/dataelements", visualiserHandler.GetDataElements)
			visualiser.POST("/datavalues", visualiserHandler.GetDataValues)
			visualiser.GET("/themes", visualiserHandler.GetThemes)
			visualiser.POST("/dataelements/theme", visualiserHandler.GetDataElementsByTheme)
			visualiser.GET("/hiv/summary", visualiserHandler.GetHIVSummary)
			visualiser.GET("/hiv/tested", visualiserHandler.GetHIVTested)
			visualiser.GET("/hiv/regimen", visualiserHandler.GetHIVRegimen)
		}

		// ------------------------------
		// Surveillance
		// ------------------------------
		surveillance := protected.Group("/surveillance")
		{
			surveillance.GET("/weeks", surveillanceHandler.ListEpiWeeksByYear)
			surveillance.GET("/alerts", surveillanceHandler.ListAlerts)
			surveillance.GET("/diseases", surveillanceHandler.ListDiseases)

			surveillance.GET("/regions", surveillanceHandler.ListRegions)

			surveillance.GET("/districts", surveillanceHandler.ListDistricts)
			surveillance.GET("/regions/:regionID/districts", surveillanceHandler.ListDistrictsByRegion)
			surveillance.POST("/districts", surveillanceHandler.UpsertDistrict)

			surveillance.GET("/districts/:districtID/subcounties", surveillanceHandler.ListSubcountiesByDistrict)
			surveillance.GET("/subcounties/:id", surveillanceHandler.GetSubcountyByID)
			surveillance.POST("/subcounties", surveillanceHandler.UpsertSubcounty)
			surveillance.DELETE("/subcounties/:id", surveillanceHandler.DeleteSubcounty)

			surveillance.GET("/facility-weekly-metrics/week/:epiWeekID", surveillanceHandler.ListFacilityWeeklyMetricsByWeek)
			surveillance.GET("/facility-weekly-metrics/facility/:facilityID", surveillanceHandler.ListFacilityWeeklyMetricsByFacility)
			surveillance.GET("/facility-weekly-metrics/facility/:facilityID/disease/:diseaseID/trend", surveillanceHandler.ListFacilityDiseaseMetricsTrend)
			surveillance.GET("/facility-weekly-metrics/facility/:facilityID/indicator/:indicatorID/trend", surveillanceHandler.ListFacilityIndicatorMetricsTrend)
			surveillance.GET("/facility-weekly-metrics/week/:epiWeekID/disease/:diseaseID", surveillanceHandler.ListFacilityDiseaseMetricsByWeekAndDisease)
			surveillance.GET("/facility-weekly-metrics/disease-trend", surveillanceHandler.ListDiseaseWeeklyTrendAggregated)

			surveillance.GET("/weekly-statuses/list", surveillanceHandler.ListWeeklyStatuses)
			surveillance.GET("/weekly-statuses/detailed", surveillanceHandler.ListWeeklyStatusesDetailed)
			surveillance.GET("/weekly-statuses/district/week/:epiWeekID", surveillanceHandler.ListDistrictWeeklyStatusesByWeek)
			surveillance.GET("/weekly-statuses/region/week/:epiWeekID", surveillanceHandler.ListRegionWeeklyStatusesByWeek)
			surveillance.GET("/weekly-statuses/national/week/:epiWeekID", surveillanceHandler.ListNationalWeeklyStatusesByWeek)

			surveillance.GET("/imports", surveillanceHandler.ListImportBatches)
			surveillance.POST("/imports", surveillanceHandler.CreateImportBatch)
			surveillance.GET("/imports/:batchID", surveillanceHandler.GetImportBatchByID)
			surveillance.PATCH("/imports/:batchID/status", surveillanceHandler.UpdateImportBatchStatus)
			surveillance.GET("/imports/:batchID/raw-rows", surveillanceHandler.ListImportRawRowsByBatch)
		}

		// --------------------------------------------------
		// Admin (ADMIN ONLY + STRICTER LIMITS)
		// --------------------------------------------------
		admin := protected.Group("/admin")
		admin.Use(middleware.RequireAdmin())

		admin.Use(
			ratelimit.Middleware(
				limiter,
				ratelimit.ByUser,
				60,
				time.Minute,
			),
		)

		{
			// --------------------------------------------------
			// Admin Users
			// --------------------------------------------------
			adminUsers := admin.Group("/users")
			{
				// CRUD
				adminUsers.GET("", userHandler.ListUsers)
				adminUsers.GET("/:id", userHandler.GetUser)
				adminUsers.POST("", userHandler.CreateUser)
				adminUsers.PUT("/:id", userHandler.UpdateUser)
				adminUsers.DELETE("/:id", userHandler.DeleteUser)

				// Enable / disable
				adminUsers.PATCH("/:id/enabled", userHandler.SetUserEnabled)

				// Backward-compatible alias for existing frontend calls.
				adminUsers.PATCH("/:id/toggle", userHandler.SetUserEnabled)

				// Keycloak email actions
				adminUsers.POST("/:id/onboarding-email", userHandler.SendUserOnboardingEmail)
				adminUsers.POST("/:id/verification-email", userHandler.SendUserVerificationEmail)
				adminUsers.POST("/:id/password-reset", userHandler.SendUserPasswordResetEmail)

				// Backward-compatible alias.
				adminUsers.POST("/:id/reset-password", userHandler.ResetUserPassword)

				// Client roles
				adminUsers.GET("/:id/client-roles", userHandler.GetUserClientRoles)
				adminUsers.GET("/:id/clients/:clientID/roles", userHandler.GetUserClientRolesForClient)
				adminUsers.POST("/:id/clients/:clientID/roles", userHandler.AddUserClientRoles)
				adminUsers.DELETE("/:id/clients/:clientID/roles", userHandler.RemoveUserClientRoles)

				// Backward-compatible bulk update route.
				adminUsers.PUT("/:id/client-roles", userHandler.UpdateUserClientRoles)
			}

			// --------------------------------------------------
			// Admin Client Roles
			// --------------------------------------------------
			admin.POST("/clients/:id/roles", clientHandler.CreateClientRole)
			admin.DELETE("/clients/:id/roles/:role", clientHandler.DeleteClientRole)

			// --------------------------------------------------
			// Metrics
			// --------------------------------------------------
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

			// --------------------------------------------------
			// Audit Logs
			// --------------------------------------------------
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

			// --------------------------------------------------
			// Notifications
			// --------------------------------------------------
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

			// --------------------------------------------------
			// Announcements
			// --------------------------------------------------
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
				announcements.POST("/:id/draft", announcementHandler.MoveAnnouncementToDraft)
				announcements.POST("/:id/schedule", announcementHandler.ScheduleAnnouncement)
				announcements.POST("/:id/archive", announcementHandler.ArchiveAnnouncement)

				announcements.PATCH("/:id/pin", announcementHandler.SetAnnouncementPinned)
				announcements.PATCH("/:id/priority", announcementHandler.SetAnnouncementPriority)
			}
		}
	}

	return r
}
