package bootstrap

import (
	"context"

	router "github.com/moh-sso-dashboard/internal/api"
	"github.com/moh-sso-dashboard/internal/api/handler"
	"github.com/moh-sso-dashboard/internal/cache"
	"github.com/moh-sso-dashboard/internal/config"
	storepkg "github.com/moh-sso-dashboard/internal/db/sqlc"
	announcementfeature "github.com/moh-sso-dashboard/internal/features/announcements"
	auditfeature "github.com/moh-sso-dashboard/internal/features/audit"
	clientfeature "github.com/moh-sso-dashboard/internal/features/clients"
	emailfeature "github.com/moh-sso-dashboard/internal/features/email"
	sessionfeature "github.com/moh-sso-dashboard/internal/features/sessions"
	storagelocationfeature "github.com/moh-sso-dashboard/internal/features/storage_locations"
	userfeature "github.com/moh-sso-dashboard/internal/features/users"
	"github.com/moh-sso-dashboard/internal/keycloak"
	db "github.com/moh-sso-dashboard/internal/migrate"
	"github.com/moh-sso-dashboard/internal/storage"
	"github.com/redis/go-redis/v9"
)

type handlers struct {
	Router router.HandlerSet
	Health *handler.HealthHandler
}

type handlerDependencies struct {
	Config         *config.Config
	Store          storepkg.Store
	Cache          *cache.RedisCache
	Databases      databases
	Repositories   repositories
	Services       services
	FileStorage    storage.Storage
	StorageFactory *storage.StorageFactory
	AdminKeycloak  *keycloak.KeyAdminClient
	Redis          *redis.Client
}

func buildHandlers(deps handlerDependencies) handlers {
	authHandler := handler.NewAuthHandler(
		deps.Services.Auth,
		deps.Services.Audit,
		deps.Services.Notifications,
		deps.Config,
	)
	clientHandler := clientfeature.NewHandler(deps.Services.Clients, deps.Services.Audit, deps.Cache)
	userHandler := userfeature.NewHandler(deps.Services.Users, deps.Services.Audit, deps.Cache)
	metricsHandler := handler.NewMetricsHandler(deps.Services.Metrics)
	auditHandler := auditfeature.NewHandler(deps.Store, deps.Cache)
	notificationsHandler := handler.NewNotificationsHandler(deps.Services.Notifications)

	documentHandler := handler.NewDocumentHandler(
		deps.Services.Documents,
		deps.Services.Audit,
		deps.Services.StorageLocations,
		deps.FileStorage,
		deps.StorageFactory,
	)

	storageLocationHandler := storagelocationfeature.NewHandler(
		deps.Services.StorageLocations,
		deps.Services.Audit,
	)
	sessionHandler := sessionfeature.NewHandler(deps.Services.Sessions)
	dataQualityHandler := handler.NewDataQualityHandler(deps.Databases.DWH)
	announcementHandler := announcementfeature.NewHandler(
		deps.Services.Announcements,
		deps.Services.Audit,
	)
	adminunitsHandler := handler.NewAdminUnitsHandler(deps.Config, deps.Databases.DWH)
	visualiserHandler := handler.NewVisualiserHandler(deps.Config, deps.Databases.DWH)
	geoJSONHandler := handler.NewGeoJSONHandler("./assets/geojson")
	emailHandler := emailfeature.NewHandler(deps.Services.Email, deps.Repositories.Email)

	surveillanceHandler := handler.NewSurveillanceHandler(
		deps.Services.EpiWeeks,
		deps.Services.Diseases,
		deps.Services.Locations,
		deps.Services.FacilityWeeklyMetrics,
		deps.Services.WeeklyStatus,
		deps.Services.Alerts,
		deps.Services.SurveillanceImport,
	)

	documentTemplateHandler := handler.NewDocumentTemplateHandler(
		deps.Services.DocumentTemplates,
		deps.Services.DocumentTemplateSheets,
		deps.Services.DocumentTemplateColumns,
	)
	documentTemplateSheetHandler := handler.NewDocumentTemplateSheetHandler(
		deps.Services.DocumentTemplateSheets,
	)
	documentTemplateColumnHandler := handler.NewDocumentTemplateColumnHandler(
		deps.Services.DocumentTemplateColumns,
	)

	healthHandler := handler.NewHealthHandler(
		func(ctx context.Context) error { return db.PingDB(ctx, deps.Databases.Primary) },
		func(ctx context.Context) error { return db.PingDB(ctx, deps.Databases.Remote) },
		func(ctx context.Context) error { return deps.AdminKeycloak.Authenticate() },
		deps.Redis,
	)

	return handlers{
		Router: router.HandlerSet{
			Auth:                    authHandler,
			Clients:                 clientHandler,
			Users:                   userHandler,
			Metrics:                 metricsHandler,
			Audit:                   auditHandler,
			Notifications:           notificationsHandler,
			Documents:               documentHandler,
			DocumentTemplates:       documentTemplateHandler,
			DocumentTemplateSheets:  documentTemplateSheetHandler,
			DocumentTemplateColumns: documentTemplateColumnHandler,
			StorageLocations:        storageLocationHandler,
			Sessions:                sessionHandler,
			DataQuality:             dataQualityHandler,
			Announcements:           announcementHandler,
			AdminUnits:              adminunitsHandler,
			Visualiser:              visualiserHandler,
			Surveillance:            surveillanceHandler,
			GeoJSON:                 geoJSONHandler,
			Email:                   emailHandler,
		},
		Health: healthHandler,
	}
}
