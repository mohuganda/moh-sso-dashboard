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
