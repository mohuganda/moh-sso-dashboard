package sessions

import "github.com/moh-sso-dashboard/internal/keycloak"

type Service interface {
	GetUserSessions(userID string) ([]keycloak.Session, error)
	LogoutSession(sessionID string) error
}

type sessionService struct {
	sessionRepo Repository
}

func NewService(sessionRepo Repository) Service {
	return &sessionService{
		sessionRepo: sessionRepo,
	}
}

func (s *sessionService) GetUserSessions(userID string) ([]keycloak.Session, error) {
	sessions, err := s.sessionRepo.GetUserSessions(userID)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *sessionService) LogoutSession(sessionID string) error {
	err := s.sessionRepo.LogoutSession(sessionID)
	if err != nil {
		return err
	}
	return nil
}
