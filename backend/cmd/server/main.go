package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	router "github.com/moh-sso-dashboard/internal/api"
	"github.com/moh-sso-dashboard/internal/api/handler"
	"github.com/moh-sso-dashboard/internal/cache"
	"github.com/moh-sso-dashboard/internal/config"
	storepkg "github.com/moh-sso-dashboard/internal/db/sqlc"
	kcClientPkg "github.com/moh-sso-dashboard/internal/keycloak"
	logger "github.com/moh-sso-dashboard/internal/log"
	migrate "github.com/moh-sso-dashboard/internal/migrate"
	"github.com/moh-sso-dashboard/internal/storage"
	"github.com/moh-sso-dashboard/internal/worker"

	authRepo "github.com/moh-sso-dashboard/internal/repository/auth"
	clientRepo "github.com/moh-sso-dashboard/internal/repository/client"
	documentRepo "github.com/moh-sso-dashboard/internal/repository/document"
	metricsRepo "github.com/moh-sso-dashboard/internal/repository/metrics"
	notificationsRepo "github.com/moh-sso-dashboard/internal/repository/notifications"
	processRepo "github.com/moh-sso-dashboard/internal/repository/processes"
	userRepo "github.com/moh-sso-dashboard/internal/repository/user"

	"github.com/moh-sso-dashboard/internal/ratelimit"
	"github.com/moh-sso-dashboard/internal/service"

	importSvc "github.com/moh-sso-dashboard/internal/service/import"

	"github.com/rs/zerolog"
)

func main() {

	// --------------------------------------------------
	// Root Context (shared by server + worker)
	// --------------------------------------------------
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --------------------------------------------------
	// Load Configuration
	// --------------------------------------------------
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	appLogger := logger.NewLogger()
	appLogger.SetLevel(zerolog.InfoLevel)
	appLogger.Info("Starting server in environment: " + cfg.Environment)

	// --------------------------------------------------
	// Database (pq + database/sql)
	// --------------------------------------------------
	conn, err := sql.Open(cfg.DbDriver, cfg.DbSource())
	if err != nil {
		appLogger.Fatal("Cannot open database connection: ", err)
	}
	defer conn.Close()

	// Connection pool tuning (important in production)
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(10)
	conn.SetConnMaxLifetime(30 * time.Minute)

	if err := conn.PingContext(ctx); err != nil {
		appLogger.Fatal("Cannot connect to database: ", err)
	}
	appLogger.Info("Successfully connected to database")

	remoteConn, err := sql.Open(cfg.DbDriver, cfg.RemoteDbSource())

	if err != nil {
		appLogger.Fatal("Cannot open remote database connection: ", err)
	}
	defer conn.Close()

	// Connection pool tuning (important in production)
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(10)
	conn.SetConnMaxLifetime(30 * time.Minute)

	if err := conn.PingContext(ctx); err != nil {
		appLogger.Fatal("Cannot connect to  remote database: ", err)
	}
	appLogger.Info("Successfully connected to remote database")

	// Run migrations
	if err := migrate.MigrateDB(conn, "file://internal/db/migrations"); err != nil {
		appLogger.Fatal("Cannot migrate db: ", err)
	}

	// --------------------------------------------------
	// Redis
	// --------------------------------------------------
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
	appLogger.Info("Successfully connected to Redis")

	cacheAdapter := cache.NewRedisCache(rdb)
	rateLimiter := ratelimit.New(rdb)

	// --------------------------------------------------
	// Keycloak
	// --------------------------------------------------
	adminKC := kcClientPkg.NewAdminClient(
		cfg.KeycloakBaseUrl,
		cfg.KeycloakRealm,
		cfg.KeycloakAdminClientID,
		cfg.KeycloakAdminClientSecret,
	)

	if err := adminKC.Authenticate(); err != nil {
		appLogger.Fatal("Failed to authenticate Keycloak admin service account: ", err)
	}

	webKC := kcClientPkg.NewWebClient(
		cfg.KeycloakBaseUrl,
		cfg.KeycloakRealm,
		cfg.KeycloakWebClientID,
		cfg.KeycloakWebClientSecret,
		cacheAdapter,
	)

	appLogger.Info("Keycloak clients initialized successfully")

	// --------------------------------------------------
	// Initialize Storage
	// --------------------------------------------------
	storageProvider := cfg.StorageProvider

	fileStorage, err := storage.NewFileStorage(storageProvider, cfg)
	if err != nil {
		appLogger.Fatal("failed to initialize storage: ", err)
	}

	appLogger.Info("Storage provider initialized: " + storageProvider)

	// --------------------------------------------------
	// Infrastructure Store
	// --------------------------------------------------
	store := storepkg.NewStore(conn)

	// --------------------------------------------------
	// Repositories
	// --------------------------------------------------
	authRepository := authRepo.NewAuthRepository(webKC, cfg)
	clientRepository := clientRepo.NewClientRepository(adminKC, cfg, store, *appLogger)
	userRepository := userRepo.NewUserRepository(adminKC, cfg, store, *appLogger)
	metricsRepository := metricsRepo.NewMetricsRepository(cfg, store, *appLogger)
	notificationsRepository := notificationsRepo.NewNotificationsRepository(store, *appLogger)
	documentRepository := documentRepo.NewDocumentRepository(cfg, store, *appLogger)
	processRepository := processRepo.NewProcessRepository(cfg, store, *appLogger)
	fileRepository := documentRepo.NewFileRepository(remoteConn)

	// --------------------------------------------------
	// Services
	// --------------------------------------------------
	authService := service.NewAuthService(authRepository, rdb)
	metricsService := service.NewMetricsService(metricsRepository)
	auditService := service.NewAuditService(store, cacheAdapter)

	publisher := cache.NewNotificationPublisher(rdb)
	notificationsService := service.NewNotificationsService(notificationsRepository, publisher)

	documentService := service.NewDocumentService(documentRepository, notificationsService, fileStorage)
	clientService := service.NewClientService(clientRepository, notificationsService)
	userService := service.NewUserService(userRepository, notificationsService)

	importService := importSvc.NewService(
		documentRepository,
		processRepository,
		fileStorage,
	)

	// --------------------------------------------------
	// Background Worker (NON-BLOCKING)
	// --------------------------------------------------
	w := worker.NewWorker(
		processRepository,
		importService,
		3*time.Second,
		fileStorage,
	)

	go func() {
		appLogger.Info("Background worker started")
		w.Start(ctx)
	}()

	// --------------------------------------------------
	// Handlers
	// --------------------------------------------------
	authHandler := handler.NewAuthHandler(authService, auditService, notificationsService, cfg)
	clientHandler := handler.NewClientHandler(clientService, auditService, cacheAdapter)
	userHandler := handler.NewUserHandler(userService, auditService, cacheAdapter)
	metricsHandler := handler.NewMetricsHandler(metricsService)
	auditHandler := handler.NewAuditHandler(store, cacheAdapter)
	notificationsHandler := handler.NewNotificationsHandler(notificationsService)
	documentHandler := handler.NewDocumentHandler(documentService, auditService, fileStorage)

	// Health handler
	healthHandler := handler.NewHealthHandler(
		func(ctx context.Context) error {
			return conn.PingContext(ctx)
		},
		func(ctx context.Context) error {
			return adminKC.Authenticate()
		},
		rdb,
	)

	// --------------------------------------------------
	// Router
	// --------------------------------------------------
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
	)

	r.GET("/health/live", healthHandler.HandleLive)
	r.GET("/health/ready", healthHandler.HandleReady)
	r.GET("/health", healthHandler.HandleHealth)

	addr := ":" + cfg.ServerPort

	ln, err := net.Listen("tcp4", addr)
	if err != nil {
		appLogger.Fatal("Failed to bind IPv4 listener: ", err)
	}

	server := &http.Server{
		Handler: r,
	}

	// Start HTTP server
	go func() {
		appLogger.Info("Server listening on ", addr)
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Server failed: ", err)
		}
	}()

	// --------------------------------------------------
	// Graceful Shutdown
	// --------------------------------------------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutdown signal received")

	// Stop worker
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		appLogger.Fatal("Server forced to shutdown: ", err)
	}

	appLogger.Info("Server exited properly")
}
