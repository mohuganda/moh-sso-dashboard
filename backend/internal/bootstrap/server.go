package bootstrap

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/config"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/service"
)

func runHTTPServer(
	cancel context.CancelFunc,
	cfg *config.Config,
	appLogger *logger.Logger,
	notifications service.NotificationsService,
	handler *gin.Engine,
) {
	server := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           handler,
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

	notifications.NotifySystemShutdown(context.Background(), "shutdown signal received")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("Server shutdown error: ", err)
	}

	appLogger.Info("MOH SSO Dashboard exited cleanly")
}
