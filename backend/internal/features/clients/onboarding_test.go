package clients

import (
	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/config"
	"github.com/moh-sso-dashboard/internal/keycloak"
	"github.com/moh-sso-dashboard/internal/model"
	"testing"
)

type onboardingRepo struct {
	Repository
	existing *model.Client
}

func (r onboardingRepo) ListClients() ([]model.Client, error) {
	if r.existing == nil {
		return nil, nil
	}
	return []model.Client{*r.existing}, nil
}
func (r onboardingRepo) GetClientByID(uuid.UUID) (*model.Client, error) { return r.existing, nil }
func TestOnboardingPreviewDetectsDriftAndProtectsExistingClients(t *testing.T) {
	spec := keycloak.OnboardingSpec{Version: 1, ClientID: "app-test", Name: "App", Owner: "Team", SupportEmail: "support@example.org", Profile: "public-web", Environment: "staging", RootURL: "https://app.example.org", RedirectURIs: []string{"https://app.example.org/callback"}, LogoutURIs: []string{"https://app.example.org/login"}, WebOrigins: []string{"https://app.example.org"}}
	s := &Service{repo: onboardingRepo{}, cfg: &config.Config{Environment: "staging", KeycloakRealm: "moh-staging", KeycloakExternalURL: "https://auth.example.org"}}
	first, client, err := s.previewOnboarding(spec)
	if err != nil {
		t.Fatal(err)
	}
	if first.Operation != "create" || client.Enabled || first.ReviewToken == "" {
		t.Fatal("incorrect preview")
	}
	spec.Owner = "Other team"
	second, _, err := s.previewOnboarding(spec)
	if err != nil || first.ReviewToken == second.ReviewToken {
		t.Fatal("review token did not track draft changes")
	}
	client.ID = uuid.NewString()
	client.Attributes["portal.onboarding.profile"] = "other"
	s.repo = onboardingRepo{existing: client}
	if _, _, err = s.previewOnboarding(spec); err == nil {
		t.Fatal("unmanaged client takeover allowed")
	}
	client.Attributes["portal.onboarding.profile"] = "public-web"
	client.Enabled = true
	if _, _, err = s.previewOnboarding(spec); err == nil {
		t.Fatal("active client update allowed")
	}
	client.Enabled = false
	third, _, err := s.previewOnboarding(spec)
	if err != nil || third.Operation != "update" || third.Existing == nil {
		t.Fatalf("managed update rejected: %v", err)
	}
}
