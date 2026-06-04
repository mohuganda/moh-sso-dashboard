package bootstrap

import (
	"context"

	router "github.com/moh-sso-dashboard/internal/api"
	"github.com/moh-sso-dashboard/internal/config"
	storepkg "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/storage"

	"github.com/rs/zerolog"
)

func Run() {
	// ==================================================
	// Root Context
	// ==================================================
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ==================================================
	// CONFIG
	// ==================================================
	cfg, err := config.LoadConfig(".")
	if err != nil {
		panic("cannot load config: " + err.Error())
	}

	appLogger := logger.NewLogger()
	appLogger.SetLevel(zerolog.InfoLevel)
	appLogger.Info("Starting MOH SSO Dashboard - Environment: " + cfg.Environment)

	appLogger.Info(
		"notification config loaded",
		"platform_name", cfg.Notification.PlatformName,
		"system_admin_name", cfg.Notification.SystemAdminName,
		"system_admin_email", cfg.Notification.SystemAdminEmail,
		"admin_dashboard_url", cfg.Notification.AdminDashboardURL,
	)

	appLogger.Info(
		"smtp config loaded",
		"smtp_host", cfg.SMTP.Host,
		"smtp_port", cfg.SMTP.Port,
		"smtp_from_email", cfg.SMTP.FromEmail,
		"smtp_from_name", cfg.SMTP.FromName,
	)

	// ==================================================
	// DATABASES
	// ==================================================
	dbs, err := initDatabases(ctx, cfg)
	if err != nil {
		appLogger.Fatal("Failed to initialize databases: ", err)
	}
	defer dbs.Close()

	// ==================================================
	// REDIS
	// ==================================================
	cacheRuntime := initCache(ctx, cfg)
	rdb := cacheRuntime.Redis
	cacheAdapter := cacheRuntime.Cache
	rateLimiter := cacheRuntime.RateLimiter

	// ==================================================
	// KEYCLOAK
	// ==================================================
	keycloakClients, err := initKeycloak(cfg, cacheAdapter)
	if err != nil {
		appLogger.Fatal("Keycloak admin authentication failed: ", err)
	}
	adminKC := keycloakClients.Admin
	webKC := keycloakClients.Web

	// ==================================================
	// STORAGE
	// ==================================================
	fileStorage, err := storage.NewFileStorage(cfg.StorageProvider, cfg)
	if err != nil {
		appLogger.Fatal("Storage initialization failed: ", err)
	}

	storageFactory := storage.NewStorageFactory(cfg)

	// ==================================================
	// STORE
	// ==================================================
	store := storepkg.NewStore(dbs.Primary)

	// ==================================================
	// REPOSITORIES
	// ==================================================
	repos := buildRepositories(cfg, store, adminKC, webKC, appLogger)

	// ==================================================
	// SERVICES
	// ==================================================
	services := buildServices(serviceDependencies{
		Config:       cfg,
		Store:        store,
		Cache:        cacheAdapter,
		CacheClient:  rdb,
		Databases:    dbs,
		Repositories: repos,
		FileStorage:  fileStorage,
		Logger:       appLogger,
	})

	// ==================================================
	// HANDLERS
	// ==================================================
	handlers := buildHandlers(handlerDependencies{
		Config:         cfg,
		Store:          store,
		Cache:          cacheAdapter,
		Databases:      dbs,
		Repositories:   repos,
		Services:       services,
		FileStorage:    fileStorage,
		StorageFactory: storageFactory,
		AdminKeycloak:  adminKC,
		Redis:          rdb,
	})

	startBackgroundWorkers(ctx, workerDependencies{
		ProcessRepository:              repos.Processes,
		ImportService:                  services.Import,
		FileStorage:                    fileStorage,
		EmailRepository:                repos.Email,
		SMTPService:                    services.SMTP,
		NotificationDeliveryRepository: repos.NotificationDelivery,
		EmailService:                   services.Email,
		Logger:                         appLogger,
	})

	services.Notifications.NotifySystemStartup(ctx, cfg.Environment)

	// ==================================================
	// ROUTER
	// ==================================================
	r := router.SetupRouter(router.RouterDependencies{
		KeycloakClient: webKC,
		Limiter:        rateLimiter,
		AuditService:   services.Audit,
		Handlers:       handlers.Router,
	})

	router.RegisterHealthRoutes(r, handlers.Health)

	runHTTPServer(cancel, cfg, appLogger, services.Notifications, r)
}
