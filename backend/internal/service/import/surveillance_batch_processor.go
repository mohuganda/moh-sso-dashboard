package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
	"github.com/moh-sso-dashboard/internal/service"
)

type SurveillanceBatchProcessor struct {
	repo                   interfaces.ImportRepository
	facilityMetricsService *service.SurveillanceFacilityWeeklyMetricsService
	alertsService          *service.SurveillanceAlertService
}

type SurveillanceBatchProcessPayload struct {
	BatchID uuid.UUID `json:"batch_id"`
}

func NewSurveillanceBatchProcessor(
	repo interfaces.ImportRepository,
	facilityMetricsService *service.SurveillanceFacilityWeeklyMetricsService,
	alertsService *service.SurveillanceAlertService,
) *SurveillanceBatchProcessor {
	return &SurveillanceBatchProcessor{
		repo:                   repo,
		facilityMetricsService: facilityMetricsService,
		alertsService:          alertsService,
	}
}

func (s *SurveillanceBatchProcessor) Process(ctx context.Context, p db.Process) error {
	batchID, err := s.resolveBatchID(ctx, p)
	if err != nil {
		return err
	}

	return s.processBatch(ctx, batchID)
}

func (s *SurveillanceBatchProcessor) resolveBatchID(ctx context.Context, p db.Process) (uuid.UUID, error) {

	if p.DocumentID != uuid.Nil {
		batch, err := s.repo.GetLatestImportBatchByDocumentID(ctx, p.DocumentID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to resolve latest batch by document_id: %w", err)
		}
		return batch.ID, nil
	}

	return uuid.Nil, fmt.Errorf("batch_id is required")
}

func (s *SurveillanceBatchProcessor) processBatch(ctx context.Context, batchID uuid.UUID) error {
	batch, err := s.repo.GetImportBatchByID(ctx, batchID)
	if err != nil {
		return err
	}

	_, err = s.repo.UpdateImportBatchStatus(ctx, db.UpdateImportBatchStatusParams{
		ID:     batchID,
		Status: "PROCESSING",
		Notes: sql.NullString{
			String: fmt.Sprintf("Processing imported raw rows for %s", batch.DatasetType),
			Valid:  true,
		},
	})
	if err != nil {
		return err
	}

	var processErr error

	switch batch.DatasetType {
	case "facility_metrics":
		if s.facilityMetricsService == nil {
			processErr = fmt.Errorf("facility metrics service is not configured")
			break
		}
		processErr = s.facilityMetricsService.ProcessFacilityMetrics(ctx, batchID)

	case "alerts":
		if s.alertsService == nil {
			processErr = fmt.Errorf("alerts service is not configured")
			break
		}
		processErr = s.alertsService.ProcessAlerts(ctx, batchID)

	default:
		processErr = fmt.Errorf("unsupported dataset type: %s", batch.DatasetType)
	}

	if processErr != nil {
		_ = s.repo.FailImportBatch(ctx, batchID, processErr.Error())
		return processErr
	}

	successRows := batch.TotalRows - batch.FailedRows
	if successRows < 0 {
		successRows = 0
	}

	if err := s.repo.CompleteImportBatch(ctx, batchID, successRows, batch.FailedRows); err != nil {
		return err
	}

	return nil
}
