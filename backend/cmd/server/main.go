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
	announcementRepo "github.com/moh-sso-dashboard/internal/repository/announcements"
	authRepo "github.com/moh-sso-dashboard/internal/repository/auth"
	clientRepo "github.com/moh-sso-dashboard/internal/repository/client"
	documentRepo "github.com/moh-sso-dashboard/internal/repository/document"
	documentTemplateRepo "github.com/moh-sso-dashboard/internal/repository/document_template"
	documentTemplateColumnRepo "github.com/moh-sso-dashboard/internal/repository/document_template_column"
	documentTemplateSheetRepo "github.com/moh-sso-dashboard/internal/repository/document_template_sheet"
	emailRepo "github.com/moh-sso-dashboard/internal/repository/email"
	metricsRepo "github.com/moh-sso-dashboard/internal/repository/metrics"
	notificationsRepo "github.com/moh-sso-dashboard/internal/repository/notifications"
	processRepo "github.com/moh-sso-dashboard/internal/repository/processes"
	sessionRepository "github.com/moh-sso-dashboard/internal/repository/session"
	storageLocationRepo "github.com/moh-sso-dashboard/internal/repository/storage_locations"
	repository "github.com/moh-sso-dashboard/internal/repository/surveillance"
	userRepo "github.com/moh-sso-dashboard/internal/repository/user"
	"github.com/moh-sso-dashboard/internal/service"
	importSvc "github.com/moh-sso-dashboard/internal/service/import"
	"github.com/moh-sso-dashboard/internal/storage"
	"github.com/moh-sso-dashboard/internal/worker"

	"github.com/rs/zerolog"
)

func main() {
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

	// ==================================================
	// DATABASES
	// ==================================================
	primaryDB, err := db.InitDB(ctx, db.DBConfig{
		Driver:          cfg.DBDriver,
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

	remoteDB, err := db.InitDB(ctx, db.DBConfig{
		Driver:          cfg.DBDriver,
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

	dwhDB, err := db.InitDB(ctx, db.DBConfig{
		Driver:          cfg.DBDriver,
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

	if err := db.MigrateDB(primaryDB, "file://internal/db/migrations"); err != nil {
		appLogger.Fatal("Migration failed: ", err)
	}

	if err := db.MigrateDB(remoteDB, "file://internal/db/migrations/remote"); err != nil {
		appLogger.Fatal("Remote DB migration failed: ", err)
	}

	// ==================================================
	// REDIS
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

	// ==================================================
	// KEYCLOAK
	// ==================================================
	adminKC := kcClientPkg.NewAdminClient(
		cfg.KeycloakBaseURL,
		cfg.KeycloakRealm,
		cfg.KeycloakAdminClientID,
		cfg.KeycloakAdminClientSecret,
	)

	if err := adminKC.Authenticate(); err != nil {
		appLogger.Fatal("Keycloak admin authentication failed: ", err)
	}

	webKC := kcClientPkg.NewWebClient(
		cfg.KeycloakBaseURL,
		cfg.KeycloakRealm,
		cfg.KeycloakWebClientID,
		cfg.KeycloakWebClientSecret,
		cacheAdapter,
		cfg,
	)

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
	store := storepkg.NewStore(primaryDB)

	// ==================================================
	// REPOSITORIES
	// ==================================================
	authRepository := authRepo.NewAuthRepository(webKC, adminKC, cfg)
	clientRepository := clientRepo.NewClientRepository(adminKC, cfg, store, *appLogger)
	userRepository := userRepo.NewUserRepository(adminKC, cfg, store, *appLogger)
	metricsRepository := metricsRepo.NewMetricsRepository(cfg, store, *appLogger)
	notificationsRepository := notificationsRepo.NewNotificationsRepository(store, *appLogger)

	documentRepository := documentRepo.NewDocumentRepository(cfg, store, *appLogger)
	templateImportRepository := documentRepo.NewTemplateImportRepository()

	documentTemplateRepository := documentTemplateRepo.NewDocumentTemplateRepository(store)
	documentTemplateColumnRepository := documentTemplateColumnRepo.NewDocumentTemplateColumnRepository(store)
	documentTemplateSheetRepository := documentTemplateSheetRepo.NewDocumentTemplateSheetRepository(store)

	processRepository := processRepo.NewProcessRepository(cfg, store, *appLogger)
	fileRepository := documentRepo.NewFileRepository()
	storageRepo := storageLocationRepo.NewStorageRepositoryRepository(cfg, store, *appLogger)
	sessionRepository := sessionRepository.NewSessionRepository(adminKC, cfg, *appLogger)
	announcementRepository := announcementRepo.NewAnnouncementRepository(store, *appLogger)
	emailRepository := emailRepo.NewEmailRepository(cfg, store, *appLogger)

	surveillanceRepositories := repository.NewRepositories(store)
	regionRepository := surveillanceRepositories.Regions
	districtRepository := surveillanceRepositories.Districts
	subCountyRepository := surveillanceRepositories.SubCounties
	diseaseRepository := surveillanceRepositories.Diseases
	epiWeekRepository := surveillanceRepositories.EpiWeeks
	facilityWeeklyMetricsRepository := surveillanceRepositories.FacilityMetrics
	weeklyStatusRepository := surveillanceRepositories.WeeklyStatus
	importRepository := surveillanceRepositories.Imports
	alertRepository := surveillanceRepositories.Alerts

	// ==================================================
	// SERVICES
	// ==================================================

	authService := service.NewAuthService(authRepository, rdb)
	metricsService := service.NewMetricsService(metricsRepository)
	auditService := service.NewAuditService(store, cacheAdapter)
	storageLocationService := service.NewStorageLocationService(storageRepo)
	sessionService := service.NewSessionService(sessionRepository)

	publisher := cache.NewNotificationPublisher(rdb)

	notificationsService := service.NewNotificationsService(
		notificationsRepository,
		publisher,
	)

	documentService := service.NewDocumentService(
		documentRepository,
		processRepository,
		notificationsService,
		fileStorage,
	)

	clientService := service.NewClientService(
		clientRepository,
		notificationsService,
	)

	userService := service.NewUserService(
		userRepository,
		notificationsService,
	)

	announcementService := service.NewAnnouncementService(
		announcementRepository,
		notificationsService,
	)

	diseaseService := service.NewSurveillanceDiseaseService(
		appLogger,
		diseaseRepository,
	)

	epiWeekService := service.NewSurveillanceEpiWeekService(
		appLogger,
		epiWeekRepository,
	)

	locationService := service.NewSurveillanceLocationService(
		appLogger,
		regionRepository,
		districtRepository,
		subCountyRepository,
	)

	facilityWeeklyMetricsService := service.NewSurveillanceFacilityWeeklyMetricsService(
		appLogger,
		facilityWeeklyMetricsRepository,
		importRepository,
	)

	weeklyStatusService := service.NewSurveillanceWeeklyStatusService(
		appLogger,
		weeklyStatusRepository,
	)

	surveillanceImportService := service.NewSurveillanceImportService(
		importRepository,
	)

	alertsService := service.NewSurveillanceAlertService(
		appLogger,
		alertRepository,
		importRepository,
	)

	// ==================================================
	// TEMPLATE SERVICES
	// ==================================================

	documentTemplateService := service.NewDocumentTemplateService(
		documentTemplateRepository,
		documentTemplateSheetRepository,
		documentTemplateColumnRepository,
	)

	documentTemplateSheetService := service.NewDocumentTemplateSheetService(
		documentTemplateSheetRepository,
	)

	documentTemplateColumnService := service.NewDocumentTemplateColumnService(
		documentTemplateColumnRepository,
	)

	importService := importSvc.NewService(
		documentRepository,
		templateImportRepository,
		processRepository,
		fileRepository,
		importRepository,
		documentTemplateService,
		facilityWeeklyMetricsService,
		weeklyStatusService,
		alertsService,
		fileStorage,
		remoteDB,
	)
	templateManager, err := service.NewTemplateManager(appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize email template manager: ", err)
	}

	if err := service.RegisterDefaultTemplates(templateManager); err != nil {
		appLogger.Fatal("Failed to register default email templates: ", err)
	}

	smtpService, err := service.NewSMTPService(cfg, templateManager, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize SMTP service: ", err)
	}

	queueService, err := service.NewQueueService(emailRepository, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize email queue service: ", err)
	}

	emailAppService, err := service.NewEmailService(smtpService, queueService, templateManager)
	if err != nil {
		appLogger.Fatal("Failed to initialize email application service: ", err)
	}
	// ==================================================
	// HANDLERS
	// ==================================================
	authHandler := handler.NewAuthHandler(authService, auditService, notificationsService, cfg)
	clientHandler := handler.NewClientHandler(clientService, auditService, cacheAdapter)
	userHandler := handler.NewUserHandler(userService, auditService, cacheAdapter)
	metricsHandler := handler.NewMetricsHandler(metricsService)
	auditHandler := handler.NewAuditHandler(store, cacheAdapter)
	notificationsHandler := handler.NewNotificationsHandler(notificationsService)

	documentHandler := handler.NewDocumentHandler(
		documentService,
		auditService,
		storageLocationService,
		fileStorage,
		storageFactory,
		primaryDB,
		remoteDB,
	)

	storageLocationHandler := handler.NewStorageLocationHandler(storageLocationService, auditService)
	sessionHandler := handler.NewSessionHandler(sessionService)
	dataQualityHandler := handler.NewDataQualityHandler(dwhDB)
	announcementHandler := handler.NewAnnouncementHandler(announcementService, auditService)
	adminunitsHandler := handler.NewAdminUnitsHandler(cfg, dwhDB)
	visualiserHandler := handler.NewVisualiserHandler(cfg, dwhDB)
	geoJSONHandler := handler.NewGeoJSONHandler("./assets/geojson")

	emailHandler := handler.NewEmailHandler(emailAppService, emailRepository)

	surveillanceHandler := handler.NewSurveillanceHandler(
		epiWeekService,
		diseaseService,
		locationService,
		facilityWeeklyMetricsService,
		weeklyStatusService,
		alertsService,
		surveillanceImportService,
	)

	documentTemplateHandler := handler.NewDocumentTemplateHandler(
		documentTemplateService,
		documentTemplateSheetService,
		documentTemplateColumnService,
		primaryDB,
		remoteDB,
	)

	documentTemplateSheetHandler := handler.NewDocumentTemplateSheetHandler(
		documentTemplateSheetService,
	)

	documentTemplateColumnHandler := handler.NewDocumentTemplateColumnHandler(
		documentTemplateColumnService,
	)

	healthHandler := handler.NewHealthHandler(
		func(ctx context.Context) error { return db.PingDB(ctx, primaryDB) },
		func(ctx context.Context) error { return db.PingDB(ctx, remoteDB) },
		func(ctx context.Context) error { return adminKC.Authenticate() },
		rdb,
	)

	// ==================================================
	// BACKGROUND WORKERS
	// ==================================================

	documentWorker, err := worker.NewDocumentWorker(
		processRepository,
		importService,
		3*time.Second,
		fileStorage,
		appLogger,
	)
	if err != nil {
		appLogger.Fatal("Failed to initialize document worker: ", err)
	}
	go func() {
		appLogger.Info("Background document worker started")
		if err := documentWorker.Start(ctx); err != nil {
			appLogger.Error("Document worker stopped with error: ", err)
		}
	}()
	emailWorker, err := worker.NewEmailWorker(
		emailRepository,
		smtpService,
		3*time.Second,
		20,
		3,
		appLogger,
	)
	if err != nil {
		appLogger.Fatal("Failed to initialize email worker: ", err)
	}
	go func() {
		appLogger.Info("Background email worker started")
		if err := emailWorker.Start(ctx); err != nil {
			appLogger.Error("Email worker stopped with error: ", err)
		}
	}()

	// ==================================================
	// ROUTER
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
		documentTemplateHandler,
		documentTemplateSheetHandler,
		documentTemplateColumnHandler,
		storageLocationHandler,
		sessionHandler,
		dataQualityHandler,
		announcementHandler,
		adminunitsHandler,
		visualiserHandler,
		surveillanceHandler,
		geoJSONHandler,
		emailHandler,
	)

	r.GET("/health/live", healthHandler.HandleLive)
	r.GET("/health/ready", healthHandler.HandleReady)
	r.GET("/health", healthHandler.HandleHealth)

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
