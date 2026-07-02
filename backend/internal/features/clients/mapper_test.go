package clients

import (
	"testing"

	"github.com/moh-sso-dashboard/internal/model"
)

func TestFilterAccessibleClientsReturnsAssignedApplicationClients(t *testing.T) {
	clients := []model.Client{
		{
			ClientID:   "dashboard-web",
			Name:       "Portal",
			Attributes: map[string]string{"ui.icon": "dashboard", "portal.system": "true"},
		},
		{
			ClientID:   "outbreak-management",
			Name:       "Outbreak Management",
			Attributes: map[string]string{"ui.icon": "outbreak", "portal.system": "true"},
		},
		{
			ClientID:   "data-statistics",
			Name:       "Data & Statistics",
			Attributes: map[string]string{"ui.icon": "reporting", "portal.system": "true"},
		},
		{
			ClientID:   "dashboard-admin",
			Name:       "Technical admin client",
			Attributes: map[string]string{"ui.icon": "settings", "portal.system": "true"},
		},
	}

	filtered := filterAccessibleClients(
		clients,
		map[string][]string{
			"dashboard-web":       {"dashboard-web_access", "portal_user"},
			"outbreak-management": {"outbreak-management_access", "viewer"},
			"data-statistics":     {"data-statistics_access", "report_viewer"},
		},
		false,
	)

	got := make([]string, 0, len(filtered))
	for _, client := range filtered {
		got = append(got, client.ClientID)
	}

	want := []string{"dashboard-web", "outbreak-management", "data-statistics"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}

func TestFilterAccessibleClientsSkipsClientsWithoutUiMetadata(t *testing.T) {
	filtered := filterAccessibleClients(
		[]model.Client{
			{
				ClientID:   "outbreak-management",
				Name:       "Outbreak Management",
				Attributes: map[string]string{"ui.icon": "outbreak", "portal.system": "true"},
			},
			{
				ClientID: "technical-client",
				Name:     "Technical",
			},
		},
		map[string][]string{
			"outbreak-management": {"outbreak-management_access"},
			"technical-client":    {"technical-client_access"},
		},
		false,
	)

	if len(filtered) != 1 || filtered[0].ClientID != "outbreak-management" {
		t.Fatalf("expected only Outbreak Management client, got %v", filtered)
	}
}
