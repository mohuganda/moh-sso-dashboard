package keycloak

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateClientPreservesExplicitSecurityAndMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/realms/test/protocol/openid-connect/token" {
			json.NewEncoder(w).Encode(map[string]any{"access_token": "fixture", "expires_in": 300})
			return
		}
		if r.Method != "POST" || r.URL.Path != "/admin/realms/test/clients" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(404)
			return
		}
		var body CreateClientParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body.Enabled || body.StandardFlowEnabled || !body.ServiceAccountsEnabled || body.Attributes["portal.owner"] != "Team" || body.Attributes["ui.home"] != "" {
			t.Error("security flags or metadata overwritten")
		}
		w.Header().Set("Location", "/admin/realms/test/clients/fixture-id")
		w.WriteHeader(201)
	}))
	defer server.Close()
	client := NewAdminClient(server.URL, "test", "fixture-admin", "fixture-secret")
	if _, err := client.CreateClient(CreateClientParams{ClientID: "service-test", ServiceAccountsEnabled: true, Attributes: map[string]string{"portal.owner": "Team", "portal.onboarding.profile": "service"}}); err != nil {
		t.Fatal(err)
	}
}
