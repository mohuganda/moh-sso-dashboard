package bootstrap

import (
	"github.com/moh-sso-dashboard/internal/cache"
	"github.com/moh-sso-dashboard/internal/config"
	storepkg "github.com/moh-sso-dashboard/internal/db/sqlc"
	announcementfeature "github.com/moh-sso-dashboard/internal/features/announcements"
	clientfeature "github.com/moh-sso-dashboard/internal/features/clients"
	documenttemplatefeature "github.com/moh-sso-dashboard/internal/features/document_templates"
	documentfeature "github.com/moh-sso-dashboard/internal/features/documents"
	sessionfeature "github.com/moh-sso-dashboard/internal/features/sessions"
	storagelocationfeature "github.com/moh-sso-dashboard/internal/features/storage_locations"
	surveillancefeature "github.com/moh-sso-dashboard/internal/features/surveillance"
	userfeature "github.com/moh-sso-dashboard/internal/features/users"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/service"
	importsvc "github.com/moh-sso-dashboard/internal/service/import"
	"github.com/moh-sso-dashboard/internal/storage"
	"github.com/moh-sso-dashboard/internal/worker"
	"github.com/redis/go-redis/v9"
)

type services struct {
	Email                   service.EmailService
	SMTP                    worker.EmailSender
	Notifications           service.NotificationsService
	Auth                    service.AuthService
	Metrics                 *service.MetricsService
	Audit                   *service.AuditService
	StorageLocations        storagelocationfeature.Service
	Sessions                sessionfeature.Service
	Documents               *documentfeature.Service
	Clients                 *clientfeature.Service
	Users                   *userfeature.Service
	Announcements           *announcementfeature.Service
	Diseases                *surveillancefeature.DiseaseService
	EpiWeeks                *surveillancefeature.EpiWeekService
	Locations               *surveillancefeature.LocationService
	Facilities              *surveillancefeature.FacilityService
	FacilityWeeklyMetrics   *surveillancefeature.FacilityWeeklyMetricsService
	WeeklyStatus            *surveillancefeature.WeeklyStatusService
	SurveillanceImport      *surveillancefeature.ImportService
	Alerts                  *surveillancefeature.AlertService
	DocumentTemplates       documenttemplatefeature.Service
	DocumentTemplateSheets  documenttemplatefeature.SheetService
	DocumentTemplateColumns documenttemplatefeature.ColumnService
	Import                  *importsvc.Service
}

type serviceDependencies struct {
	Config       *config.Config
	Store        storepkg.Store
	Cache        *cache.RedisCache
	CacheClient  *redis.Client
	Databases    databases
	Repositories repositories
	FileStorage  storage.Storage
	Logger       *logger.Logger
}

func buildServices(deps serviceDependencies) services {
	templateManager, err := service.NewTemplateManager(deps.Logger)
	if err != nil {
		deps.Logger.Fatal("Failed to initialize email template manager: ", err)
	}

	if err := service.RegisterDefaultTemplates(templateManager); err != nil {
		deps.Logger.Fatal("Failed to register default email templates: ", err)
	}

	deps.Logger.Info("Default email templates registered")

	smtpService, err := service.NewSMTPService(deps.Config, templateManager, deps.Logger)
	if err != nil {
		deps.Logger.Fatal("Failed to initialize SMTP service: ", err)
	}

	queueService, err := service.NewQueueService(deps.Repositories.Email, deps.Logger)
	if err != nil {
		deps.Logger.Fatal("Failed to initialize email queue service: ", err)
	}

	emailService, err := service.NewEmailService(smtpService, queueService, templateManager)
	if err != nil {
		deps.Logger.Fatal("Failed to initialize email application service: ", err)
	}

	deps.Logger.Info("Email application service initialized")

	publisher := cache.NewNotificationPublisher(deps.CacheClient)
	notificationsService := service.NewNotificationsService(
		deps.Config,
		deps.Repositories.Notifications,
		deps.Repositories.NotificationDelivery,
		publisher,
	)

	authService := service.NewAuthService(deps.Repositories.Auth, deps.CacheClient)
	metricsService := service.NewMetricsService(deps.Repositories.Metrics)
	auditService := service.NewAuditService(
		deps.Store,
		deps.Cache,
		notificationsService,
		deps.Config,
	)

	storageLocationService := service.NewStorageLocationService(
		deps.Repositories.StorageLocations,
		notificationsService,
		deps.Config,
	)

	sessionService := service.NewSessionService(deps.Repositories.Sessions)
	documentService := service.NewDocumentService(
		deps.Repositories.Documents,
		deps.Repositories.Processes,
		notificationsService,
		deps.FileStorage,
		deps.Config,
	)

	clientService := service.NewClientService(
		deps.Repositories.Clients,
		notificationsService,
		deps.Config,
	)

	userService := service.NewUserService(
		deps.Repositories.Users,
		notificationsService,
		deps.Config,
	)

	announcementService := service.NewAnnouncementService(
		deps.Repositories.Announcements,
		notificationsService,
		deps.Config,
	)

	diseaseService := service.NewSurveillanceDiseaseService(
		deps.Logger,
		deps.Repositories.Surveillance.Diseases,
		notificationsService,
		deps.Config,
	)

	epiWeekService := service.NewSurveillanceEpiWeekService(
		deps.Logger,
		deps.Repositories.Surveillance.EpiWeeks,
		notificationsService,
	)

	locationService := service.NewSurveillanceLocationService(
		deps.Logger,
		deps.Repositories.Surveillance.Regions,
		deps.Repositories.Surveillance.Districts,
		deps.Repositories.Surveillance.SubCounties,
		notificationsService,
		deps.Config,
	)

	facilityService := service.NewSurveillanceFacilityService(
		deps.Logger,
		deps.Repositories.Surveillance.Facilities,
		notificationsService,
	)

	facilityWeeklyMetricsService := service.NewSurveillanceFacilityWeeklyMetricsService(
		deps.Logger,
		deps.Repositories.Surveillance.FacilityMetrics,
		deps.Repositories.Surveillance.Imports,
	)

	weeklyStatusService := service.NewSurveillanceWeeklyStatusService(
		deps.Logger,
		deps.Repositories.Surveillance.WeeklyStatus,
	)

	surveillanceImportService := service.NewSurveillanceImportService(
		deps.Repositories.Surveillance.Imports,
	)

	alertsService := service.NewSurveillanceAlertService(
		deps.Logger,
		deps.Repositories.Surveillance.Alerts,
		deps.Repositories.Surveillance.Imports,
		notificationsService,
		deps.Config,
	)

	documentTemplateService := service.NewDocumentTemplateService(
		deps.Repositories.DocumentTemplates,
		deps.Repositories.DocumentTemplateSheets,
		deps.Repositories.DocumentTemplateColumns,
		notificationsService,
		deps.Config,
	)

	documentTemplateSheetService := service.NewDocumentTemplateSheetService(
		deps.Repositories.DocumentTemplateSheets,
		notificationsService,
		deps.Config,
	)

	documentTemplateColumnService := service.NewDocumentTemplateColumnService(
		deps.Repositories.DocumentTemplateColumns,
		notificationsService,
		deps.Config,
	)

	importService := importsvc.NewService(
		deps.Repositories.Documents,
		deps.Repositories.DocumentStockImports,
		deps.Repositories.Processes,
		deps.Repositories.DocumentFiles,
		deps.Repositories.Surveillance.Imports,
		documentTemplateService,
		facilityWeeklyMetricsService,
		weeklyStatusService,
		alertsService,
		deps.FileStorage,
		deps.Databases.Remote,
	)

	return services{
		Email:                   emailService,
		SMTP:                    smtpService,
		Notifications:           notificationsService,
		Auth:                    authService,
		Metrics:                 metricsService,
		Audit:                   auditService,
		StorageLocations:        storageLocationService,
		Sessions:                sessionService,
		Documents:               documentService,
		Clients:                 clientService,
		Users:                   userService,
		Announcements:           announcementService,
		Diseases:                diseaseService,
		EpiWeeks:                epiWeekService,
		Locations:               locationService,
		Facilities:              facilityService,
		FacilityWeeklyMetrics:   facilityWeeklyMetricsService,
		WeeklyStatus:            weeklyStatusService,
		SurveillanceImport:      surveillanceImportService,
		Alerts:                  alertsService,
		DocumentTemplates:       documentTemplateService,
		DocumentTemplateSheets:  documentTemplateSheetService,
		DocumentTemplateColumns: documentTemplateColumnService,
		Import:                  importService,
	}
}
