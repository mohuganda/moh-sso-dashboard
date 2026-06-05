package document_templates

import (
	"github.com/moh-sso-dashboard/internal/config"
	baseservice "github.com/moh-sso-dashboard/internal/service"
)

type Service = baseservice.DocumentTemplateService
type SheetService = baseservice.DocumentTemplateSheetService
type ColumnService = baseservice.DocumentTemplateColumnService

func NewService(
	repo Repository,
	sheetRepo SheetRepository,
	columnRepo ColumnRepository,
	notifications baseservice.NotificationsService,
	cfg ...*config.Config,
) Service {
	return baseservice.NewDocumentTemplateService(repo, sheetRepo, columnRepo, notifications, cfg...)
}

func NewSheetService(
	repo SheetRepository,
	notifications baseservice.NotificationsService,
	cfg ...*config.Config,
) SheetService {
	return baseservice.NewDocumentTemplateSheetService(repo, notifications, cfg...)
}

func NewColumnService(
	repo ColumnRepository,
	notifications baseservice.NotificationsService,
	cfg ...*config.Config,
) ColumnService {
	return baseservice.NewDocumentTemplateColumnService(repo, notifications, cfg...)
}
