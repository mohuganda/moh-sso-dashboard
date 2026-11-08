package users

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/http/apierror"
	models "github.com/moh-sso-dashboard/internal/model"
)

type statusRepository struct {
	Repository
	user   models.User
	writes int
}

func (r *statusRepository) GetUserByID(uuid.UUID) (*models.User, error) { return &r.user, nil }
func (r *statusRepository) UpdateUser(user *models.User) error {
	r.writes++
	r.user = *user
	return nil
}
func (r *statusRepository) ToggleUserEnabled(_ context.Context, _ string, enabled bool) error {
	r.writes++
	r.user.Enabled = enabled
	return nil
}

func TestCannotDisableSelfThroughEitherServicePath(t *testing.T) {
	actor := uuid.New()
	disabled := false
	for _, path := range []string{"status", "profile"} {
		t.Run(path, func(t *testing.T) {
			repo := &statusRepository{user: models.User{ID: actor.String(), Enabled: true}}
			service := NewService(repo, nil)
			var err error
			if path == "status" {
				err = service.SetUserEnabled(context.Background(), actor, false, actor)
			} else {
				_, err = service.UpdateUser(context.Background(), UpdateUserRequest{ID: actor.String(), Enabled: &disabled}, actor)
			}
			var apiErr *apierror.APIError
			if !errors.As(err, &apiErr) || apiErr.Code != "SELF_DISABLE_FORBIDDEN" || apiErr.HTTPStatus != http.StatusForbidden {
				t.Fatalf("expected self-disable rejection, got %v", err)
			}
			if repo.writes != 0 || !repo.user.Enabled {
				t.Fatal("self-disable changed the user")
			}
		})
	}
}

func TestStatusHandlerRequiresExplicitBooleanAndBlocksSelf(t *testing.T) {
	gin.SetMode(gin.TestMode)
	actor, other := uuid.New(), uuid.New()
	for _, tc := range []struct {
		name, body     string
		target         uuid.UUID
		status, writes int
	}{
		{"missing", `{}`, other, 400, 0},
		{"null", `{"enabled":null}`, other, 400, 0},
		{"wrong type", `{"enabled":"false"}`, other, 400, 0},
		{"self disable", `{"enabled":false}`, actor, 403, 0},
		{"self enable", `{"enabled":true}`, actor, 204, 0},
		{"other disable", `{"enabled":false}`, other, 204, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &statusRepository{user: models.User{ID: tc.target.String(), Enabled: true}}
			handler := NewHandler(NewService(repo, nil), nil, nil)
			router := gin.New()
			router.PATCH("/users/:id/enabled", func(c *gin.Context) { c.Set("user_id", actor.String()); handler.SetUserEnabled(c) })
			request := httptest.NewRequest(http.MethodPatch, "/users/"+tc.target.String()+"/enabled", strings.NewReader(tc.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != tc.status || repo.writes != tc.writes {
				t.Fatalf("status=%d writes=%d body=%s", response.Code, repo.writes, response.Body.String())
			}
		})
	}
}

func TestEditingOwnProfilePreservesEnabledState(t *testing.T) {
	actor := uuid.New()
	repo := &statusRepository{user: models.User{ID: actor.String(), Enabled: true, Username: "admin", Email: "admin@example.invalid"}}
	user, err := NewService(repo, nil).UpdateUser(context.Background(), UpdateUserRequest{ID: actor.String(), FirstName: "Updated"}, actor)
	if err != nil || user == nil || !user.Enabled || repo.writes != 1 {
		t.Fatalf("profile edit failed: user=%v err=%v writes=%d", user, err, repo.writes)
	}
}
