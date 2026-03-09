package service

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/moh-sso-dashboard/internal/keycloak"
	repository "github.com/moh-sso-dashboard/internal/repository/auth"
)

type AuthService interface {
	ProcessAuthCode(code string, codeVerifier string) (*keycloak.TokenResponse, error)
	GetAccessToken(refreshToken string) (*keycloak.TokenResponse, error)
	GetMe(accessToken string) (*keycloak.AuthUser, error)
	SaveSession(userID, access, refresh string, expires int) error
	LogOut(refreshToken string) error
	GetLogoutURL(idTokenHint, postLogoutRedirectURI string) string
}

type authService struct {
	authRepo repository.AuthRepository
	redis    *redis.Client
}

func NewAuthService(authRepo repository.AuthRepository, redisClient *redis.Client) AuthService {
	return &authService{
		authRepo: authRepo,
		redis:    redisClient,
	}
}

func (s *authService) ProcessAuthCode(code string, codeVerifier string) (*keycloak.TokenResponse, error) {
	if code == "" || codeVerifier == "" {
		return nil, fmt.Errorf("missing auth code or pkce verifier")
	}
	tokens, err := s.authRepo.ExchangeCode(code, codeVerifier)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	return tokens, nil
}

func (s *authService) GetAccessToken(refreshToken string) (*keycloak.TokenResponse, error) {
	accessToken, err := s.authRepo.GetAccessToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}
	return accessToken, nil
}

func (s *authService) GetMe(accessToken string) (*keycloak.AuthUser, error) {
	userProfile, err := s.authRepo.GetMe(accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	return userProfile, nil
}

func (s *authService) SaveSession(userID, access, refresh string, expires int) error {
	ctx := context.Background()
	key := fmt.Sprintf("session:%s", userID)

	data := map[string]interface{}{
		"access_token":  access,
		"refresh_token": refresh,
		"expires_in":    expires,
	}

	if err := s.redis.HSet(ctx, key, data).Err(); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	s.redis.Expire(ctx, key, time.Duration(expires)*time.Second)
	return nil
}

func (s *authService) LogOut(refreshToken string) error {
	err := s.authRepo.Logout(refreshToken)
	if err != nil {
		return fmt.Errorf("failed to logout user: %w", err)
	}

	return nil
}

func (s *authService) GetLogoutURL(idTokenHint, postLogoutRedirectURI string) string {
	return s.authRepo.GetLogoutURL(idTokenHint, postLogoutRedirectURI)
}
