package service

import (
	"context"
	"io"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
)

type SurveillanceImportResult struct {
	FileName     string `json:"fileName"`
	SourceName   string `json:"sourceName,omitempty"`
	RowsRead     int    `json:"rowsRead"`
	RowsImported int    `json:"rowsImported"`
	RowsSkipped  int    `json:"rowsSkipped"`
	Message      string `json:"message"`
}

type SurveillanceImportService struct {
	importRepo interfaces.ImportRepository
}

func NewSurveillanceImportService(importRepo interfaces.ImportRepository) *SurveillanceImportService {
	return &SurveillanceImportService{
		importRepo: importRepo,
	}
}

func (s *SurveillanceImportService) ListImportBatches(ctx context.Context) ([]db.SurveillanceImportBatch, error) {
	return s.importRepo.ListImportBatches(ctx)
}

func (s *SurveillanceImportService) GetImportBatchByID(ctx context.Context, batchID uuid.UUID) (db.SurveillanceImportBatch, error) {
	return s.importRepo.GetImportBatchByID(ctx, batchID)
}

func (s *SurveillanceImportService) CreateImportBatch(ctx context.Context, arg db.CreateImportBatchParams) (db.SurveillanceImportBatch, error) {
	return s.importRepo.CreateImportBatch(ctx, arg)
}

func (s *SurveillanceImportService) UpdateImportBatchStatus(ctx context.Context, arg db.UpdateImportBatchStatusParams) (db.SurveillanceImportBatch, error) {
	return s.importRepo.UpdateImportBatchStatus(ctx, arg)
}

func (s *SurveillanceImportService) ListImportRawRowsByBatch(ctx context.Context, batchID uuid.UUID) ([]db.SurveillanceImportRawRow, error) {
	return s.importRepo.ListImportRawRowsByBatch(ctx, batchID)
}

func (s *SurveillanceImportService) ImportSurveillanceCSV(
	ctx context.Context,
	reader io.Reader,
	fileName string,
	sourceName string,
) error {
	return s.importRepo.ImportCSV(ctx, reader, fileName, sourceName)
}
