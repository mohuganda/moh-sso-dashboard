package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	DBCheck       func(ctx context.Context) error
	RemoteDBCheck func(ctx context.Context) error
	KeycloakCheck func(ctx context.Context) error
	Redis         *redis.Client
	startedAt     time.Time
}

func NewHealthHandler(
	dbCheck func(ctx context.Context) error,
	remoteDbCheck func(ctx context.Context) error,
	kcCheck func(ctx context.Context) error,
	redis *redis.Client,
) *HealthHandler {
	return &HealthHandler{
		DBCheck:       dbCheck,
		RemoteDBCheck: remoteDbCheck,
		KeycloakCheck: kcCheck,
		startedAt:     time.Now(),
		Redis:         redis,
	}
}

// --------------------------------------------------
// Liveness Probe
// Used by Kubernetes to check if process is alive
// --------------------------------------------------
func (h *HealthHandler) HandleLive(c *gin.Context) {
	response.OK(c, http.StatusOK, gin.H{
		"status":         "alive",
		"uptime_seconds": int(time.Since(h.startedAt).Seconds()),
	})
}

// --------------------------------------------------
// Readiness Probe (STRICT)
// Fails if critical dependencies are down
// --------------------------------------------------
func (h *HealthHandler) HandleReady(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Check Redis (if enabled)
	if h.Redis != nil {
		if err := h.Redis.Ping(ctx).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "not ready",
				"service": "redis",
				"error":   err.Error(),
			})
			return
		}
	}

	if err := h.RemoteDBCheck(ctx); err != nil {
		response.OK(c, http.StatusServiceUnavailable, gin.H{
			"status":         "not_ready",
			"reason":         "database_unavailable",
			"error":          err.Error(),
			"uptime_seconds": int(time.Since(h.startedAt).Seconds()),
		})
		return
	}

	if err := h.DBCheck(ctx); err != nil {
		response.OK(c, http.StatusServiceUnavailable, gin.H{
			"status":         "not_ready",
			"reason":         "database_unavailable",
			"error":          err.Error(),
			"uptime_seconds": int(time.Since(h.startedAt).Seconds()),
		})
		return
	}

	// Optional strict Keycloak check
	if err := h.KeycloakCheck(ctx); err != nil {
		response.OK(c, http.StatusServiceUnavailable, gin.H{
			"status":         "not_ready",
			"reason":         "keycloak_unavailable",
			"error":          err.Error(),
			"uptime_seconds": int(time.Since(h.startedAt).Seconds()),
		})
		return
	}

	response.OK(c, http.StatusOK, gin.H{
		"status":         "ready",
		"uptime_seconds": int(time.Since(h.startedAt).Seconds()),
	})
}

// --------------------------------------------------
// Aggregated Health (DEGRADED allowed)
// Does NOT fail pod unless DB is down
// --------------------------------------------------
func (h *HealthHandler) HandleHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	status := "ok"

	db := "ok"
	if err := h.DBCheck(ctx); err != nil {
		db = "unavailable"
		status = "degraded"
	}

	remoteDb := "ok"
	if err := h.RemoteDBCheck(ctx); err != nil {
		remoteDb = "unavailable"
		status = "degraded"
	}

	keycloak := "ok"
	if err := h.KeycloakCheck(ctx); err != nil {
		keycloak = "unavailable"
		status = "degraded"
	}

	redis := "ok"
	if err := h.Redis.Ping(ctx); err != nil {
		db = "unavailable"
		status = "degraded"
	}

	httpStatus := http.StatusOK
	if db == "unavailable" {
		httpStatus = http.StatusServiceUnavailable
	}

	response.OK(c, httpStatus, gin.H{
		"status": status,
		"components": gin.H{
			"database": db,
			"remoteDB": remoteDb,
			"keycloak": keycloak,
			"redis":    redis,
		},
		"uptime_seconds": int(time.Since(h.startedAt).Seconds()),
	})
}
