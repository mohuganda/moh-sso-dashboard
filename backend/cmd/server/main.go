package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	router "github.com/moh-sso-dashboard/internal/api"
	"github.com/moh-sso-dashboard/internal/api/handler"
	"github.com/moh-sso-dashboard/internal/cache"
	"github.com/moh-sso-dashboard/internal/config"
	storepkg "github.com/moh-sso-dashboard/internal/db/sqlc"
	kcClientPkg "github.com/moh-sso-dashboard/internal/keycloak"
	logger "github.com/moh-sso-dashboard/internal/log"
	db "github.com/moh-sso-dashboard/internal/migrate"
	"github.com/moh-sso-dashboard/internal/ratelimit"
	"github.com/moh-sso-dashboard/internal/service"
	importSvc "github.com/moh-sso-dashboard/internal/service/import"
	"github.com/moh-sso-dashboard/internal/storage"
	"github.com/moh-sso-dashboard/internal/worker"

	announcementRepo "github.com/moh-sso-dashboard/internal/repository/announcements"
	authRepo "github.com/moh-sso-dashboard/internal/repository/auth"
	clientRepo "github.com/moh-sso-dashboard/internal/repository/client"
	documentRepo "github.com/moh-sso-dashboard/internal/repository/document"
	metricsRepo "github.com/moh-sso-dashboard/internal/repository/metrics"
	notificationsRepo "github.com/moh-sso-dashboard/internal/repository/notifications"
	processRepo "github.com/moh-sso-dashboard/internal/repository/processes"
	sessionRepository "github.com/moh-sso-dashboard/internal/repository/session"
	storageLocationRepo "github.com/moh-sso-dashboard/internal/repository/storage_locations"
	userRepo "github.com/moh-sso-dashboard/internal/repository/user"

	"github.com/rs/zerolog"
)

func main() {

	// ==================================================
	// Root Context
	// ==================================================
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ==================================================
	// Load Config
	// ==================================================
	cfg, err := config.LoadConfig(".")
	if err != nil {
		panic("cannot load config: " + err.Error())
	}

	appLogger := logger.NewLogger()
	appLogger.SetLevel(zerolog.InfoLevel)
	appLogger.Info("Starting MOH SSO Dashboard - Environment: " + cfg.Environment)

	// ==================================================
	// Initialize Primary DB
	// ==================================================
	primaryDB, err := db.InitDB(ctx, db.DBConfig{
		Driver:          cfg.DbDriver,
		DSN:             cfg.DbSource(),
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 30 * time.Minute,
		WaitTimeout:     30 * time.Second,
	})
	if err != nil {
		appLogger.Fatal("Failed to initialize primary DB: ", err)
	}
	defer primaryDB.Close()

	// ==================================================
	// Initialize Remote DB
	// ==================================================
	remoteDB, err := db.InitDB(ctx, db.DBConfig{
		Driver:          cfg.DbDriver,
		DSN:             cfg.RemoteDbSource(),
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 30 * time.Minute,
		WaitTimeout:     30 * time.Second,
	})
	if err != nil {
		appLogger.Fatal("Failed to initialize remote DB: ", err)
	}
	defer remoteDB.Close()

	// ==================================================
	// Initialize DWH DB
	// ==================================================

	dwhDB, err := db.InitDB(ctx, db.DBConfig{
		Driver:          cfg.DbDriver,
		DSN:             cfg.DwhDbSource(),
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 30 * time.Minute,
		WaitTimeout:     30 * time.Second,
	})
	if err != nil {
		appLogger.Fatal("Failed to initialize dwh DB: ", err)
	}
	defer dwhDB.Close()

	// ==================================================
	// Run Migrations
	// ==================================================
	if err := db.MigrateDB(primaryDB, "file://internal/db/migrations"); err != nil {
		appLogger.Fatal("Migration failed: ", err)
	}

	// ==================================================
	// Redis
	// ==================================================
	rdb := cache.NewRedisClient(cache.RedisConfig{
		Host:         cfg.RedisHost,
		Port:         cfg.RedisPort,
		Password:     cfg.RedisPassword,
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	cache.MustPing(ctx, rdb)
	cacheAdapter := cache.NewRedisCache(rdb)
	rateLimiter := ratelimit.New(rdb)

	appLogger.Info("Redis connected")

	// ==================================================
	// Keycloak
	// ==================================================
	adminKC := kcClientPkg.NewAdminClient(
		cfg.KeycloakBaseUrl,
		cfg.KeycloakRealm,
		cfg.KeycloakAdminClientID,
		cfg.KeycloakAdminClientSecret,
	)

	if err := adminKC.Authenticate(); err != nil {
		appLogger.Fatal("Keycloak admin authentication failed: ", err)
	}

	webKC := kcClientPkg.NewWebClient(
		cfg.KeycloakBaseUrl,
		cfg.KeycloakRealm,
		cfg.KeycloakWebClientID,
		cfg.KeycloakWebClientSecret,
		cacheAdapter,
		cfg,
	)

	appLogger.Info("Keycloak clients initialized")

	// ==================================================
	// Storage
	// ==================================================
	fileStorage, err := storage.NewFileStorage(cfg.StorageProvider, cfg)
	if err != nil {
		appLogger.Fatal("Storage initialization failed: ", err)
	}

	storageFactory := storage.NewStorageFactory(cfg)
	appLogger.Info("Storage provider initialized: " + cfg.StorageProvider)

	// ==================================================
	// Store (SQLC)
	// ==================================================
	store := storepkg.NewStore(primaryDB)

	// ==================================================
	// Repositories
	// ==================================================
	authRepository := authRepo.NewAuthRepository(webKC, adminKC, cfg)
	clientRepository := clientRepo.NewClientRepository(adminKC, cfg, store, *appLogger)
	userRepository := userRepo.NewUserRepository(adminKC, cfg, store, *appLogger)
	metricsRepository := metricsRepo.NewMetricsRepository(cfg, store, *appLogger)
	notificationsRepository := notificationsRepo.NewNotificationsRepository(store, *appLogger)
	documentRepository := documentRepo.NewDocumentRepository(cfg, store, *appLogger)
	processRepository := processRepo.NewProcessRepository(cfg, store, *appLogger)
	fileRepository := documentRepo.NewFileRepository()
	storageRepo := storageLocationRepo.NewStorageRepositoryRepository(cfg, store, *appLogger)
	sessionRepository := sessionRepository.NewSessionRepository(adminKC, cfg, *appLogger)
	announcementRepository := announcementRepo.NewAnnouncementRepository(store, *appLogger)

	// ==================================================
	// Services
	// ==================================================
	authService := service.NewAuthService(authRepository, rdb)
	metricsService := service.NewMetricsService(metricsRepository)
	auditService := service.NewAuditService(store, cacheAdapter)
	storageLocationService := service.NewStorageLocationService(storageRepo)
	sessionService := service.NewSessionService(sessionRepository)

	publisher := cache.NewNotificationPublisher(rdb)
	notificationsService := service.NewNotificationsService(notificationsRepository, publisher)

	documentService := service.NewDocumentService(documentRepository, processRepository, notificationsService, fileStorage)
	clientService := service.NewClientService(clientRepository, notificationsService)
	userService := service.NewUserService(userRepository, notificationsService)
	announcementService := service.NewAnnouncementService(announcementRepository, notificationsService)

	importService := importSvc.NewService(
		documentRepository,
		processRepository,
		fileRepository,
		fileStorage,
		remoteDB,
	)

	// ==================================================
	// Background Worker
	// ==================================================
	w := worker.NewWorker(
		processRepository,
		importService,
		3*time.Second,
		fileStorage,
	)

	go func() {
		appLogger.Info("Background worker started")
		if err := w.Start(ctx); err != nil {
			appLogger.Error("Worker stopped with error: ", err)
		}
	}()

	// ==================================================
	// Handlers
	// ==================================================
	authHandler := handler.NewAuthHandler(authService, auditService, notificationsService, cfg)
	clientHandler := handler.NewClientHandler(clientService, auditService, cacheAdapter)
	userHandler := handler.NewUserHandler(userService, auditService, cacheAdapter)
	metricsHandler := handler.NewMetricsHandler(metricsService)
	auditHandler := handler.NewAuditHandler(store, cacheAdapter)
	notificationsHandler := handler.NewNotificationsHandler(notificationsService)
	documentHandler := handler.NewDocumentHandler(documentService, auditService, storageLocationService, fileStorage, storageFactory)
	storageLocationHandler := handler.NewStorageLocationHandler(storageLocationService, auditService)
	sessionHandler := handler.NewSessionHandler(sessionService)
	announcementHandler := handler.NewAnnouncementHandler(announcementService, auditService)
	adminunitsHandler := handler.NewAdminUnitsHandler(cfg, dwhDB)
	visualiserHandler := handler.NewVisualiserHandler(cfg, dwhDB)

	healthHandler := handler.NewHealthHandler(
		func(ctx context.Context) error { return db.PingDB(ctx, primaryDB) },
		func(ctx context.Context) error { return db.PingDB(ctx, remoteDB) },
		func(ctx context.Context) error { return adminKC.Authenticate() },
		rdb,
	)

	// ==================================================
	// Router
	// ==================================================
	r := router.SetupRouter(
		webKC,
		rateLimiter,
		authHandler,
		clientHandler,
		userHandler,
		metricsHandler,
		auditService,
		auditHandler,
		notificationsHandler,
		documentHandler,
		storageLocationHandler,
		sessionHandler,
		announcementHandler,
		adminunitsHandler,
		visualiserHandler,
	)

	r.GET("/health/live", healthHandler.HandleLive)
	r.GET("/health/ready", healthHandler.HandleReady)
	r.GET("/health", healthHandler.HandleHealth)

	// ==================================================
	// HTTP Server (Hardened)
	// ==================================================
	server := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           r,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	go func() {
		appLogger.Info("Server listening on :" + cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Server failed: ", err)
		}
	}()

	// ==================================================
	// Graceful Shutdown
	// ==================================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutdown signal received")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("Server shutdown error: ", err)
	}

	appLogger.Info("MOH SSO Dashboard exited cleanly")
}
