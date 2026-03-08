package session

import (
	"errors"
	"fmt"

	"github.com/moh-sso-dashboard/internal/config"
	"github.com/moh-sso-dashboard/internal/keycloak"
	logger "github.com/moh-sso-dashboard/internal/log"
)

type sessionRepository struct {
	keycloakClient *keycloak.KeyAdminClient
}

func NewSessionRepository(
	keycloakClient *keycloak.KeyAdminClient,
	config *config.Config,
	log logger.Logger,
) SessionRepository {

	return &sessionRepository{
		keycloakClient: keycloakClient,
	}
}

func (r *sessionRepository) GetUserSessions(userId string) ([]keycloak.Session, error) {
	if userId == "" {
		return nil, errors.New("userId not provided")
	}
	sessions, err := r.keycloakClient.GetUserSessions(userId)
	if err != nil {
		return nil, fmt.Errorf("keycloak exchange failed: %w", err)
	}
	return sessions, nil
}

func (r *sessionRepository) LogoutSession(sessionId string) error {
	if sessionId == "" {
		return errors.New("userId not provided")
	}

	err := r.keycloakClient.LogoutSession(sessionId)
	if err != nil {
		return err
	}
	return nil
}
