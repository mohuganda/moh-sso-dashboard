package documents

import (
	"github.com/moh-sso-dashboard/internal/config"
	processrepo "github.com/moh-sso-dashboard/internal/repository/processes"
	baseservice "github.com/moh-sso-dashboard/internal/service"
	"github.com/moh-sso-dashboard/internal/storage"
)

type Service = baseservice.DocumentService

type CreateDocumentInput = baseservice.CreateDocumentInput
type EditDocumentInput = baseservice.EditDocumentInput

func NewService(
	repo Repository,
	processRepo processrepo.ProcessRepository,
	notifications baseservice.NotificationsService,
	fileStorage storage.Storage,
	cfg ...*config.Config,
) *Service {
	return baseservice.NewDocumentService(repo, processRepo, notifications, fileStorage, cfg...)
}
