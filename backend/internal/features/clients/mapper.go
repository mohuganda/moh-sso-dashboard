package clients

import (
	"github.com/moh-sso-dashboard/internal/keycloak"
	"github.com/moh-sso-dashboard/internal/model"
)

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

func toClientResponse(client *model.Client) ClientResponse {
	if client == nil {
		return ClientResponse{}
	}

	return ClientResponse{
		ID:                                 client.ID,
		ClientID:                           client.ClientID,
		Name:                               client.Name,
		Description:                        client.Description,
		RootURL:                            client.RootURL,
		BaseURL:                            client.BaseURL,
		AdminURL:                           client.AdminURL,
		SurrogateAuthRequired:              client.SurrogateAuthRequired,
		Enabled:                            client.Enabled,
		AlwaysDisplayInConsole:             client.AlwaysDisplayInConsole,
		ClientAuthenticatorType:            client.ClientAuthenticatorType,
		RedirectUris:                       client.RedirectUris,
		WebOrigins:                         client.WebOrigins,
		NotBefore:                          client.NotBefore,
		BearerOnly:                         client.BearerOnly,
		ConsentRequired:                    client.ConsentRequired,
		StandardFlow:                       client.StandardFlow,
		ImplicitFlow:                       client.ImplicitFlow,
		DirectAccess:                       client.DirectAccess,
		ServiceAccounts:                    client.ServiceAccounts,
		PublicClient:                       client.PublicClient,
		FrontChannelLogout:                 client.FrontChannelLogout,
		Protocol:                           client.Protocol,
		FullScopeAllowed:                   client.FullScopeAllowed,
		NodeReRegistrationTimeout:          client.NodeReRegistrationTimeout,
		DefaultClientScopes:                client.DefaultClientScopes,
		OptionalClientScopes:               client.OptionalClientScopes,
		Access:                             client.Access,
		ClientTemplate:                     client.ClientTemplate,
		UseTemplateConfig:                  client.UseTemplateConfig,
		RootClientRealm:                    client.RootClientRealm,
		AuthorizationServicesEnabled:       client.AuthorizationServicesEnabled,
		ProtocolMappers:                    toProtocolMapperResponses(client.ProtocolMappers),
		DefaultRoles:                       client.DefaultRoles,
		AuthorizationURL:                   client.AuthorizationURL,
		TlsRequired:                        client.TlsRequired,
		Attributes:                         client.Attributes,
		AuthenticationFlowBindingOverrides: client.AuthenticationFlowBindingOverrides,
		RegisteredNodes:                    client.RegisteredNodes,
	}
}

func toClientResponses(clients []model.Client) []ClientResponse {
	out := make([]ClientResponse, 0, len(clients))
	for i := range clients {
		out = append(out, toClientResponse(&clients[i]))
	}
	return out
}

func toProtocolMapperResponses(items []model.ProtocolMapper) []ProtocolMapper {
	out := make([]ProtocolMapper, 0, len(items))
	for _, item := range items {
		out = append(out, ProtocolMapper{
			ID:              item.ID,
			Name:            item.Name,
			Protocol:        item.Protocol,
			ProtocolMapper:  item.ProtocolMapper,
			ConsentRequired: item.ConsentRequired,
			ConsentText:     item.ConsentText,
			Config:          item.Config,
		})
	}
	return out
}

func toClientRoleResponse(role keycloak.ClientRoleRep) ClientRoleResponse {
	return ClientRoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
	}
}

func toClientRoleResponses(roles []keycloak.ClientRoleRep) []ClientRoleResponse {
	out := make([]ClientRoleResponse, 0, len(roles))
	for _, role := range roles {
		out = append(out, toClientRoleResponse(role))
	}
	return out
}
