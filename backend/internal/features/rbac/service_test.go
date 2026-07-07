package rbac

import (
	"context"
	"encoding/json"
	"testing"
)

func TestParseImportSeedAcceptsJSONSeed(t *testing.T) {
	payload := json.RawMessage(`{
		"systems": [
			{
				"clientId": "test-system",
				"displayName": "Test System",
				"roles": [{"name": "viewer"}]
			}
		],
		"realmRoles": [{"name": "admin"}]
	}`)

	seed, err := parseImportSeed(payload)
	if err != nil {
		t.Fatalf("parseImportSeed returned error: %v", err)
	}
	if got := seed.Systems[0].ClientID; got != "test-system" {
		t.Fatalf("expected clientId test-system, got %q", got)
	}
	if got := seed.RealmRoles[0].Name; got != "admin" {
		t.Fatalf("expected realm role admin, got %q", got)
	}
}

func TestParseImportSeedAcceptsJSONStringYAMLSeed(t *testing.T) {
	yamlPayload := `systems:
  - clientId: test-system
    displayName: Test System
    roles:
      - name: viewer
realmRoles:
  - name: admin
`
	payloadBytes, err := json.Marshal(yamlPayload)
	if err != nil {
		t.Fatalf("marshal YAML string: %v", err)
	}

	seed, err := parseImportSeed(payloadBytes)
	if err != nil {
		t.Fatalf("parseImportSeed returned error: %v", err)
	}
	if got := seed.Systems[0].ClientID; got != "test-system" {
		t.Fatalf("expected clientId test-system, got %q", got)
	}
}

func TestPreviewChangeFlagsHighRiskOperations(t *testing.T) {
	service := NewService(nil)

	preview, err := service.PreviewChange(context.Background(), ChangePreviewRequest{
		Action:       "delete-role",
		ResourceType: "system-role",
		ResourceID:   "role-id",
	})
	if err != nil {
		t.Fatalf("PreviewChange returned error: %v", err)
	}
	if !preview.HighRisk {
		t.Fatal("expected delete-role to be high risk")
	}
	if preview.RiskLevel != "high" {
		t.Fatalf("expected high risk level, got %q", preview.RiskLevel)
	}
}

func TestNormalizeDecisionSupportsPromptAliases(t *testing.T) {
	cases := map[string]string{
		"approve": "approved",
		"reject":  "rejected",
		"cancel":  "cancelled",
		"apply":   "applied",
	}
	for input, want := range cases {
		if got := normalizeDecision(input); got != want {
			t.Fatalf("normalizeDecision(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestValidateSystemLaunchURLAcceptsPortalAndAppPaths(t *testing.T) {
	for _, value := range []string{"/portal/apps/dwh/surveillance", "/apps/dwh/surveillance", "https://portal.example.org/apps/dwh"} {
		if err := validateSystemLaunchURL(value); err != nil {
			t.Fatalf("validateSystemLaunchURL(%q) returned error: %v", value, err)
		}
	}
}

func TestValidateOptionalURLAcceptsRootRelativeSupportLinks(t *testing.T) {
	for _, value := range []string{"/support/surveillance", "/docs/systems/surveillance", "https://docs.example.org/surveillance"} {
		if err := validateOptionalURL("supportUrl", value); err != nil {
			t.Fatalf("validateOptionalURL(%q) returned error: %v", value, err)
		}
	}
}

func TestKeycloakClientURLResolvesPortalRelativeLaunchURL(t *testing.T) {
	service := NewService(nil)
	service.SetFrontendBaseURL("http://localhost:3000/portal")

	got := service.keycloakClientURL("/portal/apps/outbreak-management")
	want := "http://localhost:3000/portal/apps/outbreak-management"
	if got != want {
		t.Fatalf("keycloakClientURL returned %q, want %q", got, want)
	}
}

func TestKeycloakClientURLResolvesAppPathUnderPortalBase(t *testing.T) {
	service := NewService(nil)
	service.SetFrontendBaseURL("http://localhost:3000/portal")

	got := service.keycloakClientURL("/apps/dwh/surveillance")
	want := "http://localhost:3000/portal/apps/dwh/surveillance"
	if got != want {
		t.Fatalf("keycloakClientURL returned %q, want %q", got, want)
	}
}

func TestKeycloakClientURLKeepsAbsoluteLaunchURL(t *testing.T) {
	service := NewService(nil)
	service.SetFrontendBaseURL("http://localhost:3000/portal")

	got := service.keycloakClientURL("https://systems.health.go.ug/outbreak")
	want := "https://systems.health.go.ug/outbreak"
	if got != want {
		t.Fatalf("keycloakClientURL returned %q, want %q", got, want)
	}
}

func TestKeycloakClientURLReturnsEmptyWithoutAbsoluteFrontendBase(t *testing.T) {
	service := NewService(nil)
	service.SetFrontendBaseURL("/portal")

	if got := service.keycloakClientURL("/portal/apps/outbreak-management"); got != "" {
		t.Fatalf("keycloakClientURL returned %q, want empty string", got)
	}
}

func TestParseRealmExportRequiresExplicitPortalEnrollment(t *testing.T) {
	payload := []byte(`{
		"roles":{"realm":[],"client":{"portal-app":[{"name":"portal-app_access"}],"technical-app":[{"name":"technical-app_access"}]}},
		"clients":[
			{"clientId":"portal-app","name":"Portal App","enabled":true,"attributes":{"portal.system":"true","ui.systemType":"external","ui.launchMode":"new_tab","ui.launchUrl":"https://example.org/app"}},
			{"clientId":"technical-app","name":"Technical App","enabled":true,"attributes":{"ui.icon":"application"}}
		]
	}`)
	discovered, err := parseRealmExport(payload)
	if err != nil {
		t.Fatalf("parseRealmExport returned error: %v", err)
	}
	if len(discovered.Systems) != 1 || discovered.Systems[0].ClientID != "portal-app" {
		t.Fatalf("expected only explicitly enrolled portal app, got %+v", discovered.Systems)
	}
	if discovered.Systems[0].LaunchMode != "new_tab" || discovered.Systems[0].SystemType != "external" {
		t.Fatalf("expected explicit external behavior, got %+v", discovered.Systems[0])
	}
}

func TestParseRealmExportKeepsKnownLegacySystem(t *testing.T) {
	payload := []byte(`{
		"roles":{"realm":[],"client":{"legacy-app":[{"name":"legacy-app_access"}]}},
		"clients":[{"clientId":"legacy-app","name":"Legacy App","enabled":true,"attributes":{"ui.home":"/portal/apps/legacy"}}]
	}`)
	discovered, err := parseRealmExport(payload, map[string]bool{"legacy-app": true})
	if err != nil {
		t.Fatalf("parseRealmExport returned error: %v", err)
	}
	if len(discovered.Systems) != 1 || discovered.Systems[0].ClientID != "legacy-app" {
		t.Fatalf("expected known legacy app, got %+v", discovered.Systems)
	}
}
