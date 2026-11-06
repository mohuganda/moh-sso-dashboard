package keycloak

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"slices"

	"github.com/moh-sso-dashboard/internal/model"
)

// ReconcileClientProfile changes only this client's scopes and named mappers.
// A failed provision remains disabled and can be retried from its onboarding draft.
func (c *KeyAdminClient) ReconcileClientProfile(client *model.Client) error {
	path := "clients/" + url.PathEscape(client.ID)
	get := func(path string, target any) error {
		response, err := c.Get(path)
		if err != nil {
			return err
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("read client configuration: HTTP %d", response.StatusCode)
		}
		return json.NewDecoder(response.Body).Decode(target)
	}
	write := func(method, path string, body any) error {
		var response *http.Response
		var err error
		switch method {
		case "PUT":
			response, err = c.Put(path, body)
		case "POST":
			response, err = c.Post(path, body)
		default:
			response, err = c.Delete(path)
		}
		if err != nil {
			return err
		}
		defer response.Body.Close()
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			return fmt.Errorf("configure client: HTTP %d", response.StatusCode)
		}
		return nil
	}
	type scope struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	var available []scope
	if err := get("client-scopes", &available); err != nil {
		return err
	}
	for kind, wanted := range map[string][]string{"default": client.DefaultClientScopes, "optional": client.OptionalClientScopes} {
		var assigned []scope
		if err := get(path+"/"+kind+"-client-scopes", &assigned); err != nil {
			return err
		}
		for _, current := range assigned {
			if !slices.Contains(wanted, current.Name) {
				if err := write("DELETE", path+"/"+kind+"-client-scopes/"+url.PathEscape(current.ID), nil); err != nil {
					return err
				}
			}
		}
		for _, name := range wanted {
			found := false
			for _, current := range assigned {
				if current.Name == name {
					found = true
				}
			}
			if found {
				continue
			}
			id := ""
			for _, item := range available {
				if item.Name == name {
					id = item.ID
				}
			}
			if id == "" {
				return fmt.Errorf("required client scope is missing: %s", name)
			}
			if err := write("PUT", path+"/"+kind+"-client-scopes/"+url.PathEscape(id), nil); err != nil {
				return err
			}
		}
	}
	var mappers []model.ProtocolMapper
	if err := get(path+"/protocol-mappers/models", &mappers); err != nil {
		return err
	}
	for _, desired := range client.ProtocolMappers {
		method, target := "POST", path+"/protocol-mappers/models"
		for _, current := range mappers {
			if current.Name == desired.Name {
				desired.ID = current.ID
				method = "PUT"
				target += "/" + url.PathEscape(current.ID)
				break
			}
		}
		if err := write(method, target, desired); err != nil {
			return err
		}
	}
	actual, err := c.GetClientByID(client.ID)
	if err != nil {
		return err
	}
	if actual.Enabled || actual.PublicClient != client.PublicClient || actual.StandardFlowEnabled != client.StandardFlow || actual.ImplicitFlowEnabled || actual.DirectAccessGrantsEnabled || actual.ServiceAccountsEnabled != client.ServiceAccounts || actual.FullScopeAllowed || !reflect.DeepEqual(actual.RedirectURIs, client.RedirectUris) || !reflect.DeepEqual(actual.WebOrigins, client.WebOrigins) {
		return fmt.Errorf("client security verification failed; leave client disabled")
	}
	for kind, wanted := range map[string][]string{"default": client.DefaultClientScopes, "optional": client.OptionalClientScopes} {
		actualScopes := actual.DefaultClientScopes
		if kind == "optional" {
			actualScopes = actual.OptionalClientScopes
		}
		if len(actualScopes) != len(wanted) {
			return fmt.Errorf("client scope verification failed")
		}
		for _, scope := range wanted {
			if !slices.Contains(actualScopes, scope) {
				return fmt.Errorf("client scope verification failed")
			}
		}
	}
	for _, desired := range client.ProtocolMappers {
		found := false
		for _, mapper := range actual.ProtocolMappers {
			if mapper.Name == desired.Name && mapper.ProtocolMapper == desired.ProtocolMapper {
				found = true
				for key, value := range desired.Config {
					if mapper.Config[key] != value {
						return fmt.Errorf("client mapper verification failed")
					}
				}
			}
		}
		if !found {
			return fmt.Errorf("client mapper verification failed")
		}
	}
	for key, value := range client.Attributes {
		if actual.Attributes[key] != value {
			return fmt.Errorf("client attribute verification failed: %s", key)
		}
	}
	return nil
}
