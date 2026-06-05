package clients

import "github.com/moh-sso-dashboard/internal/model"

func filterAccessibleClients(
	clients []model.Client,
	clientRoles map[string][]string,
	isAdmin bool,
) []model.Client {
	filtered := make([]model.Client, 0)

	for _, client := range clients {
		if client.Attributes == nil || client.Attributes["ui.icon"] == "" {
			continue
		}

		if isAdmin {
			filtered = append(filtered, client)
			continue
		}

		expectedRole := client.ClientID + "_access"
		for _, role := range clientRoles[client.ClientID] {
			if role == expectedRole {
				filtered = append(filtered, client)
				break
			}
		}
	}

	return filtered
}
