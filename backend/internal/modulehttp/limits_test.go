package modulehttp

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/ratelimit"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type limitStore struct {
	ready   bool
	allowed bool
	err     error
	key     string
	calls   int
	limit   int
	window  time.Duration
}

func (s *limitStore) Available() bool { return s.ready }
func (s *limitStore) Allow(_ context.Context, key string, limit int, window time.Duration) (bool, int, time.Time, error) {
	s.key = key
	s.calls++
	s.limit = limit
	s.window = window
	return s.allowed, 2, time.Now().Add(time.Minute), s.err
}
func TestPerUserLimitRequiresVerifiedIdentityAndFailsClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, mode := range []string{"anonymous", "unavailable", "store-error", "limited", "allowed"} {
		store := &limitStore{ready: mode != "unavailable", allowed: mode == "allowed"}
		if mode == "store-error" {
			store.err = errors.New("private redis error")
		}
		engine := gin.New()
		group := engine.Group("/api/v1", func(c *gin.Context) {
			if mode != "anonymous" {
				c.Set(authz.ContextKey, authz.Context{UserID: "verified-subject"})
			}
		})
		NewRegistrar(group).Handle("GET /limit", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headers, err := PerUserLimit(store)(r, "audit-export", 10, time.Minute)
			for name, values := range headers {
				for _, value := range values {
					w.Header().Add(name, value)
				}
			}
			if mode == "limited" {
				var retry interface{ RetryAfter() time.Duration }
				if !errors.As(err, &retry) || retry.RetryAfter() <= 0 {
					t.Errorf("rate limit error is missing retry timing: %v", err)
				}
			}
			if err != nil {
				status := 503
				if v, ok := err.(interface{ StatusCode() int }); ok {
					status = v.StatusCode()
				}
				w.WriteHeader(status)
				return
			}
			w.WriteHeader(204)
		}))
		req := httptest.NewRequest("GET", "/api/v1/limit", nil)
		req.Header.Set("X-User-ID", "spoofed")
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		want := map[string]int{"anonymous": 401, "unavailable": 503, "store-error": 503, "limited": 429, "allowed": 204}[mode]
		if w.Code != want {
			t.Fatalf("%s: %d want %d", mode, w.Code, want)
		}
		if mode == "anonymous" || mode == "unavailable" {
			if store.calls != 0 {
				t.Fatal("unverified/unavailable limit accessed store")
			}
		} else if store.key != "rl:audit-export:user:verified-subject" || store.limit != 10 || store.window != time.Minute {
			t.Fatal("policy/identity changed", store)
		}
		if mode == "limited" || mode == "allowed" {
			if w.Header().Get("X-RateLimit-Policy") != "audit-export" || w.Header().Get("X-RateLimit-Limit") != "10" {
				t.Fatal(w.Header())
			}
		}
	}
	var missing *ratelimit.Limiter
	if missing.Available() {
		t.Fatal("typed nil limiter reports available")
	}
}
