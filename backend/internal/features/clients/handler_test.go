package clients

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestClientRoutesIncludeUpdateRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	RegisterProtectedRoutes(router.Group(""), &Handler{})

	for _, route := range router.Routes() {
		if route.Method == http.MethodPut && route.Path == "/clients/:id" {
			return
		}
	}

	t.Fatal("expected PUT /clients/:id to be registered")
}

func TestGetClientRejectsInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := &Handler{}
	router.GET("/clients/:id", handler.GetClient)

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/clients/not-a-uuid", nil)

	router.ServeHTTP(res, req)

	assertErrorCode(t, res, http.StatusBadRequest, "INVALID_UUID")
}

func TestToggleClientEnabledRejectsInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := &Handler{}
	router.PATCH("/clients/:id/toggle", handler.ToggleClientEnabled)

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/clients/not-a-uuid/toggle", nil)

	router.ServeHTTP(res, req)

	assertErrorCode(t, res, http.StatusBadRequest, "INVALID_UUID")
}

func assertErrorCode(t *testing.T, res *httptest.ResponseRecorder, status int, code string) {
	t.Helper()

	if res.Code != status {
		t.Fatalf("expected status %d, got %d: %s", status, res.Code, res.Body.String())
	}

	var body struct {
		Success bool `json:"success"`
		Error   struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if body.Success {
		t.Fatal("expected failure response")
	}
	if body.Error.Code != code {
		t.Fatalf("expected error code %q, got %q", code, body.Error.Code)
	}
}
