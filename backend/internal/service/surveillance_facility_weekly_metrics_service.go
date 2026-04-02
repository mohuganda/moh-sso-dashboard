package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
)

type facilityWeeklyMetricsPayload struct {
	Facility    string `json:"facility"`
	Disease     string `json:"disease"`
	EpiWeek     int32  `json:"epi_week"`
	Year        int32  `json:"year"`
	MetricValue int32  `json:"value"`
	Status      string `json:"status"`
}

type SurveillanceFacilityWeeklyMetricsService struct {
	log                       *logger.Logger
	facilityWeeklyMetricsRepo interfaces.FacilityWeeklyMetricsRepository
	surveillanceImportRepo    interfaces.ImportRepository
}

func NewSurveillanceFacilityWeeklyMetricsService(
	log *logger.Logger,
	facilityWeeklyMetricsRepo interfaces.FacilityWeeklyMetricsRepository,
	surveillanceImportRepo interfaces.ImportRepository,
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

func (s *SurveillanceFacilityWeeklyMetricsService) ProcessFacilityMetrics(
	ctx context.Context,
	batchID uuid.UUID,
) error {
	if batchID == uuid.Nil {
		return fmt.Errorf("batch id is required")
	}

	return s.facilityWeeklyMetricsRepo.WithTx(ctx, func(q db.Querier) error {
		rows, err := s.surveillanceImportRepo.ListImportRawRowsByBatch(ctx, batchID)
		if err != nil {
			return fmt.Errorf("list import raw rows by batch: %w", err)
		}

		var successRows int32
		var failedRows int32

		for _, raw := range rows {
			if err := s.processFacilityMetricRow(ctx, q, raw); err != nil {
				failedRows++

				if markErr := s.surveillanceImportRepo.MarkRawRowFailed(ctx, raw.ID, err.Error()); markErr != nil {
					return fmt.Errorf("mark raw row failed: %w", markErr)
				}

				continue
			}

			successRows++

			if err := s.surveillanceImportRepo.MarkRawRowProcessed(ctx, raw.ID); err != nil {
				return fmt.Errorf("mark raw row processed: %w", err)
			}
		}

		if err := s.surveillanceImportRepo.CompleteImportBatch(ctx,
			batchID,
			successRows,
			failedRows,
		); err != nil {
			return fmt.Errorf("complete batch: %w", err)
		}

		if failedRows > 0 {
			return nil
		}

		if err := s.surveillanceImportRepo.DeleteProcessedRawRows(ctx, batchID); err != nil {
			return fmt.Errorf("delete processed raw rows: %w", err)
		}

		return nil
	})
}

func (s *SurveillanceFacilityWeeklyMetricsService) processFacilityMetricRow(
	ctx context.Context,
	q db.Querier,
	raw db.SurveillanceImportRawRow,
) error {

	payload, err := parseFacilityWeeklyMetricsPayload(raw.Payload)
	if err != nil {
		return err
	}

	facility, err := q.GetFacilityByName(ctx, payload.Facility)
	if err != nil {
		return fmt.Errorf("find facility %q: %w", payload.Facility, err)
	}

	disease, err := q.GetDiseaseByName(ctx, payload.Disease)
	if err != nil {
		return fmt.Errorf("find disease %q: %w", payload.Disease, err)
	}

	epiWeek, err := q.GetEpiWeekByYearWeek(ctx, db.GetEpiWeekByYearWeekParams{
		EpiYear: payload.Year,
		EpiWeek: payload.EpiWeek,
	})
	if err != nil {
		return fmt.Errorf("find epi week %q: %w", epiWeek.EpiWeek, err)
	}

	_, err = s.facilityWeeklyMetricsRepo.UpsertDiseaseMetric(ctx, db.UpsertFacilityWeeklyDiseaseMetricParams{
		FacilityID:  facility.ID,
		DiseaseID:   disease.ID,
		EpiWeekID:   epiWeek.ID,
		MetricValue: string(payload.MetricValue),
	})
	if err != nil {
		return fmt.Errorf("upsert facility weekly metric: %w", err)
	}

	return nil
}

func parseFacilityWeeklyMetricsPayload(rawPayload []byte) (facilityWeeklyMetricsPayload, error) {
	var payload facilityWeeklyMetricsPayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return facilityWeeklyMetricsPayload{}, fmt.Errorf("invalid payload json: %w", err)

	}

	payload.Facility = strings.TrimSpace(payload.Facility)
	payload.Disease = strings.TrimSpace(payload.Disease)
	payload.Status = strings.TrimSpace(payload.Status)

	if payload.Facility == "" {
		return facilityWeeklyMetricsPayload{}, fmt.Errorf("facility is required")
	}

	if payload.Disease == "" {
		return facilityWeeklyMetricsPayload{}, fmt.Errorf("disease is required")
	}

	if payload.Status == "" {
		return facilityWeeklyMetricsPayload{}, fmt.Errorf("status is required")
	}

	if payload.Year <= 0 {
		return facilityWeeklyMetricsPayload{}, fmt.Errorf("year is required")
	}

	if payload.EpiWeek <= 0 {
		return facilityWeeklyMetricsPayload{}, fmt.Errorf("epi_week is required")
	}

	return payload, nil
}
