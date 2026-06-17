package ratelimit

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Policy struct {
	Name    string
	Limit   int
	Window  time.Duration
	KeyFunc func(*gin.Context) string
}

func Middleware(
	limiter *Limiter,
	keyFunc func(*gin.Context) string,
	limit int,
	window time.Duration,
) gin.HandlerFunc {
	return MiddlewareForPolicy(limiter, Policy{
		Name:    "default",
		Limit:   limit,
		Window:  window,
		KeyFunc: keyFunc,
	})
}

func MiddlewareForPolicy(limiter *Limiter, policy Policy) gin.HandlerFunc {
	return func(c *gin.Context) {
		policy = policy.withDefaults()
		if limiter == nil || limiter.redis == nil || policy.Limit <= 0 {
			c.Next()
			return
		}

		key := policy.key(c)

		allowed, remaining, reset, err := limiter.Allow(
			c.Request.Context(),
			key,
			policy.Limit,
			policy.Window,
		)

		// headers (nice for frontend)
		c.Header("X-RateLimit-Policy", policy.Name)
		c.Header("X-RateLimit-Limit", fmt.Sprint(policy.Limit))
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

func (policy Policy) withDefaults() Policy {
	if strings.TrimSpace(policy.Name) == "" {
		policy.Name = "default"
	}
	if policy.Window == 0 {
		policy.Window = time.Minute
	}
	if policy.KeyFunc == nil {
		policy.KeyFunc = ByIP
	}
	return policy
}

func (policy Policy) key(c *gin.Context) string {
	identity := policy.KeyFunc(c)
	if strings.HasPrefix(identity, "rl:") {
		return identity
	}
	return Key(policy.Name, identity)
}
