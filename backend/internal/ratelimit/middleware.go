package ratelimit

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/http/response"
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
	policy = policy.withDefaults()
	return func(c *gin.Context) {
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

		SetHeaders(c.Writer.Header(), policy.Name, policy.Limit, remaining, reset)

		if err != nil {
			// fail-open
			c.Next()
			return
		}

		if !allowed {
			response.Fail(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests, slow down")
			c.Abort()
			return
		}

		c.Next()
	}
}

// SetHeaders applies the platform rate policy metadata to an HTTP response.
func SetHeaders(header http.Header, policy string, limit, remaining int, reset time.Time) {
	header.Set("X-RateLimit-Policy", policy)
	header.Set("X-RateLimit-Limit", strconv.Itoa(limit))
	header.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
	header.Set("X-RateLimit-Reset", reset.Format(time.RFC3339))
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
