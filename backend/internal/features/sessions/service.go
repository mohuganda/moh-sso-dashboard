package sessions

import baseservice "github.com/moh-sso-dashboard/internal/service"

type Service = baseservice.SessionService

func NewService(repo Repository) Service {
	return baseservice.NewSessionService(repo)
}
