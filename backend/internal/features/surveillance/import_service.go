package surveillance

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type ImportService struct {
	importRepo ImportRepository
}

func NewImportService(importRepo ImportRepository) *ImportService {
	return &ImportService{
		importRepo: importRepo,
	}
}

func (s *ImportService) ListImportBatches(ctx context.Context) ([]db.SurveillanceImportBatch, error) {
	return s.importRepo.ListImportBatches(ctx)
}

func (s *ImportService) GetImportBatchByID(ctx context.Context, batchID uuid.UUID) (db.SurveillanceImportBatch, error) {
	return s.importRepo.GetImportBatchByID(ctx, batchID)
}

func (s *ImportService) CreateImportBatch(ctx context.Context, arg db.CreateImportBatchParams) (db.SurveillanceImportBatch, error) {
	return s.importRepo.CreateImportBatch(ctx, arg)
}

func (s *ImportService) UpdateImportBatchStatus(ctx context.Context, arg db.UpdateImportBatchStatusParams) (db.SurveillanceImportBatch, error) {
	return s.importRepo.UpdateImportBatchStatus(ctx, arg)
}

func (s *ImportService) ListImportRawRowsByBatch(ctx context.Context, batchID uuid.UUID) ([]db.SurveillanceImportRawRow, error) {
	return s.importRepo.ListImportRawRowsByBatch(ctx, batchID)
}
