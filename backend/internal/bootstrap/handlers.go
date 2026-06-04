package bootstrap

import (
	"context"

	router "github.com/moh-sso-dashboard/internal/api"
	"github.com/moh-sso-dashboard/internal/api/handler"
	"github.com/moh-sso-dashboard/internal/cache"
	"github.com/moh-sso-dashboard/internal/config"
	storepkg "github.com/moh-sso-dashboard/internal/db/sqlc"
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
	clientHandler := handler.NewClientHandler(deps.Services.Clients, deps.Services.Audit, deps.Cache)
	userHandler := handler.NewUserHandler(deps.Services.Users, deps.Services.Audit, deps.Cache)
	metricsHandler := handler.NewMetricsHandler(deps.Services.Metrics)
	auditHandler := handler.NewAuditHandler(deps.Store, deps.Cache)
	notificationsHandler := handler.NewNotificationsHandler(deps.Services.Notifications)

	documentHandler := handler.NewDocumentHandler(
		deps.Services.Documents,
		deps.Services.Audit,
		deps.Services.StorageLocations,
		deps.FileStorage,
		deps.StorageFactory,
	)

	storageLocationHandler := handler.NewStorageLocationHandler(
		deps.Services.StorageLocations,
		deps.Services.Audit,
	)
	sessionHandler := handler.NewSessionHandler(deps.Services.Sessions)
	dataQualityHandler := handler.NewDataQualityHandler(deps.Databases.DWH)
	announcementHandler := handler.NewAnnouncementHandler(
		deps.Services.Announcements,
		deps.Services.Audit,
	)
	adminunitsHandler := handler.NewAdminUnitsHandler(deps.Config, deps.Databases.DWH)
	visualiserHandler := handler.NewVisualiserHandler(deps.Config, deps.Databases.DWH)
	geoJSONHandler := handler.NewGeoJSONHandler("./assets/geojson")
	emailHandler := handler.NewEmailHandler(deps.Services.Email, deps.Repositories.Email)

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
