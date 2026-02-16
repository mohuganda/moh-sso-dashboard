package keycloak

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/moh-sso-dashboard/internal/cache"
)

// Client represents the Keycloak API client
// - Web auth/refresh/logout use dashboard-web (authorization_code + refresh_token)
type KCUserInfo struct {
	UserID            string `json:"sub"`
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	Name              string `json:"name"`
}

type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
	IDToken          string `json:"id_token"`
}

type AuthUser struct {
	ID            string              `json:"id"`
	Username      string              `json:"username"`
	Email         string              `json:"email"`
	FirstName     string              `json:"firstName"`
	LastName      string              `json:"lastName"`
	FullName      string              `json:"fullName"`
	IsAdmin       bool                `json:"isAdmin"`
	IsUser        bool                `json:"isUser"`
	RealmRoles    []string            `json:"realmRoles"`
	ClientRoles   map[string][]string `json:"clientRoles"`
	Enabled       bool                `json:"enabled"`
	EmailVerified bool                `json:"emailVerified"`
	LastLoginAt   *time.Time          `json:"lastLoginAt"`
}

// ----------------------------------------------------
// CLIENT
// ----------------------------------------------------

type Client struct {
	BaseURL      string
	Realm        string
	ClientID     string
	ClientSecret string
	Token        string
	httpClient   *http.Client
	cache        *cache.RedisCache
}

func NewWebClient(
	baseURL string,
	realm string,
	clientID string,
	clientSecret string,
	cache *cache.RedisCache, // <-- FIXED (pointer)
) *Client {

	baseURL = strings.TrimSuffix(baseURL, "/")

	return &Client{
		BaseURL:      baseURL,
		Realm:        realm,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		cache:        cache,
	}
}

// ----------------------------------------------------
// WEB AUTH (authorization_code)
// ----------------------------------------------------
func (c *Client) ExchangeCodeForToken(
	code string,
	redirectURI string,
	codeVerifier string,
) (*TokenResponse, error) {

	tokenURL := fmt.Sprintf(
		"%s/realms/%s/protocol/openid-connect/token",
		c.BaseURL,
		c.Realm,
	)

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.ClientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	form.Set("code_verifier", codeVerifier)

	res, err := c.httpClient.PostForm(tokenURL, form)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed [%d]: %s", res.StatusCode, string(bodyBytes))
	}

	var tokenRes TokenResponse
	if err := json.Unmarshal(bodyBytes, &tokenRes); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenRes, nil
}

// ----------------------------------------------------
// REFRESH TOKEN (WEB CLIENT)
// ----------------------------------------------------
func (c *Client) AccessToken(refreshToken string) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf(
		"%s/realms/%s/protocol/openid-connect/token",
		c.BaseURL,
		c.Realm,
	)

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.ClientSecret)

	res, err := c.httpClient.PostForm(tokenURL, form)
	if err != nil {
		return nil, fmt.Errorf("refresh token failed: %w", err)
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh token failed [%d]: %s", res.StatusCode, string(bodyBytes))
	}

	var tokenRes TokenResponse
	if err := json.Unmarshal(bodyBytes, &tokenRes); err != nil {
		return nil, fmt.Errorf("failed to decode refresh response: %w", err)
	}

	return &tokenRes, nil
}

// ----------------------------------------------------
// LOGOUT (WEB CLIENT)
// ----------------------------------------------------
func (c *Client) LogOut(refreshToken string) error {
	if refreshToken == "" {
		return fmt.Errorf("missing refresh token for logout")
	}

	logoutURL := fmt.Sprintf(
		"%s/realms/%s/protocol/openid-connect/logout",
		c.BaseURL,
		c.Realm,
	)

	form := url.Values{}
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.ClientSecret)
	form.Set("refresh_token", refreshToken)

	res, err := c.httpClient.PostForm(logoutURL, form)
	if err != nil {
		return fmt.Errorf("logout request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("logout failed [%d]: %s", res.StatusCode, string(body))
	}

	return nil
}

// ----------------------------------------------------
// TOKEN → FULL USER PROFILE (ADMIN API)
// ----------------------------------------------------
func (c *Client) Me(accessToken string) (*AuthUser, error) {

	ctx := context.Background()

	ui, err := c.getUserInfo(accessToken)
	if err != nil {
		return nil, err
	}

	if ui.UserID == "" {
		return nil, errors.New("userinfo missing user_id")
	}

	cacheKey := "auth:user:" + ui.UserID

	var cached AuthUser
	if ok, err := c.cache.Get(ctx, cacheKey, &cached); err == nil && ok {
		return &cached, nil
	}

	// ----------------------------------------------------
	// Extract roles directly from token claims
	// ----------------------------------------------------

	claims, err := c.parseToken(accessToken)
	if err != nil {
		return nil, err
	}

	realmRoles := []string{}
	clientRoles := map[string][]string{}
	isAdmin := false
	isUser := false

	// Realm roles
	if realmAccess, ok := claims["realm_access"].(map[string]any); ok {
		if roles, ok := realmAccess["roles"].([]any); ok {
			for _, r := range roles {
				role := fmt.Sprint(r)
				realmRoles = append(realmRoles, role)

				if role == "admin" {
					isAdmin = true
				}
				if role == "user" {
					isUser = true
				}
			}
		}
	}

	// Client roles
	if resourceAccess, ok := claims["resource_access"].(map[string]any); ok {
		for clientID, v := range resourceAccess {
			if clientData, ok := v.(map[string]any); ok {
				if roles, ok := clientData["roles"].([]any); ok {
					for _, r := range roles {
						clientRoles[clientID] = append(
							clientRoles[clientID],
							fmt.Sprint(r),
						)
					}
				}
			}
		}
	}

	firstName := ui.GivenName
	lastName := ui.FamilyName

	fullName := strings.TrimSpace(firstName + " " + lastName)
	if fullName == "" {
		fullName = ui.Name
	}

	user := &AuthUser{
		ID:            ui.UserID,
		Username:      ui.PreferredUsername,
		Email:         ui.Email,
		FirstName:     firstName,
		LastName:      lastName,
		FullName:      fullName,
		IsAdmin:       isAdmin,
		IsUser:        isUser,
		RealmRoles:    realmRoles,
		ClientRoles:   clientRoles,
		Enabled:       true,
		EmailVerified: ui.EmailVerified,
	}

	_ = c.cache.Set(ctx, cacheKey, user, 10*time.Minute)

	return user, nil
}

// ----------------------------------------------------
// UserInfo
// ----------------------------------------------------
func (c *Client) getUserInfo(accessToken string) (*KCUserInfo, error) {
	url := fmt.Sprintf(
		"%s/realms/%s/protocol/openid-connect/userinfo",
		c.BaseURL,
		c.Realm,
	)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("userinfo failed [%d]: %s", res.StatusCode, string(body))
	}

	var ui KCUserInfo
	if err := json.NewDecoder(res.Body).Decode(&ui); err != nil {
		return nil, err
	}

	if ui.UserID == "" {
		return nil, errors.New("userinfo response missing user_id")
	}

	return &ui, nil
}

func (c *Client) parseToken(token string) (map[string]any, error) {

	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil, errors.New("invalid JWT format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}

	return claims, nil
}
