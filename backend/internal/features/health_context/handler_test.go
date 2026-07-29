package health_context

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestRequireAccessibleSelectionRequiresHeaderForAssignedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	contextID := uuid.New()
	repository := &fakeRepository{
		nodes: map[uuid.UUID]Node{
			contextID: {
				ID: contextID, Code: "UG", Name: "Uganda",
				ContextType: ContextNational, Enabled: true,
			},
		},
		effective: []EffectiveContext{{
			Node: Node{
				ID: contextID, Code: "UG", Name: "Uganda",
				ContextType: ContextNational, Enabled: true,
			},
			ScopeMode: ScopeNodeOnly,
		}},
		descendants: map[[2]uuid.UUID]bool{},
	}
	handler := NewHandler(NewService(repository), nil)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Next()
	})
	router.GET("/documents", handler.RequireAccessibleSelection(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/documents", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "HEALTH_CONTEXT_REQUIRED") {
		t.Fatalf("expected health context error, got %s", recorder.Body.String())
	}
}

func TestRequireAccessibleSelectionAllowsUnassignedUserWithoutHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(NewService(&fakeRepository{
		nodes:       map[uuid.UUID]Node{},
		descendants: map[[2]uuid.UUID]bool{},
	}), nil)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Next()
	})
	router.GET("/documents", handler.RequireAccessibleSelection(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/documents", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d: %s", http.StatusNoContent, recorder.Code, recorder.Body.String())
	}
}

func TestRequireAccessibleSelectionStoresValidatedScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	contextID := uuid.New()
	repository := &fakeRepository{
		nodes: map[uuid.UUID]Node{
			contextID: {
				ID: contextID, Code: "KLA", Name: "Kampala",
				ContextType: ContextDistrict, Enabled: true,
			},
		},
		effective: []EffectiveContext{{
			Node: Node{
				ID: contextID, Code: "KLA", Name: "Kampala",
				ContextType: ContextDistrict, Enabled: true,
			},
			ScopeMode: ScopeNodeAndDescendants,
		}},
		descendants: map[[2]uuid.UUID]bool{},
	}
	handler := NewHandler(NewService(repository), nil)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Next()
	})
	router.GET("/documents", handler.RequireAccessibleSelection(), func(c *gin.Context) {
		if c.MustGet("health_context_id") != contextID {
			c.Status(http.StatusInternalServerError)
			return
		}
		if c.MustGet("health_context_scope_mode") != ScopeNodeAndDescendants {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/documents", nil)
	request.Header.Set("X-Health-Context-ID", contextID.String())
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d: %s", http.StatusNoContent, recorder.Code, recorder.Body.String())
	}
}
