package geojson

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetGeoJSONServesAllowlistedFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	basePath := t.TempDir()
	body := `{"type":"FeatureCollection","features":[]}`
	if err := os.WriteFile(filepath.Join(basePath, "uganda_districts.geojson"), []byte(body), 0o600); err != nil {
		t.Fatalf("write test geojson: %v", err)
	}

	router := gin.New()
	router.GET("/geojson/:name", NewHandler(basePath).GetGeoJSON)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/geojson/geo_districts", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if contentType := recorder.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "application/geo+json") {
		t.Fatalf("expected geojson content type, got %q", contentType)
	}
	if got := recorder.Header().Get("Cache-Control"); got != "public, max-age=86400" {
		t.Fatalf("expected cache header, got %q", got)
	}
	if got := recorder.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected nosniff header, got %q", got)
	}
	if strings.TrimSpace(recorder.Body.String()) != body {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}

func TestGetGeoJSONRejectsUnknownName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/geojson/:name", NewHandler(t.TempDir()).GetGeoJSON)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/geojson/not_allowed", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
	assertAPIErrorCode(t, recorder.Body.Bytes(), "NOT_FOUND")
}

func TestGetGeoJSONMissingFileReturnsSafeError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/geojson/:name", NewHandler(t.TempDir()).GetGeoJSON)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/geojson/geo_subcounties", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
	assertAPIErrorCode(t, recorder.Body.Bytes(), "GEOJSON_LOAD_FAILED")
	if strings.Contains(recorder.Body.String(), "uganda_subcounties.geojson") {
		t.Fatalf("response leaked internal filename: %s", recorder.Body.String())
	}
}

func assertAPIErrorCode(t *testing.T, payload []byte, expected string) {
	t.Helper()

	var response struct {
		Success bool `json:"success"`
		Error   struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatalf("decode error response: %v; payload=%s", err, string(payload))
	}
	if response.Success {
		t.Fatalf("expected unsuccessful response")
	}
	if response.Error.Code != expected {
		t.Fatalf("expected error code %q, got %q", expected, response.Error.Code)
	}
}
