package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
)

type SurveillanceFacilityWeeklyMetricsService struct {
	log                       *logger.Logger
	facilityWeeklyMetricsRepo interfaces.FacilityWeeklyMetricsRepository
}

func NewSurveillanceFacilityWeeklyMetricsService(
	log *logger.Logger,
	facilityWeeklyMetricsRepo interfaces.FacilityWeeklyMetricsRepository,
) *SurveillanceFacilityWeeklyMetricsService {
	return &SurveillanceFacilityWeeklyMetricsService{
		log:                       log,
		facilityWeeklyMetricsRepo: facilityWeeklyMetricsRepo,
	}
}

func (s *SurveillanceFacilityWeeklyMetricsService) ListFacilityWeeklyMetricsByWeek(
	ctx context.Context,
	epiWeekID uuid.UUID,
) ([]db.ListFacilityWeeklyDiseaseMetricsByWeekRow, error) {
	s.log.Debug(ctx, "listing facility weekly metrics by epi week", "epi_week_id", epiWeekID)

	if epiWeekID == uuid.Nil {
		return []db.ListFacilityWeeklyDiseaseMetricsByWeekRow{}, errors.New("epi week id is required")
	}

	items, err := s.facilityWeeklyMetricsRepo.ListDiseaseMetricsByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to list facility weekly metrics by epi week", "epi_week_id", epiWeekID, "error", err)
		return []db.ListFacilityWeeklyDiseaseMetricsByWeekRow{}, err
	}

	return items, nil
}

func (s *SurveillanceFacilityWeeklyMetricsService) ListFacilityWeeklyMetricsByFacility(
	ctx context.Context,
	facilityID uuid.UUID,
) ([]db.ListFacilityMetricsByFacilityRow, error) {
	s.log.Debug(ctx, "listing facility weekly metrics by facility", "facility_id", facilityID)

	if facilityID == uuid.Nil {
		return []db.ListFacilityMetricsByFacilityRow{}, errors.New("facility id is required")
	}

	items, err := s.facilityWeeklyMetricsRepo.ListByFacility(ctx, facilityID)
	if err != nil {
		s.log.Error(ctx, "failed to list facility weekly metrics by facility", "facility_id", facilityID, "error", err)
		return []db.ListFacilityMetricsByFacilityRow{}, err
	}

	return items, nil
}

func (s *SurveillanceFacilityWeeklyMetricsService) UpsertFacilityWeeklyIndicatorMetric(ctx context.Context, arg db.UpsertFacilityWeeklyIndicatorMetricParams) (db.FacilityWeeklyMetric, error) {
	s.log.Debug(ctx, "upserting facility weekly indicator metric")

	item, err := s.facilityWeeklyMetricsRepo.UpsertIndicatorMetric(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to upsert facility weekly indicator metric", "error", err)
		return db.FacilityWeeklyMetric{}, err
	}

	return item, nil
}

func (s *SurveillanceFacilityWeeklyMetricsService) GetFacilityMetricBySourceRecordID(ctx context.Context, sourceRecordID string) (db.FacilityWeeklyMetric, error) {
	s.log.Debug(ctx, "getting facility metric by source record id", "source_record_id", sourceRecordID)

	item, err := s.facilityWeeklyMetricsRepo.GetBySourceRecordID(ctx, sourceRecordID)
	if err != nil {
		s.log.Error(ctx, "failed to get facility metric by source record id", "source_record_id", sourceRecordID, "error", err)
		return db.FacilityWeeklyMetric{}, err
	}

	return item, nil
}

func (s *SurveillanceFacilityWeeklyMetricsService) ListFacilityWeeklyDiseaseMetricsByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListFacilityWeeklyDiseaseMetricsByWeekRow, error) {
	s.log.Debug(ctx, "listing facility weekly disease metrics by week")

	items, err := s.facilityWeeklyMetricsRepo.ListDiseaseMetricsByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to list facility weekly disease metrics by week", "error", err)
		return []db.ListFacilityWeeklyDiseaseMetricsByWeekRow{}, err
	}

	return items, nil
}

func (s *SurveillanceFacilityWeeklyMetricsService) ListFacilityWeeklyIndicatorMetricsByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListFacilityWeeklyIndicatorMetricsByWeekRow, error) {
	s.log.Debug(ctx, "listing facility weekly indicator metrics by week")

	items, err := s.facilityWeeklyMetricsRepo.ListIndicatorMetricsByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to list facility weekly indicator metrics by week", "error", err)
		return []db.ListFacilityWeeklyIndicatorMetricsByWeekRow{}, err
	}

	return items, nil
}

func (s *SurveillanceFacilityWeeklyMetricsService) UpsertFacilityWeeklyDiseaseMetric(ctx context.Context, arg db.UpsertFacilityWeeklyDiseaseMetricParams) (db.FacilityWeeklyMetric, error) {
	s.log.Debug(ctx, "upserting facility weekly disease metric")

	item, err := s.facilityWeeklyMetricsRepo.UpsertDiseaseMetric(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to upsert facility weekly disease metric", "error", err)
		return db.FacilityWeeklyMetric{}, err
	}

	return item, nil
}

func (s *SurveillanceFacilityWeeklyMetricsService) ListFacilityMetricsByFacility(ctx context.Context, facilityID uuid.UUID) ([]db.ListFacilityMetricsByFacilityRow, error) {
	s.log.Debug(ctx, "listing facility metrics by facility", "facility_id", facilityID)

	items, err := s.facilityWeeklyMetricsRepo.ListByFacility(ctx, facilityID)
	if err != nil {
		s.log.Error(ctx, "failed to list facility metrics by facility", "facility_id", facilityID, "error", err)
		return []db.ListFacilityMetricsByFacilityRow{}, err
	}

	return items, nil
}
