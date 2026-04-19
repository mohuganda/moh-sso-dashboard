package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
)

const alertImportSource = "csv_import"

type alertsPayload struct {
	Created     string `json:"created"`
	Narrative   string `json:"narrative"`
	District    string `json:"district"`
	Disease     string `json:"disease"`
	EpiWeek     int32  `json:"weeks"`
	SubmittedBy string `json:"submitted_by"`
}

type parsedAlertsPayload struct {
	CreatedAt   time.Time
	Narrative   string
	District    string
	Disease     string
	EpiWeek     int32
	Year        int32
	SubmittedBy string
}

type AlertListParams struct {
	EpiWeekID  uuid.UUID
	DiseaseID  uuid.UUID
	DistrictID uuid.UUID
	RegionID   uuid.UUID
}

type SurveillanceAlertService struct {
	log                    *logger.Logger
	alertRepo              interfaces.AlertRepository
	surveillanceImportRepo interfaces.ImportRepository
}

func NewSurveillanceAlertService(
	log *logger.Logger,
	alertRepo interfaces.AlertRepository,
	surveillanceImportRepo interfaces.ImportRepository,
) *SurveillanceAlertService {
	return &SurveillanceAlertService{
		log:                    log,
		alertRepo:              alertRepo,
		surveillanceImportRepo: surveillanceImportRepo,
	}
}

func (s *SurveillanceAlertService) GetAlertByID(ctx context.Context, id uuid.UUID) (db.Alert, error) {
	if err := requireUUID("alert id", id); err != nil {
		return db.Alert{}, err
	}

	s.log.Debug(ctx, "getting alert by id", "alert_id", id)

	item, err := s.alertRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error(ctx, "failed to get alert by id", "alert_id", id, "error", err)
		return db.Alert{}, err
	}

	return item, nil
}

func (s *SurveillanceAlertService) ListAlertsByDisease(
	ctx context.Context,
	diseaseID uuid.UUID,
) ([]db.ListAlertsByDiseaseRow, error) {
	if err := requireUUID("disease id", diseaseID); err != nil {
		return nil, err
	}

	s.log.Debug(ctx, "listing alerts by disease", "disease_id", diseaseID)

	items, err := s.alertRepo.ListByDisease(ctx, diseaseID)
	if err != nil {
		s.log.Error(ctx, "failed to list alerts by disease", "disease_id", diseaseID, "error", err)
		return nil, err
	}

	return items, nil
}

func (s *SurveillanceAlertService) ListAlertsByDistrict(
	ctx context.Context,
	districtID uuid.UUID,
) ([]db.ListAlertsByDistrictRow, error) {
	if err := requireUUID("district id", districtID); err != nil {
		return nil, err
	}

	s.log.Debug(ctx, "listing alerts by district", "district_id", districtID)

	items, err := s.alertRepo.ListByDistrict(ctx, districtID)
	if err != nil {
		s.log.Error(ctx, "failed to list alerts by district", "district_id", districtID, "error", err)
		return nil, err
	}

	return items, nil
}

func (s *SurveillanceAlertService) ListAlertsByWeek(
	ctx context.Context,
	epiWeekID uuid.UUID,
) ([]db.ListAlertsByWeekRow, error) {
	if err := requireUUID("epi week id", epiWeekID); err != nil {
		return nil, err
	}

	s.log.Debug(ctx, "listing alerts by week", "epi_week_id", epiWeekID)

	items, err := s.alertRepo.ListByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to list alerts by week", "epi_week_id", epiWeekID, "error", err)
		return nil, err
	}

	return items, nil
}

func (s *SurveillanceAlertService) ListAlerts(
	ctx context.Context,
	params db.ListAlertsParams,
) ([]db.ListAlertsRow, error) {
	s.log.Debug(ctx, "listing alerts")

	items, err := s.alertRepo.ListAlerts(ctx, params)
	if err != nil {
		s.log.Error(ctx, "failed to list alerts", "error", err)
		return nil, err
	}

	return items, nil
}
func (s *SurveillanceAlertService) ProcessAlerts(ctx context.Context, batchID uuid.UUID) error {
	if err := requireUUID("batch id", batchID); err != nil {
		return err
	}

	s.log.Info(ctx, "processing alert batch", "batch_id", batchID)

	return s.alertRepo.WithTx(ctx, func(q db.Querier) error {
		rows, err := s.surveillanceImportRepo.ListImportRawRowsByBatch(ctx, batchID)
		if err != nil {
			return fmt.Errorf("list import raw rows by batch: %w", err)
		}

		var successRows int32
		var failedRows int32

		for _, raw := range rows {
			if err := s.processAlertRow(ctx, q, raw); err != nil {
				failedRows++

				s.log.Warn(
					ctx,
					"failed to process alert row",
					"batch_id", batchID,
					"raw_row_id", raw.ID,
					"row_number", raw.RowNumber,
					"error", err,
				)

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
			"completed alert batch processing",
			"batch_id", batchID,
			"success_rows", successRows,
			"failed_rows", failedRows,
		)

		return nil
	})
}

func (s *SurveillanceAlertService) processAlertRow(
	ctx context.Context,
	q db.Querier,
	raw db.SurveillanceImportRawRow,
) error {
	payload, err := parseAlertsPayload(raw.Payload)
	if err != nil {
		return err
	}

	district, err := q.GetDistrictByName(ctx, payload.District)
	if err != nil {
		return fmt.Errorf("district not found: %s", payload.District)
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

	externalID := buildAlertExternalID(payload)

	params := db.CreateAlertParams{
		ExternalID: sql.NullString{
			String: externalID,
			Valid:  externalID != "",
		},
		DiseaseID: disease.ID,
		DistrictID: uuid.NullUUID{
			UUID:  district.ID,
			Valid: true,
		},
		EpiWeekID: uuid.NullUUID{
			UUID:  epiWeek.ID,
			Valid: true,
		},
		OccurredOn: sql.NullTime{
			Time:  payload.CreatedAt,
			Valid: true,
		},
		CreatedOn: sql.NullTime{
			Time:  payload.CreatedAt,
			Valid: true,
		},
		Narrative: payload.Narrative,
		SubmittedBy: sql.NullString{
			String: payload.SubmittedBy,
			Valid:  payload.SubmittedBy != "",
		},
		Status: db.AlertStatusNEW,
		SourceName: sql.NullString{
			String: alertImportSource,
			Valid:  true,
		},
	}

	if _, err := s.alertRepo.Create(ctx, params); err != nil {
		return fmt.Errorf(
			"create alert for district=%s disease=%s week=%d year=%d: %w",
			payload.District,
			payload.Disease,
			payload.EpiWeek,
			payload.Year,
			err,
		)
	}

	return nil
}

func parseAlertsPayload(rawPayload []byte) (parsedAlertsPayload, error) {
	var raw alertsPayload

	if err := json.Unmarshal(rawPayload, &raw); err != nil {
		return parsedAlertsPayload{}, fmt.Errorf("invalid payload json: %w", err)
	}

	raw.Created = strings.TrimSpace(raw.Created)
	raw.Narrative = strings.TrimSpace(raw.Narrative)
	raw.District = strings.TrimSpace(raw.District)
	raw.Disease = strings.TrimSpace(raw.Disease)
	raw.SubmittedBy = strings.TrimSpace(raw.SubmittedBy)

	if raw.Created == "" {
		return parsedAlertsPayload{}, fmt.Errorf("created is required")
	}

	createdAt, err := parseDateOnly(raw.Created)
	if err != nil {
		return parsedAlertsPayload{}, fmt.Errorf("invalid created date %q: %w", raw.Created, err)
	}

	if raw.District == "" {
		return parsedAlertsPayload{}, fmt.Errorf("district is required")
	}

	if raw.Disease == "" {
		return parsedAlertsPayload{}, fmt.Errorf("disease is required")
	}

	if raw.Narrative == "" {
		return parsedAlertsPayload{}, fmt.Errorf("narrative is required")
	}

	if raw.EpiWeek <= 0 {
		return parsedAlertsPayload{}, fmt.Errorf("weeks is required")
	}

	return parsedAlertsPayload{
		CreatedAt:   createdAt,
		Narrative:   raw.Narrative,
		District:    raw.District,
		Disease:     raw.Disease,
		EpiWeek:     raw.EpiWeek,
		Year:        int32(createdAt.Year()),
		SubmittedBy: raw.SubmittedBy,
	}, nil
}

func parseDateOnly(value string) (time.Time, error) {
	return time.Parse("2006-01-02", value)
}

func buildAlertExternalID(payload parsedAlertsPayload) string {
	parts := []string{
		strings.ToLower(strings.TrimSpace(payload.District)),
		strings.ToLower(strings.TrimSpace(payload.Disease)),
		fmt.Sprintf("%d", payload.Year),
		fmt.Sprintf("%d", payload.EpiWeek),
		strings.ToLower(strings.TrimSpace(payload.Narrative)),
	}

	for i := range parts {
		parts[i] = strings.ReplaceAll(parts[i], " ", "_")
	}

	return strings.Join(parts, "|")
}
