package bootstrap

import (
	"github.com/moh-sso-dashboard/internal/config"
	storepkg "github.com/moh-sso-dashboard/internal/db/sqlc"
	kcClientPkg "github.com/moh-sso-dashboard/internal/keycloak"
	logger "github.com/moh-sso-dashboard/internal/log"
	announcementRepo "github.com/moh-sso-dashboard/internal/repository/announcements"
	authRepo "github.com/moh-sso-dashboard/internal/repository/auth"
	clientRepo "github.com/moh-sso-dashboard/internal/repository/client"
	documentRepo "github.com/moh-sso-dashboard/internal/repository/document"
	documentTemplateRepo "github.com/moh-sso-dashboard/internal/repository/document_template"
	documentTemplateColumnRepo "github.com/moh-sso-dashboard/internal/repository/document_template_column"
	documentTemplateSheetRepo "github.com/moh-sso-dashboard/internal/repository/document_template_sheet"
	emailRepo "github.com/moh-sso-dashboard/internal/repository/email"
	metricsRepo "github.com/moh-sso-dashboard/internal/repository/metrics"
	notificationDeliveryRepo "github.com/moh-sso-dashboard/internal/repository/notification_delivery"
	notificationsRepo "github.com/moh-sso-dashboard/internal/repository/notifications"
	processRepo "github.com/moh-sso-dashboard/internal/repository/processes"
	sessionRepo "github.com/moh-sso-dashboard/internal/repository/session"
	storageLocationRepo "github.com/moh-sso-dashboard/internal/repository/storage_locations"
	surveillanceRepo "github.com/moh-sso-dashboard/internal/repository/surveillance"
	userRepo "github.com/moh-sso-dashboard/internal/repository/user"
)

type repositories struct {
	Auth                    authRepo.AuthRepository
	Clients                 clientRepo.ClientRepository
	Users                   userRepo.UserRepository
	Metrics                 metricsRepo.MetricsRepository
	Notifications           notificationsRepo.NotificationsRepository
	NotificationDelivery    notificationDeliveryRepo.NotificationDeliveryRepository
	Documents               documentRepo.DocumentRepository
	DocumentStockImports    documentRepo.StockImportRepository
	DocumentFiles           documentRepo.FileRepository
	DocumentTemplates       documentTemplateRepo.DocumentTemplateRepository
	DocumentTemplateColumns documentTemplateColumnRepo.DocumentTemplateColumnRepository
	DocumentTemplateSheets  documentTemplateSheetRepo.DocumentTemplateSheetRepository
	Processes               processRepo.ProcessRepository
	StorageLocations        storageLocationRepo.StorageLocationRepository
	Sessions                sessionRepo.SessionRepository
	Announcements           announcementRepo.AnnouncementRepository
	Email                   emailRepo.EmailRepository
	Surveillance            *surveillanceRepo.Repositories
}

func buildRepositories(
	cfg *config.Config,
	store storepkg.Store,
	adminKC *kcClientPkg.KeyAdminClient,
	webKC *kcClientPkg.Client,
	appLogger *logger.Logger,
) repositories {
	return repositories{
		Auth:                    authRepo.NewAuthRepository(webKC, adminKC, cfg),
		Clients:                 clientRepo.NewClientRepository(adminKC, cfg, store, *appLogger),
		Users:                   userRepo.NewUserRepository(adminKC, cfg, store, *appLogger),
		Metrics:                 metricsRepo.NewMetricsRepository(cfg, store, *appLogger),
		Notifications:           notificationsRepo.NewNotificationsRepository(store, *appLogger),
		NotificationDelivery:    notificationDeliveryRepo.NewNotificationDeliveryRepository(store, *appLogger),
		Documents:               documentRepo.NewDocumentRepository(cfg, store, *appLogger),
		DocumentStockImports:    documentRepo.NewStockImportRepository(),
		DocumentFiles:           documentRepo.NewFileRepository(),
		DocumentTemplates:       documentTemplateRepo.NewDocumentTemplateRepository(store),
		DocumentTemplateColumns: documentTemplateColumnRepo.NewDocumentTemplateColumnRepository(store),
		DocumentTemplateSheets:  documentTemplateSheetRepo.NewDocumentTemplateSheetRepository(store),
		Processes:               processRepo.NewProcessRepository(cfg, store, *appLogger),
		StorageLocations:        storageLocationRepo.NewStorageRepositoryRepository(cfg, store, *appLogger),
		Sessions:                sessionRepo.NewSessionRepository(adminKC, cfg, *appLogger),
		Announcements:           announcementRepo.NewAnnouncementRepository(store, *appLogger),
		Email:                   emailRepo.NewEmailRepository(cfg, store, *appLogger),
		Surveillance:            surveillanceRepo.NewRepositories(store),
	}
}
