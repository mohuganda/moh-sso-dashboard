package surveillance

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUnavailableObservatoryReturns503AndAutomaticallyEnablesAfterRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var available atomic.Bool
	handler := &Handler{}
	handler.SetAvailabilityCheck(available.Load)
	router := gin.New()
	routes := router.Group("/surveillance", handler.RequireAvailable())
	calls := 0
	routes.GET("/regions", func(c *gin.Context) { calls++; c.Status(http.StatusOK) })
	for _, ready := range []bool{false, true, false, true} {
		available.Store(ready)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/surveillance/regions", nil))
		if ready {
			if response.Code != http.StatusOK {
				t.Fatalf("recovered request: %d", response.Code)
			}
		} else if response.Code != http.StatusServiceUnavailable || response.Header().Get("Retry-After") != "15" || !strings.Contains(response.Body.String(), "OBSERVATORY_UNAVAILABLE") {
			t.Fatalf("unavailable request: %d %s", response.Code, response.Body.String())
		}
	}
	if calls != 2 {
		t.Fatalf("unavailable requests reached the data handler: %d", calls)
	}
}
