package storage_locations

import (
	"github.com/moh-sso-dashboard/internal/config"
	baseservice "github.com/moh-sso-dashboard/internal/service"
)

type Service = baseservice.StorageLocationService

type CreateStorageLocationInput = baseservice.CreateStorageLocationInput
type UpdateStorageLocationInput = baseservice.UpdateStorageLocationInput

func NewService(
	repo Repository,
	notifications baseservice.NotificationsService,
	cfg ...*config.Config,
) Service {
	return baseservice.NewStorageLocationService(repo, notifications, cfg...)
}
