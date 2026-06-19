package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/moh-sso-dashboard/internal/observability"
	"github.com/moh-sso-dashboard/internal/service"
	"github.com/moh-sso-dashboard/internal/utils"
)

func AuditMiddleware(audit *service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Skip auditing of health checks or static routes
		if c.FullPath() == "/health" {
			c.Next()
			return
		}

		start := time.Now()

		// Process request first
		c.Next()

		duration := time.Since(start)

		userID := c.GetString("user_id")

		// Build metadata
		meta := map[string]interface{}{
			"request_id":     observability.RequestIDFromContext(c.Request.Context()),
			"correlation_id": observability.CorrelationIDFromContext(c.Request.Context()),
			"method":         c.Request.Method,
			"status":         c.Writer.Status(),
			"ip":             c.ClientIP(),
			"user_agent":     c.Request.UserAgent(),
			"latency_ms":     duration.Milliseconds(),
		}

		err := audit.Log(
			c.Request.Context(),
			utils.ToNullUUID(userID),
			"API_CALL:"+c.FullPath(),
			meta,
		)

		if err != nil {
			// Log error internally, don't interrupt API
			log.Error().
				Err(err).
				Str("request_id", observability.RequestIDFromContext(c.Request.Context())).
				Msg("failed to record audit log")
		}
	}
}
