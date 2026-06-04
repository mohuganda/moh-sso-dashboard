package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/api/handler"
	"github.com/moh-sso-dashboard/internal/api/routes"
	"github.com/moh-sso-dashboard/internal/keycloak"
	"github.com/moh-sso-dashboard/internal/middleware"
	"github.com/moh-sso-dashboard/internal/ratelimit"
	service "github.com/moh-sso-dashboard/internal/service"
)

type RouterDependencies struct {
	KeycloakClient *keycloak.Client
	Limiter        *ratelimit.Limiter
	AuditService   *service.AuditService
	Handlers       HandlerSet
	RateLimits     RateLimits
}

type HandlerSet struct {
	Auth                    *handler.AuthHandler
	Clients                 *handler.ClientHandler
	Users                   *handler.UserHandler
	Metrics                 *handler.MetricsHandler
	Audit                   *handler.AuditHandler
	Notifications           *handler.NotificationsHandler
	Documents               *handler.DocumentHandler
	DocumentTemplates       *handler.DocumentTemplateHandler
	DocumentTemplateSheets  *handler.DocumentTemplateSheetHandler
	DocumentTemplateColumns *handler.DocumentTemplateColumnHandler
	StorageLocations        *handler.StorageLocationHandler
	Sessions                *handler.SessionHandler
	DataQuality             *handler.DataQualityHandler
	Announcements           *handler.AnnouncementHandler
	AdminUnits              *handler.AdminUnitsHandler
	Visualiser              *handler.VisualiserHandler
	Surveillance            *handler.SurveillanceHandler
	GeoJSON                 *handler.GeoJSONHandler
	Email                   *handler.EmailHandler
}

type RateLimits struct {
	AuthenticatedPerMinute int
	AdminPerMinute         int
	AuditLogPerMinute      int
	AuthLoginPerMinute     int
	AuthCallbackPerMinute  int
	AuthSessionPerMinute   int
}

func SetupRouter(deps RouterDependencies) *gin.Engine {
	rateLimits := deps.RateLimits.withDefaults()

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(cors.New(corsConfig()))

	routeDeps := routes.Dependencies{
		Limiter:                       deps.Limiter,
		Auth:                          deps.Handlers.Auth,
		Clients:                       deps.Handlers.Clients,
		Users:                         deps.Handlers.Users,
		Metrics:                       deps.Handlers.Metrics,
		Audit:                         deps.Handlers.Audit,
		Notifications:                 deps.Handlers.Notifications,
		Documents:                     deps.Handlers.Documents,
		DocumentTemplates:             deps.Handlers.DocumentTemplates,
		DocumentTemplateSheets:        deps.Handlers.DocumentTemplateSheets,
		DocumentTemplateColumns:       deps.Handlers.DocumentTemplateColumns,
		StorageLocations:              deps.Handlers.StorageLocations,
		Sessions:                      deps.Handlers.Sessions,
		DataQuality:                   deps.Handlers.DataQuality,
		Announcements:                 deps.Handlers.Announcements,
		AdminUnits:                    deps.Handlers.AdminUnits,
		Visualiser:                    deps.Handlers.Visualiser,
		Surveillance:                  deps.Handlers.Surveillance,
		GeoJSON:                       deps.Handlers.GeoJSON,
		Email:                         deps.Handlers.Email,
		AuthenticatedRateLimitPerMin:  rateLimits.AuthenticatedPerMinute,
		AdminRateLimitPerMin:          rateLimits.AdminPerMinute,
		AuditLogRateLimitPerMin:       rateLimits.AuditLogPerMinute,
		AuthLoginRateLimitPerMin:      rateLimits.AuthLoginPerMinute,
		AuthCallbackRateLimitPerMin:   rateLimits.AuthCallbackPerMinute,
		AuthSessionRateLimitPerMinute: rateLimits.AuthSessionPerMinute,
	}

	api := r.Group("/api/v1")
	routes.RegisterAuthRoutes(api, routeDeps)
	routes.RegisterPublicAnnouncementRoutes(api, routeDeps)

	protected := api.Group("")
	protected.Use(middleware.ExtractAuthContext(deps.KeycloakClient))
	protected.Use(middleware.RequireAuth())
	protected.Use(middleware.AuditMiddleware(deps.AuditService))
	protected.Use(ratelimit.Middleware(deps.Limiter, ratelimit.ByUser, routeDeps.AuthenticatedRateLimitPerMin, time.Minute))

	routes.RegisterProtectedRoutes(protected, routeDeps)
	routes.RegisterAdminRoutes(protected, routeDeps)

	return r
}

func RegisterHealthRoutes(r *gin.Engine, healthHandler *handler.HealthHandler) {
	r.GET("/health/live", healthHandler.HandleLive)
	r.GET("/health/ready", healthHandler.HandleReady)
	r.GET("/health", healthHandler.HandleHealth)
}

func (limits RateLimits) withDefaults() RateLimits {
	if limits.AuthenticatedPerMinute == 0 {
		limits.AuthenticatedPerMinute = 120
	}
	if limits.AdminPerMinute == 0 {
		limits.AdminPerMinute = 60
	}
	if limits.AuditLogPerMinute == 0 {
		limits.AuditLogPerMinute = 30
	}
	if limits.AuthLoginPerMinute == 0 {
		limits.AuthLoginPerMinute = 20
	}
	if limits.AuthCallbackPerMinute == 0 {
		limits.AuthCallbackPerMinute = 30
	}
	if limits.AuthSessionPerMinute == 0 {
		limits.AuthSessionPerMinute = 60
	}

	return limits
}

func corsConfig() cors.Config {
	return cors.Config{
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
	}
}
