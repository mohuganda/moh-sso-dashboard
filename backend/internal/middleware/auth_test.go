package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
)

func TestRequirePermissionAllowsAuthorizedContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/documents", func(c *gin.Context) {
		c.Set(authz.ContextKey, authz.NewContext("user-1", []string{authz.RoleUser}, map[string][]string{
			authz.SystemIntegratedOutbreak: {authz.IntegratedOutbreakViewer},
		}))
	}, RequirePermission(authz.PermissionDocumentsRead), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/documents", nil)
	router.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d", http.StatusNoContent, res.Code)
	}
}

func TestRequirePermissionRejectsMissingPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/documents", func(c *gin.Context) {
		c.Set(authz.ContextKey, authz.NewContext("user-1", []string{authz.RoleUser}, nil))
	}, RequirePermission(authz.PermissionDocumentsWrite), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/documents", nil)
	router.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d", http.StatusForbidden, res.Code)
	}
}
