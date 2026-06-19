package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/observability"
)

func TestRequestContextPropagatesRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestContext())
	router.GET("/ping", func(c *gin.Context) {
		response.OK(c, http.StatusOK, gin.H{"ok": true})
	})

	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	request.Header.Set(observability.RequestIDHeader, "req-123")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if got := recorder.Header().Get(observability.RequestIDHeader); got != "req-123" {
		t.Fatalf("expected response request id header req-123, got %q", got)
	}
	if body := recorder.Body.String(); !strings.Contains(body, `"requestId":"req-123"`) {
		t.Fatalf("expected response meta request id, got %s", body)
	}
}
