package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestObservatoryOutageDegradesHealthWithoutFailingCoreReadiness(t *testing.T) {
	gin.SetMode(gin.TestMode)
	healthy := func(context.Context) error { return nil }
	handler := NewHealthHandler(healthy, healthy, healthy, nil)
	handler.ObservatoryReady = func() bool { return false }
	router := gin.New()
	router.GET("/health", handler.HandleHealth)
	router.GET("/health/ready", handler.HandleReady)
	for _, path := range []string{"/health", "/health/ready"} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", path, recorder.Code, recorder.Body.String())
		}
		if path == "/health" && (!strings.Contains(recorder.Body.String(), `"status":"degraded"`) || !strings.Contains(recorder.Body.String(), `"observatory":"unavailable"`)) {
			t.Fatalf("missing degraded Observatory status: %s", recorder.Body.String())
		}
	}
	handler.DBCheck = func(context.Context) error { return errors.New("primary unavailable") }
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatal("primary database must remain required")
	}
}
