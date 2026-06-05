package sessions

import "github.com/moh-sso-dashboard/internal/keycloak"

type SessionRepository interface {
	GetUserSessions(userID string) ([]keycloak.Session, error)
	LogoutSession(sessionID string) error
}

type Repository = SessionRepository
