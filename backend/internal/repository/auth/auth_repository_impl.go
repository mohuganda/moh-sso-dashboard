package auth

import (
	"errors"
	"fmt"

	"github.com/moh-sso-dashboard/internal/config"
	"github.com/moh-sso-dashboard/internal/keycloak"
)

type keycloakAuthRepository struct {
	keycloakClient *keycloak.Client
	redirectURI    string
}

func NewAuthRepository(kcClient *keycloak.Client, keycloakAdminClient *keycloak.KeyAdminClient, config *config.Config) AuthRepository {
	return &keycloakAuthRepository{
		keycloakClient: kcClient,
		redirectURI:    config.KeycloakRedirectUri,
	}
}

func (r *keycloakAuthRepository) ExchangeCode(code string,
	codeVerifier string) (*keycloak.TokenResponse, error) {
	if r.redirectURI == "" {
		return nil, errors.New("KEYCLOAK_REDIRECT_URI environment variable is not set")
	}

	tokens, err := r.keycloakClient.ExchangeCodeForToken(code, r.redirectURI, codeVerifier)
	if err != nil {
		return nil, fmt.Errorf("keycloak exchange failed: %w", err)
	}

	return tokens, nil
}

func (r *keycloakAuthRepository) GetAccessToken(refreshToken string) (*keycloak.TokenResponse, error) {
	if r.redirectURI == "" {
		return nil, errors.New("KEYCLOAK_REDIRECT_URI environment variable is not set")
	}

	tokens, err := r.keycloakClient.AccessToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("keycloak exchange failed: %w", err)
	}

	return tokens, nil

}

func (r *keycloakAuthRepository) GetMe(accessToken string) (*keycloak.AuthUser, error) {

	if accessToken == "" {
		return nil, errors.New("refresh token not provided")
	}

	user, err := r.keycloakClient.Me(accessToken)
	if err != nil {
		return nil, fmt.Errorf("keycloak exchange failed: %w", err)
	}

	return user, nil
}

func (r *keycloakAuthRepository) Logout(refreshToken string) error {
	if refreshToken == "" {
		return errors.New("refresh token not provided")
	}
	err := r.keycloakClient.LogOut(refreshToken)
	if err != nil {
		return fmt.Errorf("keycloak exchange failed: %w", err)
	}

	return nil
}

func (r *keycloakAuthRepository) GetLogoutURL(idTokenHint, postLogoutRedirectURI string) string {
	return r.keycloakClient.GetLogoutURL(idTokenHint, postLogoutRedirectURI)
}
