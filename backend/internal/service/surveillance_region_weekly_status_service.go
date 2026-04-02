package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
)

type regionStatusPayload struct {
	Week   int32    `json:"weeks"`
	Year   int32    `json:"year"`
	Region string   `json:"region"`
	Maroon []string `json:"maroon"`
	Red    []string `json:"red"`
	Yellow []string `json:"yellow"`
	Green  []string `json:"green"`
}

type SurveillanceRegionWeeklyStatusService struct {
	log                    *logger.Logger
	regionWeeklyStatusRepo interfaces.RegionWeeklyStatusRepository
	surveillanceImportRepo interfaces.ImportRepository
}

func NewSurveillanceRegionWeeklyStatusService(
	log *logger.Logger,
	regionWeeklyStatusRepo interfaces.RegionWeeklyStatusRepository,
	surveillanceImportRepo interfaces.ImportRepository,
) *SurveillanceRegionWeeklyStatusService {
	return &SurveillanceRegionWeeklyStatusService{
		log:                    log,
		regionWeeklyStatusRepo: regionWeeklyStatusRepo,
		surveillanceImportRepo: surveillanceImportRepo,
	}
}

func (s *SurveillanceRegionWeeklyStatusService) ListRegionWeeklyStatusesByWeek(
	ctx context.Context,
	epiWeekID uuid.UUID,
) ([]db.ListRegionWeeklyDiseaseStatusesByWeekRow, error) {
	if err := requireUUID("epi week id", epiWeekID); err != nil {
		return nil, err
	}

	s.log.Debug(ctx, "listing region weekly statuses by epi week", "epi_week_id", epiWeekID)

	items, err := s.regionWeeklyStatusRepo.ListDiseaseStatusesByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(
			ctx,
			"failed to list region weekly statuses by epi week",
			"epi_week_id", epiWeekID,
			"error", err,
		)
		return nil, err
	}

	return items, nil
}

func (s *SurveillanceRegionWeeklyStatusService) ProcessRegionWeeklyStatusesByWeek(
	ctx context.Context,
	batchID uuid.UUID,
) error {
	if err := requireUUID("batch id", batchID); err != nil {
		return err
	}

	s.log.Info(ctx, "processing region weekly status batch", "batch_id", batchID)

	return s.regionWeeklyStatusRepo.WithTx(ctx, func(q db.Querier) error {
		rows, err := s.surveillanceImportRepo.ListImportRawRowsByBatch(ctx, batchID)
		if err != nil {
			return fmt.Errorf("list raw rows by batch: %w", err)
		}

		var successRows int32
		var failedRows int32

		for _, row := range rows {
			if err := s.processRegionWeeklyStatusRow(ctx, q, row); err != nil {
				failedRows++

				s.log.Warn(
					ctx,
					"failed to process region weekly status row",
					"batch_id", batchID,
					"raw_row_id", row.ID,
					"row_number", row.RowNumber,
					"error", err,
				)

				if markErr := s.surveillanceImportRepo.MarkRawRowFailed(ctx, row.ID, err.Error()); markErr != nil {
					return fmt.Errorf("mark raw row failed: %w", markErr)
				}
				continue
			}

			successRows++

			if err := s.surveillanceImportRepo.MarkRawRowProcessed(ctx, row.ID); err != nil {
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
			"completed region weekly status batch processing",
			"batch_id", batchID,
			"success_rows", successRows,
			"failed_rows", failedRows,
		)

		return nil
	})
}

func (s *SurveillanceRegionWeeklyStatusService) processRegionWeeklyStatusRow(
	ctx context.Context,
	q db.Querier,
	row db.SurveillanceImportRawRow,
) error {
	payload, err := parseRegionStatusPayload(row.Payload)
	if err != nil {
		return err
	}

	region, err := q.GetRegionByName(ctx, payload.Region)
	if err != nil {
		return fmt.Errorf("region not found: %s", payload.Region)
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

			if _, err := s.regionWeeklyStatusRepo.UpsertDiseaseStatus(ctx, db.UpsertRegionWeeklyDiseaseStatusParams{
				RegionID:  region.ID,
				DiseaseID: disease.ID,
				EpiWeekID: epiWeek.ID,
				Status:    status,
			}); err != nil {
				return fmt.Errorf(
					"upsert region weekly disease status for region=%s disease=%s week=%d year=%d: %w",
					payload.Region,
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

func parseRegionStatusPayload(rawPayload []byte) (regionStatusPayload, error) {
	var payload regionStatusPayload

	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return regionStatusPayload{}, fmt.Errorf("invalid payload json: %w", err)
	}

	payload.Region = strings.TrimSpace(payload.Region)
	payload.Maroon = normalizeDiseaseNames(payload.Maroon)
	payload.Red = normalizeDiseaseNames(payload.Red)
	payload.Yellow = normalizeDiseaseNames(payload.Yellow)
	payload.Green = normalizeDiseaseNames(payload.Green)

	if payload.Region == "" {
		return regionStatusPayload{}, fmt.Errorf("region is required")
	}

	if payload.Year <= 0 {
		return regionStatusPayload{}, fmt.Errorf("year is required")
	}

	if payload.Week <= 0 {
		return regionStatusPayload{}, fmt.Errorf("weeks is required")
	}

	return payload, nil
}
