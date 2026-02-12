package ratelimit

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Middleware(
	limiter *Limiter,
	keyFunc func(*gin.Context) string,
	limit int,
	window time.Duration,
) gin.HandlerFunc {

	return func(c *gin.Context) {

		key := keyFunc(c)

		allowed, remaining, reset, err := limiter.Allow(
			c.Request.Context(),
			key,
			limit,
			window,
		)

		// headers (nice for frontend)
		c.Header("X-RateLimit-Limit", fmt.Sprint(limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprint(remaining))
		c.Header("X-RateLimit-Reset", reset.Format(time.RFC3339))

		if err != nil {
			// fail-open
			c.Next()
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(
				http.StatusTooManyRequests,
				gin.H{
					"error":   "RATE_LIMIT_EXCEEDED",
					"message": "Too many requests, slow down",
				},
			)
			return
		}

		c.Next()
	}
}
