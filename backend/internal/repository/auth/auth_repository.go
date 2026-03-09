package auth

import "github.com/moh-sso-dashboard/internal/keycloak"

type AuthRepository interface {
	ExchangeCode(code string, codeVerifier string) (*keycloak.TokenResponse, error)
	GetAccessToken(refreshToken string) (*keycloak.TokenResponse, error)
	GetMe(accessToken string) (*keycloak.AuthUser, error)
	Logout(refreshToken string) error
	GetLogoutURL(idTokenHint, postLogoutRedirectURI string) string
}
