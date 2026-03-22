package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
)

type SurveillanceRegionWeeklyStatusService struct {
	log                    *logger.Logger
	regionWeeklyStatusRepo interfaces.RegionWeeklyStatusRepository
}

func NewSurveillanceRegionWeeklyStatusService(
	log *logger.Logger,
	regionWeeklyStatusRepo interfaces.RegionWeeklyStatusRepository,
) *SurveillanceRegionWeeklyStatusService {
	return &SurveillanceRegionWeeklyStatusService{
		log:                    log,
		regionWeeklyStatusRepo: regionWeeklyStatusRepo,
	}
}

func (s *SurveillanceRegionWeeklyStatusService) ListRegionWeeklyStatusesByWeek(
	ctx context.Context,
	epiWeekID uuid.UUID,
) ([]db.ListRegionWeeklyDiseaseStatusesByWeekRow, error) {
	s.log.Debug(ctx, "listing region weekly statuses by epi week", "epi_week_id", epiWeekID)

	if epiWeekID == uuid.Nil {
		return []db.ListRegionWeeklyDiseaseStatusesByWeekRow{}, errors.New("epi week id is required")
	}

	items, err := s.regionWeeklyStatusRepo.ListDiseaseStatusesByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to list region weekly statuses by epi week", "epi_week_id", epiWeekID, "error", err)
		return []db.ListRegionWeeklyDiseaseStatusesByWeekRow{}, err
	}

	return items, nil
}
