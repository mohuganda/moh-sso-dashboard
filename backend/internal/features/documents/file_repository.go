package documents

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type CustomImportRow struct {
	FileKey      int64
	DocumentID   uuid.UUID
	TemplateCode string
	SheetCode    string
	RowNumber    int
	ReportDate   *time.Time
	Data         map[string]any
	RowHash      string
}

type FileRepository interface {
	CreateFile(
		ctx context.Context,
		tx *sql.Tx,
		documentID uuid.UUID,
		templateCode, fileName, filePath string,
		reportDate *time.Time,
	) (int64, error)

	UpsertRowsBatch(ctx context.Context, tx *sql.Tx, rows []CustomImportRow) error

	SoftDeleteBySheet(
		ctx context.Context,
		tx *sql.Tx,
		documentID uuid.UUID,
		templateCode, sheetCode string,
		keepHashes []string,
	) error

	InvalidateByDocument(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error

	InvalidateFile(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error
}
