package clients

import (
	"github.com/moh-sso-dashboard/internal/config"
	baseservice "github.com/moh-sso-dashboard/internal/service"
)

type Service = baseservice.ClientService

type CreateClientRequest = baseservice.CreateClientRequest

func NewService(
	repo Repository,
	notifications baseservice.NotificationsService,
	cfg ...*config.Config,
) *Service {
	return baseservice.NewClientService(repo, notifications, cfg...)
}
