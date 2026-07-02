package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/version"
)

type versionEnvelope struct {
	Success bool              `json:"success"`
	Data    version.BuildInfo `json:"data"`
}

func TestHandleVersionDoesNotRequireDependencies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHealthHandler(nil, nil, nil, nil)
	router := gin.New()
	router.GET("/version", handler.HandleVersion)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/version", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var payload versionEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if !payload.Success || payload.Data.Service != "moh-sso-dashboard-backend" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestHandleLiveDoesNotDuplicateVersionPrefix(t *testing.T) {
	oldVersion := version.Version
	version.Version = "v1.2.3"
	defer func() { version.Version = oldVersion }()

	gin.SetMode(gin.TestMode)
	handler := NewHealthHandler(nil, nil, nil, nil)
	router := gin.New()
	router.GET("/health/live", handler.HandleLive)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	router.ServeHTTP(recorder, request)

	var payload struct {
		Data HealthResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Data.Version != "1.2.3" {
		t.Fatalf("version = %q, want 1.2.3", payload.Data.Version)
	}
	if payload.Data.Build == nil || payload.Data.Build.Version != "1.2.3" {
		t.Fatalf("unexpected build payload: %+v", payload.Data.Build)
	}
}
