package routes

import (
	"github.com/moh-sso-dashboard/internal/api/handler"
	announcementfeature "github.com/moh-sso-dashboard/internal/features/announcements"
	auditfeature "github.com/moh-sso-dashboard/internal/features/audit"
	clientfeature "github.com/moh-sso-dashboard/internal/features/clients"
	emailfeature "github.com/moh-sso-dashboard/internal/features/email"
	sessionfeature "github.com/moh-sso-dashboard/internal/features/sessions"
	storagelocationfeature "github.com/moh-sso-dashboard/internal/features/storage_locations"
	userfeature "github.com/moh-sso-dashboard/internal/features/users"
	"github.com/moh-sso-dashboard/internal/ratelimit"
)

type Dependencies struct {
	Limiter *ratelimit.Limiter

	Auth                    *handler.AuthHandler
	Clients                 *clientfeature.Handler
	Users                   *userfeature.Handler
	Metrics                 *handler.MetricsHandler
	Audit                   *auditfeature.Handler
	Notifications           *handler.NotificationsHandler
	Documents               *handler.DocumentHandler
	DocumentTemplates       *handler.DocumentTemplateHandler
	DocumentTemplateSheets  *handler.DocumentTemplateSheetHandler
	DocumentTemplateColumns *handler.DocumentTemplateColumnHandler
	StorageLocations        *storagelocationfeature.Handler
	Sessions                *sessionfeature.Handler
	DataQuality             *handler.DataQualityHandler
	Announcements           *announcementfeature.Handler
	AdminUnits              *handler.AdminUnitsHandler
	Visualiser              *handler.VisualiserHandler
	Surveillance            *handler.SurveillanceHandler
	GeoJSON                 *handler.GeoJSONHandler
	Email                   *emailfeature.Handler

	AuthenticatedRateLimitPerMin  int
	AdminRateLimitPerMin          int
	AuditLogRateLimitPerMin       int
	AuthLoginRateLimitPerMin      int
	AuthCallbackRateLimitPerMin   int
	AuthSessionRateLimitPerMinute int
}
