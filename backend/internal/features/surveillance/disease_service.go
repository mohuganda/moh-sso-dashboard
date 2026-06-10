package surveillance

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/config"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/model"
	sharedservice "github.com/moh-sso-dashboard/internal/service"
	"github.com/moh-sso-dashboard/internal/utils"
)

type DiseaseService struct {
	log           *logger.Logger
	diseaseRepo   DiseaseRepository
	notifications sharedservice.NotificationsService
	cfg           *config.Config
}

func NewDiseaseService(
	log *logger.Logger,
	diseaseRepo DiseaseRepository,
	notifications sharedservice.NotificationsService,
	cfg ...*config.Config,
) *DiseaseService {
	var appConfig *config.Config
	if len(cfg) > 0 {
		appConfig = cfg[0]
	}

	return &DiseaseService{
		log:           log,
		diseaseRepo:   diseaseRepo,
		notifications: notifications,
		cfg:           appConfig,
	}
}

func (s *DiseaseService) ListDiseases(ctx context.Context) ([]db.Disease, error) {
	if s == nil {
		return nil, errors.New("surveillance disease service is nil")
	}

	if s.diseaseRepo == nil {
		return nil, errors.New("disease repository is nil")
	}

	if s.log != nil {
		s.log.Debug(ctx, "listing diseases")
	}

	items, err := s.diseaseRepo.List(ctx)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to list diseases", "error", err)
		}

		return []db.Disease{}, fmt.Errorf("list diseases: %w", err)
	}

	return items, nil
}

func (s *DiseaseService) ListActiveDiseases(ctx context.Context) ([]db.Disease, error) {
	if s == nil {
		return nil, errors.New("surveillance disease service is nil")
	}

	if s.diseaseRepo == nil {
		return nil, errors.New("disease repository is nil")
	}

	if s.log != nil {
		s.log.Debug(ctx, "listing active diseases")
	}

	items, err := s.diseaseRepo.ListActive(ctx)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to list active diseases", "error", err)
		}

		return []db.Disease{}, fmt.Errorf("list active diseases: %w", err)
	}

	return items, nil
}

func (s *DiseaseService) GetDiseaseByID(
	ctx context.Context,
	id uuid.UUID,
) (db.Disease, error) {
	if s == nil {
		return db.Disease{}, errors.New("surveillance disease service is nil")
	}

	if s.diseaseRepo == nil {
		return db.Disease{}, errors.New("disease repository is nil")
	}

	if id == uuid.Nil {
		return db.Disease{}, errors.New("disease id is required")
	}

	if s.log != nil {
		s.log.Debug(ctx, "getting disease by id", "disease_id", id)
	}

	item, err := s.diseaseRepo.GetByID(ctx, id)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to get disease by id", "disease_id", id, "error", err)
		}

		return db.Disease{}, fmt.Errorf("get disease by id: %w", err)
	}

	return item, nil
}

func (s *DiseaseService) GetDiseaseByName(
	ctx context.Context,
	name string,
) (db.Disease, error) {
	if s == nil {
		return db.Disease{}, errors.New("surveillance disease service is nil")
	}

	if s.diseaseRepo == nil {
		return db.Disease{}, errors.New("disease repository is nil")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return db.Disease{}, errors.New("disease name is required")
	}

	if s.log != nil {
		s.log.Debug(ctx, "getting disease by name", "name", name)
	}

	item, err := s.diseaseRepo.GetByName(ctx, name)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to get disease by name", "name", name, "error", err)
		}

		return db.Disease{}, fmt.Errorf("get disease by name: %w", err)
	}

	return item, nil
}

func (s *DiseaseService) CreateDisease(
	ctx context.Context,
	arg db.CreateDiseaseParams,
) (db.Disease, error) {
	if s == nil {
		return db.Disease{}, errors.New("surveillance disease service is nil")
	}

	if s.diseaseRepo == nil {
		return db.Disease{}, errors.New("disease repository is nil")
	}

	arg.Name = strings.TrimSpace(arg.Name)
	if arg.Name == "" {
		return db.Disease{}, errors.New("disease name is required")
	}

	if s.log != nil {
		s.log.Info(ctx, "creating disease", "name", arg.Name)
	}

	item, err := s.diseaseRepo.Create(ctx, arg)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to create disease", "name", arg.Name, "error", err)
		}

		return db.Disease{}, fmt.Errorf("create disease: %w", err)
	}

	// In-app only.
	s.notify(ctx, model.Notification{
		Type:       "SURVEILLANCE_DISEASE_CREATED",
		Title:      "Disease created",
		Severity:   "info",
		Message:    "Surveillance disease created",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"disease_id": item.ID.String(),
			"name":       item.Name,
			"code":       nullStringValue(item.Code),
			"category":   nullStringValue(item.Category),
			"is_active":  item.IsActive,
		}),
	})

	return item, nil
}

func (s *DiseaseService) UpsertDisease(
	ctx context.Context,
	arg db.UpsertDiseaseParams,
) (db.Disease, error) {
	if s == nil {
		return db.Disease{}, errors.New("surveillance disease service is nil")
	}

	if s.diseaseRepo == nil {
		return db.Disease{}, errors.New("disease repository is nil")
	}

	arg.Name = strings.TrimSpace(arg.Name)
	if arg.Name == "" {
		return db.Disease{}, errors.New("disease name is required")
	}

	if s.log != nil {
		s.log.Info(ctx, "upserting disease", "name", arg.Name)
	}

	item, err := s.diseaseRepo.Upsert(ctx, arg)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to upsert disease", "name", arg.Name, "error", err)
		}

		return db.Disease{}, fmt.Errorf("upsert disease: %w", err)
	}

	// In-app only.
	s.notify(ctx, model.Notification{
		Type:       "SURVEILLANCE_DISEASE_UPSERTED",
		Title:      "Disease updated",
		Severity:   "info",
		Message:    "Surveillance disease created or updated",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"disease_id": item.ID.String(),
			"name":       item.Name,
			"code":       nullStringValue(item.Code),
			"category":   nullStringValue(item.Category),
			"is_active":  item.IsActive,
		}),
	})

	return item, nil
}

func (s *DiseaseService) SetDiseaseActiveState(
	ctx context.Context,
	arg db.SetDiseaseActiveStateParams,
) (db.Disease, error) {
	if s == nil {
		return db.Disease{}, errors.New("surveillance disease service is nil")
	}

	if s.diseaseRepo == nil {
		return db.Disease{}, errors.New("disease repository is nil")
	}

	if arg.ID == uuid.Nil {
		return db.Disease{}, errors.New("disease id is required")
	}

	if s.log != nil {
		s.log.Info(ctx, "setting disease active state", "disease_id", arg.ID, "is_active", arg.IsActive)
	}

	item, err := s.diseaseRepo.SetActiveState(ctx, arg)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to set disease active state", "disease_id", arg.ID, "error", err)
		}

		return db.Disease{}, fmt.Errorf("set disease active state: %w", err)
	}

	notificationType := "SURVEILLANCE_DISEASE_DISABLED"
	title := "Disease disabled"
	message := "Surveillance disease disabled"
	severity := "warning"

	if item.IsActive {
		notificationType = "SURVEILLANCE_DISEASE_ENABLED"
		title = "Disease enabled"
		message = "Surveillance disease enabled"
		severity = "info"
	}

	notification := model.Notification{
		Type:       notificationType,
		Title:      title,
		Severity:   severity,
		Message:    message,
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"disease_id": item.ID.String(),
			"name":       item.Name,
			"code":       nullStringValue(item.Code),
			"category":   nullStringValue(item.Category),
			"is_active":  item.IsActive,
		}),
	}

	// Email only when disabled. Enabling remains in-app only.
	if !item.IsActive {
		s.attachAdminEmailDelivery(
			&notification,
			"surveillance-disease-disabled",
			"Surveillance disease disabled",
			fmt.Sprintf("Surveillance disease %s was disabled.", item.Name),
			map[string]any{
				"Name":        s.systemAdminName(),
				"Platform":    s.platformName(),
				"DiseaseName": item.Name,
				"DiseaseCode": nullStringValue(item.Code),
				"Category":    nullStringValue(item.Category),
				"ActionURL":   s.adminSurveillanceURL(),
				"Details": fmt.Sprintf(
					"Disease ID: %s\nName: %s\nCode: %s\nCategory: %s\nIs Active: %v",
					item.ID.String(),
					item.Name,
					nullStringValue(item.Code),
					nullStringValue(item.Category),
					item.IsActive,
				),
			},
		)
	}

	s.notify(ctx, notification)

	return item, nil
}

func (s *DiseaseService) DeleteDisease(ctx context.Context, id uuid.UUID) error {
	if s == nil {
		return errors.New("surveillance disease service is nil")
	}

	if s.diseaseRepo == nil {
		return errors.New("disease repository is nil")
	}

	if id == uuid.Nil {
		return errors.New("disease id is required")
	}

	if s.log != nil {
		s.log.Info(ctx, "deleting disease", "disease_id", id)
	}

	item, _ := s.diseaseRepo.GetByID(ctx, id)

	if err := s.diseaseRepo.Delete(ctx, id); err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to delete disease", "disease_id", id, "error", err)
		}

		return fmt.Errorf("delete disease: %w", err)
	}

	if item.ID != uuid.Nil {
		notification := model.Notification{
			Type:       "SURVEILLANCE_DISEASE_DELETED",
			Title:      "Disease deleted",
			Severity:   "critical",
			Message:    "Surveillance disease deleted",
			TargetRole: "admin",
			Metadata: utils.MustJSON(map[string]any{
				"disease_id": item.ID.String(),
				"name":       item.Name,
				"code":       nullStringValue(item.Code),
				"category":   nullStringValue(item.Category),
				"is_active":  item.IsActive,
			}),
		}

		s.attachAdminEmailDelivery(
			&notification,
			"surveillance-disease-deleted",
			"Surveillance disease deleted",
			fmt.Sprintf("Surveillance disease %s was deleted.", item.Name),
			map[string]any{
				"Name":        s.systemAdminName(),
				"Platform":    s.platformName(),
				"DiseaseName": item.Name,
				"DiseaseCode": nullStringValue(item.Code),
				"Category":    nullStringValue(item.Category),
				"ActionURL":   s.adminSurveillanceURL(),
				"Details": fmt.Sprintf(
					"Disease ID: %s\nName: %s\nCode: %s\nCategory: %s\nWas Active: %v",
					item.ID.String(),
					item.Name,
					nullStringValue(item.Code),
					nullStringValue(item.Category),
					item.IsActive,
				),
			},
		)

		s.notify(ctx, notification)
	}

	return nil
}

func (s *DiseaseService) notify(
	ctx context.Context,
	notification model.Notification,
) {
	if s == nil || s.notifications == nil {
		return
	}

	if strings.TrimSpace(notification.TargetRole) == "" {
		notification.TargetRole = "admin"
	}

	if _, err := s.notifications.Notify(ctx, notification); err != nil {
		if s.log != nil {
			s.log.Error(
				ctx,
				"surveillance disease notification failed",
				"type", notification.Type,
				"error", err,
			)
		}
	}
}

func (s *DiseaseService) attachAdminEmailDelivery(
	notification *model.Notification,
	templateName string,
	subject string,
	textBody string,
	templateData map[string]any,
) {
	if notification == nil {
		return
	}

	adminEmail := strings.TrimSpace(s.systemAdminEmail())
	if adminEmail == "" {
		return
	}

	if templateData == nil {
		templateData = map[string]any{}
	}

	if _, ok := templateData["Name"]; !ok {
		templateData["Name"] = s.systemAdminName()
	}

	if _, ok := templateData["Platform"]; !ok {
		templateData["Platform"] = s.platformName()
	}

	if _, ok := templateData["ActionURL"]; !ok {
		templateData["ActionURL"] = s.adminSurveillanceURL()
	}

	notification.Deliveries = []model.NotificationDeliveryRequest{
		{
			Channel: model.NotificationChannelInApp,
			Recipient: map[string]any{
				"target_role": notification.TargetRole,
			},
			Payload: map[string]any{
				"title":    notification.Title,
				"message":  notification.Message,
				"type":     notification.Type,
				"severity": notification.Severity,
			},
			MaxAttempts: 1,
		},
		{
			Channel: model.NotificationChannelEmail,
			Recipient: map[string]any{
				"name":  s.systemAdminName(),
				"email": adminEmail,
			},
			TemplateName: templateName,
			TemplateData: templateData,
			Payload: map[string]any{
				"subject":   subject,
				"text_body": textBody,
			},
			MaxAttempts: 5,
		},
	}
}

func (s *DiseaseService) platformName() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.PlatformName) != "" {
		return strings.TrimSpace(s.cfg.Notification.PlatformName)
	}

	return "MOH Integrated Health Portal"
}

func (s *DiseaseService) systemAdminName() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.SystemAdminName) != "" {
		return strings.TrimSpace(s.cfg.Notification.SystemAdminName)
	}

	return "System Administrator"
}

func (s *DiseaseService) systemAdminEmail() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.SystemAdminEmail) != "" {
		return strings.TrimSpace(s.cfg.Notification.SystemAdminEmail)
	}

	return ""
}

func (s *DiseaseService) adminDashboardURL() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.AdminDashboardURL) != "" {
		return strings.TrimSpace(s.cfg.Notification.AdminDashboardURL)
	}

	return "http://localhost:3000/portal/admin/home"
}

func (s *DiseaseService) adminSurveillanceURL() string {
	base := strings.TrimRight(s.adminDashboardURL(), "/")

	if strings.HasSuffix(base, "portal/admin/home") {
		return strings.TrimSuffix(base, "portal/admin/home") + "/surveillance"
	}

	return base + "/surveillance"
}
