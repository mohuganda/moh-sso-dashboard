package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/moh-sso-dashboard/internal/observability"
)

func RequestContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(observability.RequestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		correlationID := c.GetHeader(observability.CorrelationIDHeader)
		if correlationID == "" {
			correlationID = requestID
		}

		c.Set("request_id", requestID)
		c.Set("correlation_id", correlationID)
		c.Header(observability.RequestIDHeader, requestID)
		c.Header(observability.CorrelationIDHeader, correlationID)

		ctx := observability.WithRequestID(c.Request.Context(), requestID)
		ctx = observability.WithCorrelationID(ctx, correlationID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		event := log.Info()
		if len(c.Errors) > 0 || c.Writer.Status() >= 500 {
			event = log.Error()
		} else if c.Writer.Status() >= 400 {
			event = log.Warn()
		}

		event.
			Str("request_id", observability.RequestIDFromContext(c.Request.Context())).
			Str("correlation_id", observability.CorrelationIDFromContext(c.Request.Context())).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("route", c.FullPath()).
			Int("status", c.Writer.Status()).
			Int("bytes", c.Writer.Size()).
			Dur("latency", time.Since(start)).
			Str("ip", c.ClientIP()).
			Str("user_id", c.GetString("user_id")).
			Msg("http request completed")
	}
}
