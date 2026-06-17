package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestMiddlewareForPolicyNoopsWhenLimiterIsNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ok", MiddlewareForPolicy(nil, PerUserPolicy("test-policy", 1)), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	router.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected no-op limiter to allow request, got status %d", res.Code)
	}
	if got := res.Header().Get("X-RateLimit-Policy"); got != "" {
		t.Fatalf("expected no rate limit header for nil limiter, got %q", got)
	}
}

func TestMiddlewareForPolicyNoopsWhenLimitIsDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ok", MiddlewareForPolicy(New(nil), PerUserPolicy("test-policy", 0)), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	router.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected disabled policy to allow request, got status %d", res.Code)
	}
	if got := res.Header().Get("X-RateLimit-Policy"); got != "" {
		t.Fatalf("expected no rate limit header for disabled policy, got %q", got)
	}
}

func TestPolicyKeyIncludesPolicyName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", "user-123")

	policy := Policy{
		Name:    "audit-export",
		Limit:   10,
		Window:  time.Minute,
		KeyFunc: ByUser,
	}

	if got, want := policy.key(c), "rl:audit-export:user:user-123"; got != want {
		t.Fatalf("expected key %q, got %q", want, got)
	}
}

func TestByPolicyHelpersReturnCompleteKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", "user-123")

	if got, want := ByPolicyUser("rbac-sensitive-write")(c), "rl:rbac-sensitive-write:user:user-123"; got != want {
		t.Fatalf("expected key %q, got %q", want, got)
	}
}
