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

const nationalStatusImportSource = "csv_import"

type nationalStatusPayload struct {
	Disease string `json:"disease"`
	EpiWeek int32  `json:"epi_week"`
	Year    int32  `json:"year"`
	Status  string `json:"status"`
}

type SurveillanceNationalWeeklyStatusService struct {
	log                      *logger.Logger
	nationalWeeklyStatusRepo interfaces.NationalWeeklyStatusRepository
	surveillanceImportRepo   interfaces.ImportRepository
}

func NewSurveillanceNationalWeeklyStatusService(
	log *logger.Logger,
	nationalWeeklyStatusRepo interfaces.NationalWeeklyStatusRepository,
	surveillanceImportRepo interfaces.ImportRepository,

) *SurveillanceNationalWeeklyStatusService {
	return &SurveillanceNationalWeeklyStatusService{
		log:                      log,
		nationalWeeklyStatusRepo: nationalWeeklyStatusRepo,
		surveillanceImportRepo:   surveillanceImportRepo,
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

func (s *SurveillanceNationalWeeklyStatusService) ProcessNationalWeeklyStatuses(
	ctx context.Context,
	batchID uuid.UUID,
) error {
	return s.nationalWeeklyStatusRepo.WithTx(ctx, func(q db.Querier) error {
		rows, err := s.surveillanceImportRepo.ListImportRawRowsByBatch(ctx, batchID)
		if err != nil {
			return fmt.Errorf("list raw rows by batch: %w", err)
		}

		var successRows int32
		var failedRows int32

		for _, rawRow := range rows {
			if err := s.processNationalStatusRow(ctx, q, rawRow); err != nil {
				failedRows++

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

		if failedRows > 0 {
			return nil
		}

		if err := s.surveillanceImportRepo.DeleteProcessedRawRows(ctx, batchID); err != nil {
			return fmt.Errorf("delete processed raw rows: %w", err)
		}

		return nil
	})
}

func (s *SurveillanceNationalWeeklyStatusService) processNationalStatusRow(
	ctx context.Context,
	q db.Querier,
	rawRow db.SurveillanceImportRawRow,
) error {
	payload, err := parseNationalStatusPayload(rawRow.Payload)
	if err != nil {
		return err
	}

	disease, err := q.GetDiseaseByName(ctx, payload.Disease)
	if err != nil {
		return fmt.Errorf("disease not found: %s", payload.Disease)
	}

	epiWeek, err := q.GetEpiWeekByYearWeek(ctx, db.GetEpiWeekByYearWeekParams{
		EpiYear: payload.Year,
		EpiWeek: payload.EpiWeek,
	})
	if err != nil {
		return fmt.Errorf("epi week not found: year=%d week=%d", payload.Year, payload.EpiWeek)
	}

	status := strings.TrimSpace(rawRow.Status)
	if status == "" {
		return fmt.Errorf("status is required")
	}

	if _, err := s.nationalWeeklyStatusRepo.UpsertDiseaseStatus(ctx, db.UpsertNationalWeeklyDiseaseStatusParams{
		DiseaseID: disease.ID,
		EpiWeekID: epiWeek.ID,
		Status:    db.RiskLevel(status),
	}); err != nil {
		return fmt.Errorf("upsert national weekly status: %w", err)
	}

	return nil
}

func parseNationalStatusPayload(rawPayload []byte) (nationalStatusPayload, error) {
	var payload nationalStatusPayload

	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return nationalStatusPayload{}, fmt.Errorf("invalid payload json: %w", err)
	}

	payload.Disease = strings.TrimSpace(payload.Disease)
	payload.Status = strings.TrimSpace(payload.Status)

	if payload.Disease == "" {
		return nationalStatusPayload{}, fmt.Errorf("disease is required")
	}

	if payload.Status == "" {
		return nationalStatusPayload{}, fmt.Errorf("status is required")
	}

	if payload.Year <= 0 {
		return nationalStatusPayload{}, fmt.Errorf("year is required")
	}

	if payload.EpiWeek <= 0 {
		return nationalStatusPayload{}, fmt.Errorf("epi_week is required")
	}

	return payload, nil
}
