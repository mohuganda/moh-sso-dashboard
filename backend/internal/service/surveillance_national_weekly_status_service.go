package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
)

type SurveillanceNationalWeeklyStatusService struct {
	log                      *logger.Logger
	nationalWeeklyStatusRepo interfaces.NationalWeeklyStatusRepository
}

func NewSurveillanceNationalWeeklyStatusService(
	log *logger.Logger,
	nationalWeeklyStatusRepo interfaces.NationalWeeklyStatusRepository,
) *SurveillanceNationalWeeklyStatusService {
	return &SurveillanceNationalWeeklyStatusService{
		log:                      log,
		nationalWeeklyStatusRepo: nationalWeeklyStatusRepo,
	}
}

func (s *SurveillanceNationalWeeklyStatusService) ListNationalWeeklyStatusesByWeek(
	ctx context.Context,
	epiWeekID uuid.UUID,
) ([]db.ListNationalWeeklyDiseaseStatusesByWeekRow, error) {
	s.log.Debug(ctx, "listing national weekly statuses by epi week", "epi_week_id", epiWeekID)

	if epiWeekID == uuid.Nil {
		return []db.ListNationalWeeklyDiseaseStatusesByWeekRow{}, errors.New("epi week id is required")
	}

	items, err := s.nationalWeeklyStatusRepo.ListDiseaseStatusesByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to list national weekly statuses by epi week", "epi_week_id", epiWeekID, "error", err)
		return []db.ListNationalWeeklyDiseaseStatusesByWeekRow{}, err
	}

	return items, nil
}

func (s *SurveillanceNationalWeeklyStatusService) ListNationalWeeklyIndicatorStatusesByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListNationalWeeklyIndicatorStatusesByWeekRow, error) {
	s.log.Debug(ctx, "listing national weekly indicator statuses by week")

	items, err := s.nationalWeeklyStatusRepo.ListIndicatorStatusesByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to list national weekly indicator statuses by week", "error", err)
		return []db.ListNationalWeeklyIndicatorStatusesByWeekRow{}, err
	}

	return items, nil
}
