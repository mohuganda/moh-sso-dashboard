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
	store "github.com/moh-sso-dashboard/internal/db/sqlc"
	kcClientPkg "github.com/moh-sso-dashboard/internal/keycloak"
	logger "github.com/moh-sso-dashboard/internal/log"
	db "github.com/moh-sso-dashboard/internal/migrate"

	authRepo "github.com/moh-sso-dashboard/internal/repository/auth"
	clientRepo "github.com/moh-sso-dashboard/internal/repository/client"
	metricsRepo "github.com/moh-sso-dashboard/internal/repository/metrics"
	"github.com/moh-sso-dashboard/internal/repository/notifications"
	userRepo "github.com/moh-sso-dashboard/internal/repository/user"

	"github.com/moh-sso-dashboard/internal/ratelimit"
	"github.com/moh-sso-dashboard/internal/service"

	"github.com/rs/zerolog"
)

func main() {
	// ---------------------------------------------------------------------
	// Load configuration
	// ---------------------------------------------------------------------
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	appLogger := logger.NewLogger()
	appLogger.SetLevel(zerolog.InfoLevel)
	appLogger.Info("Starting server in environment:", cfg.Environment)

	// ---------------------------------------------------------------------
	// Database
	// ---------------------------------------------------------------------
	conn, err := sql.Open(cfg.DbDriver, cfg.DbSource())
	if err != nil {
		appLogger.Fatal("Cannot open database connection:", err)
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		appLogger.Fatal("Cannot connect to database: ", err)
	}
	appLogger.Info("Successfully connected to database")

	if err := db.MigrateDB(conn, "file://internal/db/migrations"); err != nil {
		appLogger.Fatal("Cannot migrate db:", err)
	}

	// ---------------------------------------------------------------------
	// Redis
	// ---------------------------------------------------------------------
	rdb := cache.NewRedisClient(cache.RedisConfig{
		Host:         cfg.RedisHost,
		Port:         cfg.RedisPort,
		Password:     cfg.RedisPassword,
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	cache.MustPing(context.Background(), rdb)
	appLogger.Info("Successfully connected to Redis")

	cacheAdapter := cache.NewRedisCache(rdb)
	rateLimiter := ratelimit.New(rdb)

	// ---------------------------------------------------------------------
	// Keycloak
	// ---------------------------------------------------------------------
	keycloakClient := kcClientPkg.NewClient(
		cfg.KeycloakBaseUrl,
		cfg.KeycloakRealm,
		cfg.KeycloakAdminClientID,
		cfg.KeycloakAdminClientSecret,
		cfg.KeycloakWebClientID,
		cfg.KeycloakWebClientSecret,
		cacheAdapter,
	)

	if err := keycloakClient.Authenticate(); err != nil {
		appLogger.Fatal("Failed to authenticate Keycloak admin service account: ", err)
	}
	appLogger.Info("Successfully authenticated Keycloak admin service account")

	// ---------------------------------------------------------------------
	// Infrastructure
	// ---------------------------------------------------------------------
	store := store.NewStore(conn)

	// ---------------------------------------------------------------------
	// Repositories
	// ---------------------------------------------------------------------
	authRepository := authRepo.NewAuthRepository(keycloakClient, cfg)
	clientRepository := clientRepo.NewClientRepository(
		keycloakClient,
		cfg,
		store,
		*appLogger,
	)
	userRepository := userRepo.NewUserRepository(
		keycloakClient,
		cfg,
		store,
		*appLogger,
	)
	metricsRepository := metricsRepo.NewMetricsRepository(cfg, store, *appLogger)
	notificationsRepository := notifications.NewNotificationsRepository(store, *appLogger)

	// ---------------------------------------------------------------------
	// Services
	// ---------------------------------------------------------------------
	authService := service.NewAuthService(authRepository, rdb)
	metricsService := service.NewMetricsService(metricsRepository)
	auditService := service.NewAuditService(store, cacheAdapter)
	importService := service.NewImportService(store, keycloakClient)

	publisher := cache.NewNotificationPublisher(rdb)
	notificationsService := service.NewNotificationsService(notificationsRepository, publisher)

	clientService := service.NewClientService(clientRepository, notificationsService)
	userService := service.NewUserService(userRepository, notificationsService)

	// ---------------------------------------------------------------------
	// Handlers
	// ---------------------------------------------------------------------
	authHandler := handler.NewAuthHandler(authService, auditService, notificationsService, cfg)
	clientHandler := handler.NewClientHandler(clientService, auditService, cacheAdapter)
	userHandler := handler.NewUserHandler(userService, auditService, cacheAdapter)
	metricsHandler := handler.NewMetricsHandler(metricsService)
	importHandler := handler.NewImportHandler(importService, cfg)
	auditHandler := handler.NewAuditHandler(store, cacheAdapter)
	notificationsHandler := handler.NewNotificationsHandler(notificationsService)

	// ---------------------------------------------------------------------
	// ✅ Health Handler (NEW)
	// ---------------------------------------------------------------------
	healthHandler := handler.NewHealthHandler(
		func(ctx context.Context) error {
			return conn.PingContext(ctx)
		},
		func(ctx context.Context) error {
			return keycloakClient.Authenticate()
		},
		rdb,
	)

	// ---------------------------------------------------------------------
	// Router
	// ---------------------------------------------------------------------
	r := router.SetupRouter(
		keycloakClient,
		rateLimiter,
		importHandler,
		authHandler,
		clientHandler,
		userHandler,
		metricsHandler,
		auditService,
		auditHandler,
		notificationsHandler,
	)

	// ✅ Register health routes (outside auth)
	r.GET("/health/live", healthHandler.HandleLive)
	r.GET("/health/ready", healthHandler.HandleReady)
	r.GET("/health", healthHandler.HandleHealth)

	addr := ":" + cfg.ServerPort

	ln, err := net.Listen("tcp4", addr)
	if err != nil {
		appLogger.Fatal("Failed to bind IPv4 listener: ", err)
	}

	appLogger.Info("Gin server listening on IPv4 ", addr)

	server := &http.Server{
		Handler: r,
	}

	// ---------------------------------------------------------------------
	// ✅ Graceful Shutdown (Production Safe)
	// ---------------------------------------------------------------------
	go func() {
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Gin server failed: ", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		appLogger.Fatal("Server forced to shutdown:", err)
	}

	appLogger.Info("Server exiting")
}
