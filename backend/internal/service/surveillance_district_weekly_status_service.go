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

type districtStatusPayload struct {
	Week     int32    `json:"weeks"`
	Year     int32    `json:"year"`
	District string   `json:"district"`
	Maroon   []string `json:"maroon"`
	Red      []string `json:"red"`
	Yellow   []string `json:"yellow"`
	Green    []string `json:"green"`
}

type SurveillanceDistrictWeeklyStatusService struct {
	log                      *logger.Logger
	districtWeeklyStatusRepo interfaces.DistrictWeeklyStatusRepository
	surveillanceImportRepo   interfaces.ImportRepository
}

func NewSurveillanceDistrictWeeklyStatusService(
	log *logger.Logger,
	districtWeeklyStatusRepo interfaces.DistrictWeeklyStatusRepository,
	surveillanceImportRepo interfaces.ImportRepository,
) *SurveillanceDistrictWeeklyStatusService {
	return &SurveillanceDistrictWeeklyStatusService{
		log:                      log,
		districtWeeklyStatusRepo: districtWeeklyStatusRepo,
		surveillanceImportRepo:   surveillanceImportRepo,
	}
}

func (s *SurveillanceDistrictWeeklyStatusService) ListDistrictWeeklyStatusesByWeek(
	ctx context.Context,
	epiWeekID uuid.UUID,
) ([]db.ListDistrictWeeklyDiseaseStatusesByWeekRow, error) {
	if err := requireUUID("epi week id", epiWeekID); err != nil {
		return nil, err
	}

	s.log.Debug(ctx, "listing district weekly statuses by epi week", "epi_week_id", epiWeekID)

	items, err := s.districtWeeklyStatusRepo.ListDiseaseStatusesByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(
			ctx,
			"failed to list district weekly statuses by epi week",
			"epi_week_id", epiWeekID,
			"error", err,
		)
		return nil, err
	}

	return items, nil
}

func (s *SurveillanceDistrictWeeklyStatusService) ListDistrictWeeklyStatusesByDistrictAndWeek(
	ctx context.Context,
	arg db.ListDistrictStatusesByDistrictAndWeekParams,
) ([]db.ListDistrictStatusesByDistrictAndWeekRow, error) {
	if err := requireUUID("district id", arg.DistrictID); err != nil {
		return nil, err
	}
	if err := requireUUID("epi week id", arg.EpiWeekID); err != nil {
		return nil, err
	}

	s.log.Debug(
		ctx,
		"listing district weekly statuses by district and epi week",
		"district_id", arg.DistrictID,
		"epi_week_id", arg.EpiWeekID,
	)

	items, err := s.districtWeeklyStatusRepo.ListByDistrictAndWeek(ctx, arg)
	if err != nil {
		s.log.Error(
			ctx,
			"failed to list district weekly statuses by district and epi week",
			"district_id", arg.DistrictID,
			"epi_week_id", arg.EpiWeekID,
			"error", err,
		)
		return nil, err
	}

	return items, nil
}

func (s *SurveillanceDistrictWeeklyStatusService) ListDistrictWeeklyDiseaseStatusesByWeek(
	ctx context.Context,
	epiWeekID uuid.UUID,
) ([]db.ListDistrictWeeklyDiseaseStatusesByWeekRow, error) {
	if err := requireUUID("epi week id", epiWeekID); err != nil {
		return nil, err
	}

	s.log.Debug(ctx, "listing district weekly disease statuses by week", "epi_week_id", epiWeekID)

	items, err := s.districtWeeklyStatusRepo.ListDiseaseStatusesByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(
			ctx,
			"failed to list district weekly disease statuses by week",
			"epi_week_id", epiWeekID,
			"error", err,
		)
		return nil, err
	}

	return items, nil
}

func (s *SurveillanceDistrictWeeklyStatusService) ListDistrictWeeklyIndicatorStatusesByWeek(
	ctx context.Context,
	epiWeekID uuid.UUID,
) ([]db.ListDistrictWeeklyIndicatorStatusesByWeekRow, error) {
	if err := requireUUID("epi week id", epiWeekID); err != nil {
		return nil, err
	}

	s.log.Debug(ctx, "listing district weekly indicator statuses by week", "epi_week_id", epiWeekID)

	items, err := s.districtWeeklyStatusRepo.ListIndicatorStatusesByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(
			ctx,
			"failed to list district weekly indicator statuses by week",
			"epi_week_id", epiWeekID,
			"error", err,
		)
		return nil, err
	}

	return items, nil
}

func (s *SurveillanceDistrictWeeklyStatusService) ProcessDistrictStatus(
	ctx context.Context,
	batchID uuid.UUID,
) error {
	if err := requireUUID("batch id", batchID); err != nil {
		return err
	}

	s.log.Info(ctx, "processing district weekly status batch", "batch_id", batchID)

	return s.districtWeeklyStatusRepo.WithTx(ctx, func(q db.Querier) error {
		rows, err := s.surveillanceImportRepo.ListImportRawRowsByBatch(ctx, batchID)
		if err != nil {
			return fmt.Errorf("list raw rows by batch: %w", err)
		}

		var successRows int32
		var failedRows int32

		for _, rawRow := range rows {
			if err := s.processDistrictStatusRow(ctx, q, rawRow); err != nil {
				failedRows++

				s.log.Warn(
					ctx,
					"failed to process district status row",
					"batch_id", batchID,
					"raw_row_id", rawRow.ID,
					"row_number", rawRow.RowNumber,
					"error", err,
				)

				if markErr := s.surveillanceImportRepo.MarkRawRowFailed(ctx, rawRow.ID, err.Error()); markErr != nil {
					return fmt.Errorf("mark raw row failed: %w", markErr)
				}
				continue
			}

			successRows++

			if err := s.surveillanceImportRepo.MarkRawRowProcessed(ctx, rawRow.ID); err != nil {
				return fmt.Errorf("mark raw row processed: %w", err)
			}
		}

		if err := s.surveillanceImportRepo.CompleteImportBatch(ctx, batchID, successRows, failedRows); err != nil {
			return fmt.Errorf("complete batch: %w", err)
		}

		if failedRows == 0 {
			if err := s.surveillanceImportRepo.DeleteProcessedRawRows(ctx, batchID); err != nil {
				return fmt.Errorf("delete processed raw rows: %w", err)
			}
		}

		s.log.Info(
			ctx,
			"completed district weekly status batch processing",
			"batch_id", batchID,
			"success_rows", successRows,
			"failed_rows", failedRows,
		)

		return nil
	})
}

func (s *SurveillanceDistrictWeeklyStatusService) processDistrictStatusRow(
	ctx context.Context,
	q db.Querier,
	rawRow db.SurveillanceImportRawRow,
) error {
	payload, err := parseDistrictStatusPayload(rawRow.Payload)
	if err != nil {
		return err
	}

	district, err := q.GetDistrictByName(ctx, payload.District)
	if err != nil {
		return fmt.Errorf("district not found: %s", payload.District)
	}

	epiWeek, err := q.GetEpiWeekByYearWeek(ctx, db.GetEpiWeekByYearWeekParams{
		EpiYear: payload.Year,
		EpiWeek: payload.Week,
	})
	if err != nil {
		return fmt.Errorf("epi week not found: year=%d week=%d", payload.Year, payload.Week)
	}

	statusDiseaseMap := map[db.RiskLevel][]string{
		db.RiskLevelMAROON: payload.Maroon,
		db.RiskLevelRED:    payload.Red,
		db.RiskLevelYELLOW: payload.Yellow,
		db.RiskLevelGREEN:  payload.Green,
	}

	for status, diseases := range statusDiseaseMap {
		for _, diseaseName := range normalizeDiseaseNames(diseases) {
			disease, err := q.GetDiseaseByName(ctx, diseaseName)
			if err != nil {
				return fmt.Errorf("disease not found: %s", diseaseName)
			}

			if _, err := s.districtWeeklyStatusRepo.UpsertDiseaseStatus(ctx, db.UpsertDistrictWeeklyDiseaseStatusParams{
				DistrictID: district.ID,
				DiseaseID:  disease.ID,
				EpiWeekID:  epiWeek.ID,
				Status:     status,
			}); err != nil {
				return fmt.Errorf(
					"upsert district weekly disease status for district=%s disease=%s week=%d year=%d: %w",
					payload.District,
					diseaseName,
					payload.Week,
					payload.Year,
					err,
				)
			}
		}
	}

	return nil
}

func parseDistrictStatusPayload(rawPayload []byte) (districtStatusPayload, error) {
	var payload districtStatusPayload

	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return districtStatusPayload{}, fmt.Errorf("invalid payload json: %w", err)
	}

	payload.District = strings.TrimSpace(payload.District)
	payload.Maroon = normalizeDiseaseNames(payload.Maroon)
	payload.Red = normalizeDiseaseNames(payload.Red)
	payload.Yellow = normalizeDiseaseNames(payload.Yellow)
	payload.Green = normalizeDiseaseNames(payload.Green)

	if payload.District == "" {
		return districtStatusPayload{}, errors.New("district is required")
	}

	if payload.Year <= 0 {
		return districtStatusPayload{}, errors.New("year is required")
	}

	if payload.Week <= 0 {
		return districtStatusPayload{}, errors.New("weeks is required")
	}

	if len(payload.Maroon) == 0 &&
		len(payload.Red) == 0 &&
		len(payload.Yellow) == 0 &&
		len(payload.Green) == 0 {
		return districtStatusPayload{}, errors.New("at least one disease status bucket is required")
	}

	return payload, nil
}

func normalizeDiseaseNames(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))

	for _, value := range values {
		name := strings.TrimSpace(value)
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		result = append(result, name)
	}

	return result
}

func requireUUID(name string, value uuid.UUID) error {
	if value == uuid.Nil {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}
