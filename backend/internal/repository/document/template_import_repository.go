package document

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// TemplateImportRow is a single parsed data row ready to be persisted.
type TemplateImportRow struct {
	UploadID     uuid.UUID
	DocumentID   uuid.UUID
	TemplateCode string
	SheetCode    string
	RowNumber    int
	ReportDate   *time.Time
	Data         map[string]any
	RawData      map[string]any
	RowHash      string
}

// TemplateImportRepository persists data for any template upload using the
// generic import.template_uploads + import.template_row_data tables.
type TemplateImportRepository interface {
	// CreateUpload inserts a row into import.template_uploads and returns its ID.
	CreateUpload(
		ctx context.Context,
		tx *sql.Tx,
		documentID uuid.UUID,
		templateCode string,
		reportDate *time.Time,
	) (uuid.UUID, error)

	// UpsertRowsBatch inserts or updates a batch of data rows.
	UpsertRowsBatch(ctx context.Context, tx *sql.Tx, rows []TemplateImportRow) error

	// SoftDeleteBySheet marks rows is_valid=false for the given template+sheet
	// whose row_hash is NOT in keepHashes (rows removed from the latest upload).
	SoftDeleteBySheet(
		ctx context.Context,
		tx *sql.Tx,
		templateCode, sheetCode string,
		keepHashes []string,
	) error

	// InvalidateByDocument marks all rows for a given document as is_valid=false,
	// used when a document is replaced by a newer upload.
	InvalidateByDocument(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error

	// InvalidateUpload marks the template_uploads row is_valid=false.
	InvalidateUpload(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error
}
