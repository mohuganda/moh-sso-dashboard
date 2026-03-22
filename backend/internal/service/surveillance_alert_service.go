package service

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
)

type SurveillanceAlertService struct {
	log       *logger.Logger
	alertRepo interfaces.AlertRepository
}

func NewSurveillanceAlertService(
	log *logger.Logger,
	alertRepo interfaces.AlertRepository,
) *SurveillanceAlertService {
	return &SurveillanceAlertService{
		log:       log,
		alertRepo: alertRepo,
	}
}

func (s *SurveillanceAlertService) GetAlertByID(ctx context.Context, id uuid.UUID) (db.Alert, error) {
	s.log.Debug(ctx, "getting alert by id", "alert_id", id)

	item, err := s.alertRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error(ctx, "failed to get alert by id", "alert_id", id, "error", err)
		return db.Alert{}, err
	}

	return item, nil
}

func (s *SurveillanceAlertService) ListAlertsByDisease(ctx context.Context, diseaseID uuid.UUID) ([]db.ListAlertsByDiseaseRow, error) {
	s.log.Debug(ctx, "listing alerts by disease", "disease_id", diseaseID)

	items, err := s.alertRepo.ListByDisease(ctx, diseaseID)
	if err != nil {
		s.log.Error(ctx, "failed to list alerts by disease", "disease_id", diseaseID, "error", err)
		return []db.ListAlertsByDiseaseRow{}, err
	}

	return items, nil
}

func (s *SurveillanceAlertService) ListAlertsByDistrict(ctx context.Context, districtID uuid.UUID) ([]db.ListAlertsByDistrictRow, error) {
	s.log.Debug(ctx, "listing alerts by district")

	items, err := s.alertRepo.ListByDistrict(ctx, districtID)
	if err != nil {
		s.log.Error(ctx, "failed to list alerts by district", "error", err)
		return []db.ListAlertsByDistrictRow{}, err
	}

	return items, nil
}

func (s *SurveillanceAlertService) ListAlertsByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListAlertsByWeekRow, error) {
	s.log.Debug(ctx, "listing alerts by week")

	items, err := s.alertRepo.ListByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to list alerts by week", "error", err)
		return []db.ListAlertsByWeekRow{}, err
	}

	return items, nil
}
