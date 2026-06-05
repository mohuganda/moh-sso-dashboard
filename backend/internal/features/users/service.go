package users

import (
	"github.com/moh-sso-dashboard/internal/config"
	baseservice "github.com/moh-sso-dashboard/internal/service"
)

type Service = baseservice.UserService

type CreateUserRequest = baseservice.CreateUserRequest
type UpdateUserRequest = baseservice.UpdateUserRequest

func NewService(
	repo Repository,
	notifications baseservice.NotificationsService,
	cfg ...*config.Config,
) *Service {
	return baseservice.NewUserService(repo, notifications, cfg...)
}
