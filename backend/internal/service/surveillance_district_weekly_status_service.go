package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
)

type SurveillanceDistrictWeeklyStatusService struct {
	log                      *logger.Logger
	districtWeeklyStatusRepo interfaces.DistrictWeeklyStatusRepository
}

func NewSurveillanceDistrictWeeklyStatusService(
	log *logger.Logger,
	districtWeeklyStatusRepo interfaces.DistrictWeeklyStatusRepository,
) *SurveillanceDistrictWeeklyStatusService {
	return &SurveillanceDistrictWeeklyStatusService{
		log:                      log,
		districtWeeklyStatusRepo: districtWeeklyStatusRepo,
	}
}

func (s *SurveillanceDistrictWeeklyStatusService) ListDistrictWeeklyStatusesByWeek(
	ctx context.Context,
	epiWeekID uuid.UUID,
) ([]db.ListDistrictWeeklyDiseaseStatusesByWeekRow, error) {
	s.log.Debug(ctx, "listing district weekly statuses by epi week", "epi_week_id", epiWeekID)

	if epiWeekID == uuid.Nil {
		return []db.ListDistrictWeeklyDiseaseStatusesByWeekRow{}, errors.New("epi week id is required")
	}

	items, err := s.districtWeeklyStatusRepo.ListDiseaseStatusesByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to list district weekly statuses by epi week", "epi_week_id", epiWeekID, "error", err)
		return []db.ListDistrictWeeklyDiseaseStatusesByWeekRow{}, err
	}

	return items, nil
}

func (s *SurveillanceDistrictWeeklyStatusService) ListDistrictWeeklyStatusesByDistrictAndWeek(
	ctx context.Context,
	arg db.ListDistrictStatusesByDistrictAndWeekParams,
) ([]db.ListDistrictStatusesByDistrictAndWeekRow, error) {
	s.log.Debug(
		ctx,
		"listing district weekly statuses by district and epi week",
		"district_id", arg.DistrictID,
		"epi_week_id", arg.EpiWeekID,
	)

	if arg.DistrictID == uuid.Nil {
		return []db.ListDistrictStatusesByDistrictAndWeekRow{}, errors.New("district id is required")
	}

	if arg.EpiWeekID == uuid.Nil {
		return []db.ListDistrictStatusesByDistrictAndWeekRow{}, errors.New("epi week id is required")
	}

	items, err := s.districtWeeklyStatusRepo.ListByDistrictAndWeek(ctx, arg)
	if err != nil {
		s.log.Error(
			ctx,
			"failed to list district weekly statuses by district and epi week",
			"district_id", arg.DistrictID,
			"epi_week_id", arg.EpiWeekID,
			"error", err,
		)
		return []db.ListDistrictStatusesByDistrictAndWeekRow{}, err
	}

	return items, nil
}

func (s *SurveillanceDistrictWeeklyStatusService) ListDistrictWeeklyDiseaseStatusesByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListDistrictWeeklyDiseaseStatusesByWeekRow, error) {
	s.log.Debug(ctx, "listing district weekly disease statuses by week")

	items, err := s.districtWeeklyStatusRepo.ListDiseaseStatusesByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to list district weekly disease statuses by week", "error", err)
		return []db.ListDistrictWeeklyDiseaseStatusesByWeekRow{}, err
	}

	return items, nil
}

func (s *SurveillanceDistrictWeeklyStatusService) ListDistrictWeeklyIndicatorStatusesByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListDistrictWeeklyIndicatorStatusesByWeekRow, error) {
	s.log.Debug(ctx, "listing district weekly indicator statuses by week")

	items, err := s.districtWeeklyStatusRepo.ListIndicatorStatusesByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to list district weekly indicator statuses by week", "error", err)
		return []db.ListDistrictWeeklyIndicatorStatusesByWeekRow{}, err
	}

	return items, nil
}
