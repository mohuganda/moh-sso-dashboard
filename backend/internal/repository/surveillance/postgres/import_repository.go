package postgres

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type ImportRepository struct {
	db db.Store
}

func NewImportRepository(db db.Store) *ImportRepository {
	return &ImportRepository{
		db: db,
	}
}

func (r *ImportRepository) CreateImportBatch(ctx context.Context, arg db.CreateImportBatchParams) (db.SurveillanceImportBatch, error) {
	return r.db.CreateImportBatch(ctx, arg)
}

func (r *ImportRepository) GetImportBatchByID(ctx context.Context, id uuid.UUID) (db.SurveillanceImportBatch, error) {
	return r.db.GetImportBatchByID(ctx, id)
}

func (r *ImportRepository) ListImportBatches(ctx context.Context) ([]db.SurveillanceImportBatch, error) {
	return r.db.ListImportBatches(ctx)
}

func (r *ImportRepository) UpdateImportBatchStatus(ctx context.Context, arg db.UpdateImportBatchStatusParams) (db.SurveillanceImportBatch, error) {
	return r.db.UpdateImportBatchStatus(ctx, arg)
}

func (r *ImportRepository) CreateImportRawRow(ctx context.Context, arg db.CreateImportRawRowParams) (db.SurveillanceImportRawRow, error) {
	return r.db.CreateImportRawRow(ctx, arg)
}

func (r *ImportRepository) ListImportRawRowsByBatch(ctx context.Context, batchID uuid.UUID) ([]db.SurveillanceImportRawRow, error) {
	return r.db.ListImportRawRowsByBatch(ctx, batchID)
}

func (r *ImportRepository) MarkRawRowProcessed(ctx context.Context, rowID uuid.UUID) error {
	return r.db.MarkSurveillanceImportRawRowProcessed(ctx, rowID)
}

func (r *ImportRepository) MarkRawRowFailed(ctx context.Context, rowID uuid.UUID, message string) error {
	return r.db.MarkSurveillanceImportRawRowFailed(ctx, db.MarkSurveillanceImportRawRowFailedParams{
		ID:           rowID,
		ErrorMessage: sql.NullString{String: message, Valid: strings.TrimSpace(message) != ""},
	})
}

func (r *ImportRepository) CompleteImportBatch(
	ctx context.Context,
	batchID uuid.UUID,
	successRows, failedRows int32,
) error {
	return r.db.CompleteSurveillanceImportBatch(ctx, db.CompleteSurveillanceImportBatchParams{
		ID:          batchID,
		SuccessRows: successRows,
		FailedRows:  failedRows,
	})
}

func (r *ImportRepository) FailImportBatch(ctx context.Context, batchID uuid.UUID, notes string) error {
	return r.db.FailSurveillanceImportBatch(ctx, db.FailSurveillanceImportBatchParams{
		ID:    batchID,
		Notes: sql.NullString{String: notes, Valid: strings.TrimSpace(notes) != ""},
	})
}

func (r *ImportRepository) DeleteProcessedRawRows(ctx context.Context, batchID uuid.UUID) error {
	return r.db.DeleteProcessedSurveillanceImportRawRows(ctx, batchID)
}
