package announcements

import (
	"github.com/moh-sso-dashboard/internal/config"
	baseservice "github.com/moh-sso-dashboard/internal/service"
)

type Service = baseservice.AnnouncementService

func NewService(
	repo Repository,
	notifications baseservice.NotificationsService,
	cfg ...*config.Config,
) *Service {
	return baseservice.NewAnnouncementService(repo, notifications, cfg...)
}
