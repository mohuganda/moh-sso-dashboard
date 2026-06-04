package routes

import (
	"github.com/moh-sso-dashboard/internal/api/handler"
	"github.com/moh-sso-dashboard/internal/ratelimit"
)

type Dependencies struct {
	Limiter *ratelimit.Limiter

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

	AuthenticatedRateLimitPerMin  int
	AdminRateLimitPerMin          int
	AuditLogRateLimitPerMin       int
	AuthLoginRateLimitPerMin      int
	AuthCallbackRateLimitPerMin   int
	AuthSessionRateLimitPerMinute int
}
