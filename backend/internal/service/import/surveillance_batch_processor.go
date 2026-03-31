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
	repo                        interfaces.ImportRepository
	facilityMetricsService      *service.SurveillanceFacilityWeeklyMetricsService
	districtWeeklyStatusService *service.SurveillanceDistrictWeeklyStatusService
	regionWeeklyStatusService   *service.SurveillanceRegionWeeklyStatusService
	nationalStatusService       *service.SurveillanceNationalWeeklyStatusService
	alertsService               *service.SurveillanceAlertService
}

type SurveillanceBatchProcessPayload struct {
	BatchID uuid.UUID `json:"batch_id"`
}

func NewSurveillanceBatchProcessor(
	repo interfaces.ImportRepository,
	facilityMetricsService *service.SurveillanceFacilityWeeklyMetricsService,
	districtWeeklyStatusService *service.SurveillanceDistrictWeeklyStatusService,
	regionWeeklyStatusService *service.SurveillanceRegionWeeklyStatusService,
	nationalStatusService *service.SurveillanceNationalWeeklyStatusService,
	alertsService *service.SurveillanceAlertService,
) *SurveillanceBatchProcessor {
	return &SurveillanceBatchProcessor{
		repo:                   repo,
		facilityMetricsService: facilityMetricsService,
	}
}

func (s *SurveillanceBatchProcessor) Process(
	ctx context.Context,
	p db.Process,
) error {
	var payload SurveillanceBatchProcessPayload

	if payload.BatchID == uuid.Nil {
		return fmt.Errorf("batch_id is required")
	}

	return s.processBatch(ctx, payload.BatchID)
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
			String: "Processing imported raw rows",
			Valid:  true,
		},
	})
	if err != nil {
		return err
	}

	var processErr error

	switch batch.DatasetType {
	case "facility_metrics":
		processErr = s.facilityMetricsService.ProcessFacilityMetrics(ctx, batchID)
	case "district_status":
		processErr = s.districtWeeklyStatusService.ProcessDistrictStatus(ctx, batchID)
	case "region_status":
		processErr = s.regionWeeklyStatusService.ProcessRegionWeeklyStatusesByWeek(ctx, batchID)
	case "national_status":
		processErr = s.nationalStatusService.ProcessNationalWeeklyStatuses(ctx, batchID)
	case "alerts":
		processErr = s.alertsService.ProcessAlerts(ctx, batchID)
	default:
		processErr = fmt.Errorf("unsupported dataset type: %s", batch.DatasetType)
	}

	if processErr != nil {
		_, _ = s.repo.UpdateImportBatchStatus(ctx, db.UpdateImportBatchStatusParams{
			ID:     batchID,
			Status: "FAILED",
			Notes: sql.NullString{
				String: processErr.Error(),
				Valid:  true,
			},
		})
		return processErr
	}

	_, err = s.repo.UpdateImportBatchStatus(ctx, db.UpdateImportBatchStatusParams{
		ID:     batchID,
		Status: "COMPLETED",
		Notes: sql.NullString{
			String: "Batch processed successfully",
			Valid:  true,
		},
	})
	if err != nil {
		return err
	}

	return nil
}
