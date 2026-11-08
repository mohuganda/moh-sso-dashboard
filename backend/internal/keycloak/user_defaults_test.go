package keycloak

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/moh-sso-dashboard/internal/model"
)

func TestCreateHumanUserAssignsBaselineWithoutChangingRealmDefaults(t *testing.T) {
	for _, failure := range []string{"", "missing-role", "grant-denied"} {
		t.Run(failure, func(t *testing.T) {
			created, assigned := false, false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/realms/test/protocol/openid-connect/token":
					_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "fixture-token", "expires_in": 300})
				case "/admin/realms/test/roles/user":
					if failure == "missing-role" {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					_ = json.NewEncoder(w).Encode(RoleRep{ID: "user-role", Name: "user"})
				case "/admin/realms/test/users":
					if r.Method != http.MethodPost {
						t.Errorf("unexpected method %s", r.Method)
						w.WriteHeader(http.StatusMethodNotAllowed)
						return
					}
					var payload CreateUserRequest
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
						t.Error(err)
					}
					if values := payload.Attributes["user_type"]; len(values) != 1 || values[0] != "human" {
						t.Error("missing human account marker")
					}
					created = true
					w.Header().Set("Location", "/admin/realms/test/users/new-human")
					w.WriteHeader(http.StatusCreated)
				case "/admin/realms/test/users/new-human/role-mappings/realm":
					var roles []RoleRep
					if err := json.NewDecoder(r.Body).Decode(&roles); err != nil {
						t.Error(err)
					}
					if r.Method != http.MethodPost || len(roles) != 1 || roles[0].ID != "user-role" {
						t.Errorf("unexpected role assignment: %s %+v", r.Method, roles)
					}
					if failure == "grant-denied" {
						w.WriteHeader(http.StatusForbidden)
						return
					}
					assigned = true
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected request (realm defaults and other users must remain untouched): %s", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			client := NewAdminClient(server.URL, "test", "fixture-admin", "fixture-secret")
			id, err := client.CreateUser(&model.User{Username: "new-human", Email: "fixture@example.invalid"})
			if failure == "" {
				if err != nil || id != "new-human" || !created || !assigned {
					t.Fatalf("baseline not assigned: id=%q created=%t assigned=%t err=%v", id, created, assigned, err)
				}
			} else if failure == "missing-role" {
				if err == nil || created {
					t.Fatal("missing baseline role must fail before creating an account")
				}
			} else if err == nil || !strings.Contains(err.Error(), "standard role assignment failed") || id != "new-human" || !created || assigned {
				t.Fatalf("must report partial creation for recovery: id=%q err=%v", id, err)
			}
		})
	}
}
