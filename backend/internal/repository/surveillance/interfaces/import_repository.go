package interfaces

import (
	"context"
	"io"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type ImportRepository interface {
	CreateImportBatch(ctx context.Context, arg db.CreateImportBatchParams) (db.SurveillanceImportBatch, error)
	GetImportBatchByID(ctx context.Context, id uuid.UUID) (db.SurveillanceImportBatch, error)
	ListImportBatches(ctx context.Context) ([]db.SurveillanceImportBatch, error)
	UpdateImportBatchStatus(ctx context.Context, arg db.UpdateImportBatchStatusParams) (db.SurveillanceImportBatch, error)

	CreateImportRawRow(ctx context.Context, arg db.CreateImportRawRowParams) (db.SurveillanceImportRawRow, error)
	ListImportRawRowsByBatch(ctx context.Context, batchID uuid.UUID) ([]db.SurveillanceImportRawRow, error)
	ImportCSV(ctx context.Context, reader io.Reader, fileName string, sourceName string) error
}
