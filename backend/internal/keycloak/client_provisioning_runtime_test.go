package keycloak

import (
	"encoding/json"
	"os"
	"testing"
)

// Invoked by the disposable Docker fixture, never by a production environment.
func TestOnboardingKeycloakRuntime(t *testing.T) {
	base := os.Getenv("ONBOARDING_KEYCLOAK_URL")
	if base == "" {
		t.Skip("requires disposable Keycloak fixture")
	}
	kc := NewAdminClient(base, "moh-realm", "fixture-provisioner", os.Getenv("ONBOARDING_KEYCLOAK_SECRET"))
	for _, profile := range []string{"public-web", "confidential-web", "service"} {
		t.Run(profile, func(t *testing.T) {
			spec := OnboardingSpec{Version: 1, ClientID: "runtime-" + profile, Name: "Runtime fixture", Owner: "Test team", SupportEmail: "fixture@example.invalid", Profile: profile, Environment: "production"}
			if profile != "service" {
				spec.RootURL = "https://app.example.invalid"
				spec.RedirectURIs = []string{spec.RootURL + "/callback"}
				spec.LogoutURIs = []string{spec.RootURL + "/login", spec.RootURL + "/signed-out"}
				spec.WebOrigins = []string{spec.RootURL}
			}
			desired, err := BuildOnboardingClient(spec, "production")
			if err != nil {
				t.Fatal(err)
			}
			raw, _ := json.Marshal(desired)
			var params CreateClientParams
			if err = json.Unmarshal(raw, &params); err != nil {
				t.Fatal(err)
			}
			desired.ID, err = kc.CreateClient(params)
			if err != nil {
				t.Fatal(err)
			}
			if err = kc.ReconcileClientProfile(desired); err != nil {
				t.Fatal(err)
			}
			// Retry and update exercise existing mappers, scope reconciliation and disabled state.
			updated := *desired
			updated.ProtocolMappers = nil
			if err = kc.UpdateClient(desired.ID, &updated); err != nil {
				t.Fatal(err)
			}
			if err = kc.ReconcileClientProfile(desired); err != nil {
				t.Fatal(err)
			}
			actual, err := kc.GetClientByID(desired.ID)
			if err != nil {
				t.Fatal(err)
			}
			if actual.Enabled || actual.PublicClient != (profile == "public-web") || actual.ServiceAccountsEnabled != (profile == "service") {
				t.Fatal("profile did not persist")
			}
			// An update must not clear a generated confidential client secret.
			if profile != "public-web" && actual.Secret == "" {
				t.Fatal("confidential client lost its secret")
			}
		})
	}
}
