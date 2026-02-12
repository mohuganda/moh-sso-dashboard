package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/utils"
)

type ClientInfo struct {
	ID                        string            `json:"id"`                        // Internal unique ID (UUID)
	ClientID                  string            `json:"clientId"`                  // The client ID used for OAuth/OIDC protocol
	Name                      string            `json:"name"`                      // Display name for the client
	Description               string            `json:"description,"`              // Detailed description
	BaseURL                   string            `json:"baseUrl"`                   // Base URL for the client application
	RootURL                   string            `json:"rootUrl"`                   // Root URL for relative paths
	AdminURL                  string            `json:"adminUrl"`                  // URL to the client's admin console
	Enabled                   bool              `json:"enabled"`                   // Whether the client is active
	PublicClient              bool              `json:"publicClient"`              // True for browser-based apps without a secret
	BearerOnly                bool              `json:"bearerOnly"`                // True if the client only accepts bearer tokens
	Protocol                  string            `json:"protocol"`                  // Protocol used: "openid-connect" or "saml"
	Secret                    string            `json:"secret"`                    // Shared secret for confidential clients (if not publicClient)
	ClientAuthenticatorType   string            `json:"clientAuthenticatorType"`   // e.g., "client-secret" or "client-jwt"
	RedirectURIs              []string          `json:"redirectUris"`              // Valid redirect URIs after successful authentication
	WebOrigins                []string          `json:"webOrigins"`                // List of allowed CORS origins
	StandardFlowEnabled       bool              `json:"standardFlowEnabled"`       // Authorization Code Flow
	ImplicitFlowEnabled       bool              `json:"implicitFlowEnabled"`       // Implicit Flow (legacy)
	DirectAccessGrantsEnabled bool              `json:"directAccessGrantsEnabled"` // Resource Owner Password Credentials Grant
	ServiceAccountsEnabled    bool              `json:"serviceAccountsEnabled"`    // Enables a service account for machine-to-machine
	FullScopeAllowed          bool              `json:"fullScopeAllowed"`          // If true, all realm roles and scopes are granted
	DefaultClientScopes       []string          `json:"defaultClientScopes"`       // List of required scopes (client scopes)
	OptionalClientScopes      []string          `json:"optionalClientScopes"`      // List of optional scopes
	Attributes                map[string]string `json:"attributes"`                // Custom key/value client settings
	ProtocolMappers           []ProtocolMapper  `json:"protocolMappers"`           // Defines how claims are mapped into tokens
}

type ProtocolMapper struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Protocol       string            `json:"protocol"`
	ProtocolMapper string            `json:"protocolMapper"`
	Config         map[string]string `json:"config"`
}

type CreateClientParams struct {
	ClientID                  string            `json:"clientId"`
	Name                      string            `json:"name"`
	Description               string            `json:"description"`
	BaseURL                   string            `json:"baseUrl"`
	RootURL                   string            `json:"rootUrl"`
	RedirectURIs              []string          `json:"redirectUris"`
	WebOrigins                []string          `json:"webOrigins"`
	PublicClient              bool              `json:"publicClient"`
	Protocol                  string            `json:"protocol"`
	StandardFlowEnabled       bool              `json:"standardFlowEnabled"`
	ImplicitFlowEnabled       bool              `json:"implicitFlowEnabled"`
	DirectAccessGrantsEnabled bool              `json:"directAccessGrantsEnabled"`
	ServiceAccountsEnabled    bool              `json:"serviceAccountsEnabled"`
	Enabled                   bool              `json:"enabled"`
	Attributes                map[string]string `json:"attributes"`
}

type KeycloakUser struct {
	ID               string              `json:"id"`
	Username         string              `json:"username"`
	Email            string              `json:"email"`
	FirstName        string              `json:"firstName"`
	LastName         string              `json:"lastName"`
	Enabled          bool                `json:"enabled"`
	EmailVerified    bool                `json:"emailVerified"`
	RequiredActions  []string            `json:"requiredActions"`
	CreatedTimestamp int64               `json:"createdTimestamp"`
	Attributes       map[string][]string `json:"attributes"`
}

type UserInfo struct {
	ID                  string              `json:"id"`
	Username            string              `json:"username"`
	Email               string              `json:"email"`
	DisplayName         string              `json:"displayName"`
	FirstName           string              `json:"firstName"`
	LastName            string              `json:"lastName"`
	Enabled             bool                `json:"enabled"`
	AccountStatus       string              `json:"accountStatus"`
	EmailVerified       bool                `json:"emailVerified"`
	NeverLoggedIn       bool                `json:"neverLoggedIn"`
	RealmRoles          []string            `json:"realmRoles"`
	ClientRoles         map[string][]string `json:"clientRoles"`
	RequiredActions     []string            `json:"requiredActions"`
	FailedLoginAttempts int                 `json:"failedLoginAttempts"`
	TemporarilyLocked   bool                `json:"temporarilyLocked"`
	LastLoginAt         *time.Time          `json:"lastLoginAt"`
	LastLoginIP         string              `json:"lastLoginIp"`
	Source              string              `json:"source"`
	ImportedAt          *time.Time          `json:"importedAt"`
	Notes               string              `json:"notes"`
	CreatedAt           time.Time           `json:"createdAt"`
}

type CreateUserRequest struct {
	Username      string `json:"username"`
	Email         string `json:"email"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	Enabled       bool   `json:"enabled"`
	EmailVerified bool   `json:"emailVerified"`
	Credentials   []struct {
		Type      string `json:"type"`
		Value     string `json:"value"`
		Temporary bool   `json:"temporary"`
	} `json:"credentials"`
}

type UserRep struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Enabled  bool   `json:"enabled"`
}

type CreateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Composite   bool   `json:"composite"`
	ClientRole  bool   `json:"clientRole"`
}

type UserClientRoleAssignment struct {
	Id         string   `json:"id"`
	ClientID   string   `json:"clientId"`
	ClientName string   `json:"clientName"`
	Roles      []string `json:"roles"`
}

type ClientRoleRep struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ClientRoleRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type RoleRep struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ClientScope struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
}

type ClientTemplate string

const (
	WebClient     ClientTemplate = "web"
	ServiceClient ClientTemplate = "service"
	AdminClient   ClientTemplate = "admin"
)

var defaultClientScopes = []string{
	"profile",
	"email",
	"roles",
}

func (c *Client) EnsureRealmExists(realmName string) error {
	res, err := c.Get(fmt.Sprintf("admin/realms/%s", realmName))
	if err == nil && res.StatusCode == http.StatusOK {
		res.Body.Close()
		return nil
	}

	if res != nil {
		res.Body.Close()
	}

	payload := map[string]any{
		"realm":   realmName,
		"enabled": true,
	}
	res, err = c.Post("admin/realms", payload)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to create realm: %s", string(body))
	}
	return nil
}

//------------------------------------------
// client management
//------------------------------------------

func (c *Client) CreateClient(opts CreateClientParams) (string, error) {

	if opts.Protocol == "" {
		opts.Protocol = "openid-connect"
	}

	opts.Enabled = true
	opts.StandardFlowEnabled = true
	opts.ImplicitFlowEnabled = false
	opts.DirectAccessGrantsEnabled = false
	opts.ServiceAccountsEnabled = false

	if opts.RedirectURIs == nil {
		opts.RedirectURIs = []string{}
	}
	if opts.WebOrigins == nil {
		opts.WebOrigins = []string{}
	}
	if opts.Attributes == nil {
		opts.Attributes = map[string]string{}
	}

	opts.Attributes["realm_client"] = "false"
	opts.Attributes["created_by"] = "admin"

	if _, ok := opts.Attributes["ui.icon"]; !ok {
		opts.Attributes["ui.icon"] = "applications"
	}
	if _, ok := opts.Attributes["ui.home"]; !ok {
		opts.Attributes["ui.home"] = "/admin"
	}

	if err := utils.ValidateRedirectURIs(opts.RedirectURIs); err != nil {
		return "", err
	}

	res, err := c.Post("clients", opts)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("failed to create client: %s", string(body))
	}

	location := res.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("no Location header returned by Keycloak")
	}

	parts := strings.Split(strings.TrimSpace(location), "/")
	clientUUID := parts[len(parts)-1]
	if clientUUID == "" {
		return "", fmt.Errorf("failed to parse Keycloak client ID from Location header")
	}

	if err := c.EnsureClientBaseline(context.Background(), clientUUID); err != nil {
		return "", err
	}

	return clientUUID, nil
}

func (c *Client) GetClientByClientID(clientID string) (*ClientInfo, error) {
	query := url.Values{}
	query.Set("clientId", clientID)

	res, err := c.Get("clients?" + query.Encode())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("failed checking clientId '%s': %s", clientID, string(body))
	}

	var clients []ClientInfo
	if err := json.NewDecoder(res.Body).Decode(&clients); err != nil {
		return nil, err
	}

	if len(clients) == 0 {
		return nil, nil
	}

	return &clients[0], nil
}

func (c *Client) GetClientByID(id string) (*ClientInfo, error) {
	res, err := c.Get("clients/" + id)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("failed to get client: %s", string(body))
	}

	var cli ClientInfo
	if err := json.NewDecoder(res.Body).Decode(&cli); err != nil {
		return nil, err
	}

	return &cli, nil
}

func (c *Client) ListClients() ([]ClientInfo, error) {
	res, err := c.Get("clients")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("failed to list clients: %s", string(body))
	}

	var clients []ClientInfo
	if err := json.NewDecoder(res.Body).Decode(&clients); err != nil {
		return nil, err
	}

	return clients, nil
}

func (c *Client) ListClientsWithBaselineCheck() ([]ClientInfo, error) {
	ctx := context.Background()
	clients, err := c.ListClients()
	if err != nil {
		return nil, err
	}

	for _, client := range clients {
		if err := c.EnsureClientBaseline(ctx, client.ID); err != nil {
			return nil, fmt.Errorf(
				"client baseline check (clientId %s) failed: %w",
				client.ClientID,
				err,
			)

		}
	}

	return clients, nil
}

func (c *Client) UpdateClient(id string, payload *model.Client) error {
	res, err := c.Put("clients/"+id, payload)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf(
			"failed to update client (status %d): %s",
			res.StatusCode,
			string(body),
		)
	}

	return nil
}

func (c *Client) UpdateClientEnabled(
	ctx context.Context,
	clientId string,
	enabled bool,
) error {

	clientUUID, err := c.resolveClientUUID(ctx, clientId)
	if err != nil {
		return fmt.Errorf("failed to resolve client UUID: %w", err)
	}

	u := fmt.Sprintf("clients/%s", clientUUID)

	res, err := c.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("get client failed [%d]: %s", res.StatusCode, string(b))
	}

	var kcClient ClientInfo
	if err := json.NewDecoder(res.Body).Decode(&kcClient); err != nil {
		return fmt.Errorf("decode client failed: %w", err)
	}

	kcClient.Enabled = enabled

	putRes, err := c.Put(u, kcClient)
	if err != nil {
		return err
	}
	defer putRes.Body.Close()

	if putRes.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(putRes.Body)
		return fmt.Errorf("update client enabled failed [%d]: %s", putRes.StatusCode, string(b))
	}

	return nil
}

func (c *Client) DeleteClient(id string) error {
	res, err := c.Delete("clients/" + id)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to delete client: %s", string(body))
	}

	return nil
}

// -----------------------------------------
// user management
// -----------------------------------------

func (c *Client) CreateUser(user *model.User) (string, error) {
	payload := map[string]any{
		"username":      user.Username,
		"email":         user.Email,
		"enabled":       user.Enabled,
		"firstName":     user.FirstName,
		"lastName":      user.LastName,
		"emailVerified": user.EmailVerified,
		"credentials": []map[string]any{
			{
				"type":      "password",
				"value":     uuid.NewString(),
				"temporary": true,
			},
		},
		"requiredActions": []string{
			"UPDATE_PASSWORD",
			"VERIFY_EMAIL",
		},
		"attributes": map[string][]string{
			"created_by": {"admin"},
			"user_type":  {"human"},
		},
	}

	res, err := c.Post("users", payload)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("failed to create user: %s", string(body))
	}

	location := res.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("no Location header returned by Keycloak")
	}

	parts := strings.Split(strings.TrimSpace(location), "/")
	kcID := parts[len(parts)-1]

	if kcID == "" {
		return "", fmt.Errorf("failed to parse Keycloak user ID from Location header")
	}

	return kcID, nil
}

func (c *Client) FindUsers(ctx context.Context, q string, exact bool) ([]UserRep, error) {
	u := fmt.Sprintf("realms/%s/users", c.Realm)

	v := url.Values{}
	if q != "" {
		v.Set("search", q)
	}
	if exact {
		v.Set("exact", "true")
	}
	u = u + "?" + v.Encode()

	res, err := c.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("keycloak find users failed: status=%d body=%s", res.StatusCode, string(b))
	}

	var out []UserRep
	return out, json.NewDecoder(res.Body).Decode(&out)
}

func (c *Client) ListUsers() ([]UserInfo, error) {
	res, err := c.Get("users")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("failed to list users: %s", string(body))
	}

	var users []UserInfo
	if err := json.NewDecoder(res.Body).Decode(&users); err != nil {
		return nil, err
	}

	return users, nil
}

func (c *Client) GetUser(userID string) (*UserInfo, error) {
	res, err := c.Get("users/" + userID)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("failed to get user: %s", string(body))
	}

	var user UserInfo
	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *Client) UpdateUser(user *model.User) error {
	if user.ID == "" {
		return fmt.Errorf("missing Keycloak user ID")
	}

	payload := map[string]any{
		"username":  user.Username,
		"email":     user.Email,
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"enabled":   user.Enabled,
	}

	res, err := c.Put(("users" + user.ID),
		payload,
	)
	if err != nil {
		return fmt.Errorf("keycloak update request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to update user in keycloak: %s", string(body))
	}

	return nil
}

func (c *Client) DeleteUser(userID string) error {
	if userID == "" {
		return fmt.Errorf("invalid user ID")
	}

	res, err := c.Delete("users/" + userID)
	if err != nil {
		return fmt.Errorf("keycloak delete request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to delete user in keycloak: %s", string(body))
	}

	return nil
}

// -------------------------------------------------------------------
// Realm roles Management
// -------------------------------------------------------------------

func (c *Client) CreateRealmRole(ctx context.Context, roleName, description string) error {
	u := "roles"
	payload := CreateRoleRequest{
		Name:        roleName,
		Description: description,
		Composite:   false,
		ClientRole:  false,
	}

	res, err := c.Post(u, payload)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// 409 = already exists (safe for bootstrap)
	if res.StatusCode == http.StatusConflict {
		return nil
	}

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("create realm role failed: %s", string(b))
	}

	return nil
}

func (c *Client) ListRealmRoles(ctx context.Context) ([]RoleRep, error) {
	u := "roles"

	res, err := c.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("list realm roles failed: %s", string(b))
	}

	var roles []RoleRep
	return roles, json.NewDecoder(res.Body).Decode(&roles)
}

func (c *Client) GetRealmRoleByName(ctx context.Context, roleName string) (*RoleRep, error) {
	res, err := c.Get(fmt.Sprintf("roles/%s", url.PathEscape(roleName)))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("keycloak get role failed: status=%d body=%s", res.StatusCode, string(b))
	}

	var out RoleRep
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) AddRealmRoleToUser(ctx context.Context, userID string, role RoleRep) error {
	payload := []RoleRep{role}
	bs, _ := json.Marshal(payload)

	bytesData := bytes.NewReader(bs)

	u := fmt.Sprintf("users/%s/role-mappings/realm", userID)

	res, err := c.Post(u, bytesData)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("keycloak add role failed: status=%d body=%s", res.StatusCode, string(b))
	}
	return nil
}

func (c *Client) RemoveRealmRoleFromUser(
	ctx context.Context,
	userID string,
	role RoleRep,
) error {

	u := fmt.Sprintf(
		"users/%s/role-mappings/realm",
		userID,
	)

	bs, _ := json.Marshal([]RoleRep{role})

	bytesData := bytes.NewReader(bs)

	res, err := c.Post(u, bytesData)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("remove realm role failed: %s", string(b))
	}

	return nil
}

func (c *Client) BootstrapRealmRoles(ctx context.Context) error {
	roles := []string{
		"admin",
		"user",
		"manager",
	}

	for _, r := range roles {
		if err := c.CreateRealmRole(ctx, r, "system role"); err != nil {
			return err
		}
	}

	return nil
}

// -------------------------------------------------------------------
// ROLES: Client roles
// -------------------------------------------------------------------

func (c *Client) CreateClientRole(
	ctx context.Context,
	clientId string,
	req *model.CreateClientRoleRequest,
) error {
	clientUUID, err := c.resolveClientUUID(ctx, clientId)
	if err != nil {
		return fmt.Errorf("failed to resolve client UUID: %w", err)
	}

	path := fmt.Sprintf(
		"clients/%s/roles",
		clientUUID,
	)
	payload := map[string]any{
		"name":        req.Role,
		"description": req.Description,
		"clientRole":  true,
	}

	res, err := c.Post(path, payload)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusConflict {
		return nil
	}

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf(
			"create client role failed [%d]: %s",
			res.StatusCode,
			string(b),
		)
	}

	return nil
}

func (c *Client) GetClientRoleByName(
	ctx context.Context,
	clientID string,
	clientUUID string,
	roleName string,
) (*ClientRoleRep, error) {

	if clientID == "" || clientUUID == "" {
		return nil, fmt.Errorf("clientID and clientUUID is required")
	}
	if roleName == "" {
		return nil, fmt.Errorf("roleName is required")
	}

	path := fmt.Sprintf(
		"clients/%s/roles/%s",
		clientUUID,
		url.PathEscape(roleName),
	)

	res, err := c.Get(path)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf(
			"get client role failed [%d]: %s",
			res.StatusCode,
			string(b),
		)
	}

	var role ClientRoleRep
	if err := json.NewDecoder(res.Body).Decode(&role); err != nil {
		return nil, err
	}

	return &role, nil
}

func (c *Client) GetUserClientRoles(
	ctx context.Context,
	userID string,
	clientId string,
	clientUuid string,
) ([]ClientRoleRep, error) {

	if clientId == "" || clientUuid == "" {
		return nil, fmt.Errorf("clientId  and clientUuid are required")
	}

	path := fmt.Sprintf(
		"users/%s/role-mappings/clients/%s",
		userID,
		clientUuid,
	)

	res, err := c.Get(path)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf(
			"get user client roles failed [%d]: %s",
			res.StatusCode,
			string(b),
		)
	}

	var roles []ClientRoleRep
	if err := json.NewDecoder(res.Body).Decode(&roles); err != nil {
		return nil, err
	}

	return roles, nil
}

func (c *Client) AssignClientRolesToUser(
	ctx context.Context,
	userID string,
	clientID string,
	clientUUID string,
	roles []string,
) error {

	if clientID == "" || clientUUID == "" {
		return fmt.Errorf("clientId and clientUuid are required")
	}

	if len(roles) == 0 {
		return nil
	}

	payload := make([]ClientRoleRequest, 0, len(roles))

	for _, roleName := range roles {
		role, err := c.GetClientRoleByName(ctx, clientID, clientUUID, roleName)
		if err != nil {
			return fmt.Errorf(
				"failed to resolve role %q for client %q (uuid: %q): %w",
				roleName,
				clientID,
				clientUUID,
				err,
			)

		}

		payload = append(payload, ClientRoleRequest{
			ID:   role.ID,
			Name: role.Name,
		})
	}

	path := fmt.Sprintf(
		"users/%s/role-mappings/clients/%s",
		userID,
		clientUUID,
	)

	res, err := c.Post(path, payload)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf(
			"assign client roles failed [%d]: %s",
			res.StatusCode,
			string(b),
		)
	}

	return nil
}

func (c *Client) RemoveClientRoleFromUser(
	ctx context.Context,
	userID, clientID string,
	role ClientRoleRep,
) error {

	u := fmt.Sprintf(
		"users/%s/role-mappings/clients/%s",
		userID, clientID,
	)

	bs, _ := json.Marshal([]ClientRoleRep{role})

	res, err := c.Post(u, bytes.NewReader(bs))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("remove client role failed: %s", string(b))
	}
	return nil
}

func (c *Client) RemoveClientRolesFromUser(
	ctx context.Context,
	userID string,
	clientId string,
	clientUuid string,
	roles []string,
) error {

	if clientId == "" || clientUuid == "" {
		return fmt.Errorf("clientId and clientUuid is required")
	}

	if len(roles) == 0 {
		return nil
	}

	clientUUID, err := c.resolveClientUUID(ctx, clientId)
	if err != nil {
		return fmt.Errorf("failed to resolve client UUID: %w", err)
	}

	payload := make([]ClientRoleRep, 0, len(roles))

	for _, roleName := range roles {
		role, err := c.GetClientRoleByName(ctx, clientId, clientUuid, roleName)
		if err != nil {
			return fmt.Errorf(
				"failed to resolve role %q for client %q: %w",
				roleName,
				clientId,
				err,
			)
		}
		payload = append(payload, *role)
	}

	path := fmt.Sprintf(
		"users/%s/role-mappings/clients/%s",
		userID,
		clientUUID,
	)

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	res, err := c.DeleteWithBody(path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf(
			"remove client roles failed [%d]: %s",
			res.StatusCode,
			string(b),
		)
	}

	return nil
}

func (c *Client) ListClientRoles(
	ctx context.Context,
	clientId string,
) ([]ClientRoleRep, error) {

	clientUUID, err := c.resolveClientUUID(ctx, clientId)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve client UUID: %w", err)
	}

	path := fmt.Sprintf("clients/%s/roles", clientUUID)

	res, err := c.Get(path)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf(
			"list client roles failed [%d]: %s",
			res.StatusCode,
			string(b),
		)
	}

	var roles []ClientRoleRep
	if err := json.NewDecoder(res.Body).Decode(&roles); err != nil {
		return nil, err
	}

	return roles, nil
}

func (c *Client) DeleteClientRole(
	ctx context.Context,
	clientId string,
	role string,
) error {

	clientUUID, err := c.resolveClientUUID(ctx, clientId)
	if err != nil {
		return fmt.Errorf("failed to resolve client UUID: %w", err)
	}

	path := fmt.Sprintf(
		"clients/%s/roles/%s",
		clientUUID,
		url.PathEscape(role),
	)

	res, err := c.Delete(path)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf(
			"delete client role failed [%d]: %s",
			res.StatusCode,
			string(b),
		)
	}

	return nil
}

func (c *Client) resolveClientUUID(
	ctx context.Context,
	clientId string,
) (string, error) {

	if clientId == "" {
		return "", fmt.Errorf("resolveClientUUID: clientId is required")
	}

	path := "clients?clientId=" + url.QueryEscape(clientId)

	res, err := c.Get(path)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"resolveClientUUID: failed [%d]: %s",
			res.StatusCode,
			string(body),
		)
	}

	var clients []struct {
		ID       string `json:"id"`
		ClientID string `json:"clientId"`
	}

	if err := json.Unmarshal(body, &clients); err != nil {
		return "", err
	}

	if len(clients) == 0 {
		return "", fmt.Errorf(
			"client not found in keycloak: %s",
			clientId,
		)
	}

	return clients[0].ID, nil
}

// support helper functions
func (c *Client) SendUserOnboardingEmail(
	ctx context.Context,
	userID string,
) error {

	actions := []string{
		"UPDATE_PASSWORD",
		"VERIFY_EMAIL",
	}

	url := fmt.Sprintf(
		"users/%s/execute-actions-email",
		userID,
	)

	req, _ := json.Marshal(actions)

	resp, err := c.Put(url, bytes.NewBuffer(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("failed to send email: %s", resp.Status)
	}

	return nil
}

// -------------------------------------------------------------------
// USERS: Reset password (ADMIN)
// -------------------------------------------------------------------

func (c *Client) ResetUserPassword(
	ctx context.Context,
	userID string,
) error {

	tmpPassword := generateTemporaryPassword()

	payload := map[string]any{
		"type":      "password",
		"value":     tmpPassword,
		"temporary": true,
	}

	u := fmt.Sprintf("users/%s/reset-password", userID)

	res, err := c.Put(u, payload)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf(
			"reset user password failed [%d]: %s",
			res.StatusCode,
			string(b),
		)
	}

	actions := []string{"UPDATE_PASSWORD"}

	actionsRes, err := c.Put(
		fmt.Sprintf("users/%s/execute-actions-email", userID),
		actions,
	)
	if err != nil {
		return err
	}
	defer actionsRes.Body.Close()

	if actionsRes.StatusCode != http.StatusNoContent &&
		actionsRes.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(actionsRes.Body)
		return fmt.Errorf(
			"execute required actions failed [%d]: %s",
			actionsRes.StatusCode,
			string(b),
		)
	}

	return nil
}

func generateTemporaryPassword() string {
	return utils.RandomString(16) + "!A1"
}

func (c *Client) attachDefaultClientScopes(
	ctx context.Context,
	clientUUID string,
) error {

	scopes, err := c.ListClientScopes(ctx)
	if err != nil {
		return err
	}

	scopeByName := map[string]string{}
	for _, s := range scopes {
		scopeByName[s.Name] = s.ID
	}

	for _, name := range defaultClientScopes {
		scopeID, ok := scopeByName[name]
		if !ok {
			return fmt.Errorf("required client scope %q not found", name)
		}

		if _, err := c.Put(
			fmt.Sprintf(
				"clients/%s/default-client-scopes/%s",
				clientUUID,
				scopeID,
			),
			nil,
		); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) ListClientScopes(ctx context.Context) ([]ClientScope, error) {
	res, err := c.Get("client-scopes")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf(
			"failed to list client scopes: %s",
			string(body),
		)
	}

	var scopes []ClientScope
	if err := json.NewDecoder(res.Body).Decode(&scopes); err != nil {
		return nil, err
	}

	return scopes, nil
}

func (c *Client) EnsureClientScopes(
	ctx context.Context,
	required []string,
) (map[string]string, error) {

	existing, err := c.ListClientScopes(ctx)
	if err != nil {
		return nil, err
	}

	scopeByName := map[string]string{}
	for _, s := range existing {
		scopeByName[s.Name] = s.ID
	}

	for _, name := range required {
		if _, ok := scopeByName[name]; ok {
			continue
		}

		payload := map[string]any{
			"name":     name,
			"protocol": "openid-connect",
		}

		res, err := c.Post("client-scopes", payload)
		if err != nil {
			return nil, err
		}
		res.Body.Close()

		if res.StatusCode != http.StatusCreated {
			return nil, fmt.Errorf("failed to create client scope %q", name)
		}

		// refresh scopes
		existing, err = c.ListClientScopes(ctx)
		if err != nil {
			return nil, err
		}
		for _, s := range existing {
			scopeByName[s.Name] = s.ID
		}
	}

	return scopeByName, nil
}

func (c *Client) DetectMissingClientScopes(
	ctx context.Context,
	clientUUID string,
	required map[string]string,
) ([]string, error) {

	res, err := c.Get(fmt.Sprintf(
		"clients/%s/default-client-scopes",
		clientUUID,
	))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch client scopes")
	}

	var attached []ClientScope
	if err := json.NewDecoder(res.Body).Decode(&attached); err != nil {
		return nil, err
	}

	attachedNames := map[string]bool{}
	for _, s := range attached {
		attachedNames[s.Name] = true
	}

	var missing []string
	for name := range required {
		if !attachedNames[name] {
			missing = append(missing, name)
		}
	}

	return missing, nil
}

func ApplyClientTemplate(
	opts *CreateClientParams,
	template ClientTemplate,
) {
	switch template {

	case WebClient:
		opts.PublicClient = false
		opts.StandardFlowEnabled = true
		opts.DirectAccessGrantsEnabled = false
		opts.ServiceAccountsEnabled = false

	case ServiceClient:
		opts.PublicClient = false
		opts.StandardFlowEnabled = false
		opts.DirectAccessGrantsEnabled = true
		opts.ServiceAccountsEnabled = true

	case AdminClient:
		opts.PublicClient = false
		opts.StandardFlowEnabled = true
		opts.DirectAccessGrantsEnabled = true
		opts.ServiceAccountsEnabled = true
	}
}

func (c *Client) CreateClientWithTemplate(
	ctx context.Context,
	opts CreateClientParams,
	template ClientTemplate,
) (string, error) {

	ApplyClientTemplate(&opts, template)

	// Validate redirects
	if err := utils.ValidateRedirectURIs(opts.RedirectURIs); err != nil {
		return "", err
	}

	// Ensure scopes exist
	requiredScopes := []string{"profile", "email", "roles"}
	scopeIDs, err := c.EnsureClientScopes(ctx, requiredScopes)
	if err != nil {
		return "", err
	}

	// Create client
	clientUUID, err := c.CreateClient(opts)
	if err != nil {
		return "", err
	}

	// Detect drift
	missing, err := c.DetectMissingClientScopes(
		ctx,
		clientUUID,
		scopeIDs,
	)
	if err != nil {
		return "", err
	}

	// Fix drift
	if len(missing) > 0 {
		if err := c.FixClientScopeDrift(
			ctx,
			clientUUID,
			scopeIDs,
			missing,
		); err != nil {
			return "", err
		}
	}

	return clientUUID, nil
}

func (c *Client) FixClientScopeDrift(
	ctx context.Context,
	clientUUID string,
	scopeIDs map[string]string,
	missing []string,
) error {

	for _, name := range missing {
		scopeID, ok := scopeIDs[name]
		if !ok || scopeID == "" {
			return fmt.Errorf("cannot fix drift: scope %q not found", name)
		}

		res, err := c.Put(
			fmt.Sprintf("clients/%s/default-client-scopes/%s", clientUUID, scopeID),
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to attach client scope %q: %w", name, err)
		}
		defer res.Body.Close()

		// Keycloak typically returns 204 No Content
		if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(res.Body)
			return fmt.Errorf(
				"failed to attach client scope %q: %s",
				name,
				string(body),
			)
		}
	}

	return nil
}

func (c *Client) EnsureClientBaseline(
	ctx context.Context,
	clientUUID string,
) error {

	// 1. Ensure required scopes exist in the realm
	scopeIDs, err := c.EnsureClientScopes(
		ctx,
		[]string{"profile", "email", "roles"},
	)
	if err != nil {
		return err
	}

	// 2. Detect drift on this client
	missing, err := c.DetectMissingClientScopes(
		ctx,
		clientUUID,
		scopeIDs,
	)
	if err != nil {
		return err
	}

	// 3. Heal drift (attach missing scopes)
	if len(missing) > 0 {
		if err := c.FixClientScopeDrift(
			ctx,
			clientUUID,
			scopeIDs,
			missing,
		); err != nil {
			return err
		}
	}

	return nil
}
