package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
	"github.com/moh-sso-dashboard/internal/utils"
)

type SurveillanceDiseaseService struct {
	log           *logger.Logger
	diseaseRepo   interfaces.DiseaseRepository
	notifications NotificationsService
}

func NewSurveillanceDiseaseService(
	log *logger.Logger,
	diseaseRepo interfaces.DiseaseRepository,
	notifications ...NotificationsService,
) *SurveillanceDiseaseService {
	var notificationSvc NotificationsService
	if len(notifications) > 0 {
		notificationSvc = notifications[0]
	}

	return &SurveillanceDiseaseService{
		log:           log,
		diseaseRepo:   diseaseRepo,
		notifications: notificationSvc,
	}
}

func (s *SurveillanceDiseaseService) ListDiseases(ctx context.Context) ([]db.Disease, error) {
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

func (s *SurveillanceDiseaseService) ListActiveDiseases(ctx context.Context) ([]db.Disease, error) {
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

func (s *SurveillanceDiseaseService) GetDiseaseByID(
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

func (s *SurveillanceDiseaseService) GetDiseaseByName(
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

func (s *SurveillanceDiseaseService) CreateDisease(
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

func (s *SurveillanceDiseaseService) UpsertDisease(
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

func (s *SurveillanceDiseaseService) SetDiseaseActiveState(
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

	s.notify(ctx, model.Notification{
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
	})

	return item, nil
}

func (s *SurveillanceDiseaseService) DeleteDisease(ctx context.Context, id uuid.UUID) error {
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
		s.notify(ctx, model.Notification{
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
		})
	}

	return nil
}

func (s *SurveillanceDiseaseService) notify(
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
