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
			Attributes: map[string]string{"ui.icon": "dashboard"},
		},
		{
			ClientID:   "integrated-outbreak-system",
			Name:       "IOS",
			Attributes: map[string]string{"ui.icon": "outbreak"},
		},
		{
			ClientID:   "report-browser",
			Name:       "Reports",
			Attributes: map[string]string{"ui.icon": "reporting"},
		},
		{
			ClientID:   "dashboard-admin",
			Name:       "Technical admin client",
			Attributes: map[string]string{"ui.icon": "settings"},
		},
	}

	filtered := filterAccessibleClients(
		clients,
		map[string][]string{
			"dashboard-web":              {"dashboard-web_access", "portal_user"},
			"integrated-outbreak-system": {"integrated-outbreak-system_access", "viewer"},
			"report-browser":             {"report-browser_access", "report_viewer"},
		},
		false,
	)

	got := make([]string, 0, len(filtered))
	for _, client := range filtered {
		got = append(got, client.ClientID)
	}

	want := []string{"dashboard-web", "integrated-outbreak-system", "report-browser"}
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
				ClientID:   "integrated-outbreak-system",
				Name:       "IOS",
				Attributes: map[string]string{"ui.icon": "outbreak"},
			},
			{
				ClientID: "technical-client",
				Name:     "Technical",
			},
		},
		map[string][]string{
			"integrated-outbreak-system": {"integrated-outbreak-system_access"},
			"technical-client":           {"technical-client_access"},
		},
		false,
	)

	if len(filtered) != 1 || filtered[0].ClientID != "integrated-outbreak-system" {
		t.Fatalf("expected only IOS client, got %v", filtered)
	}
}
