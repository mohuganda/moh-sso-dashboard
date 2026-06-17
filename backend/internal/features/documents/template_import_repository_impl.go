package documents

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type templateImportRepository struct{}

func NewTemplateImportRepository() TemplateImportRepository {
	return &templateImportRepository{}
}

func (r *templateImportRepository) CreateUpload(
	ctx context.Context,
	tx *sql.Tx,
	documentID uuid.UUID,
	templateCode string,
	reportDate *time.Time,
) (uuid.UUID, error) {
	var id uuid.UUID
	err := tx.QueryRowContext(ctx, `
		INSERT INTO import.template_uploads (document_id, template_code, report_date)
		VALUES ($1, $2, $3)
		ON CONFLICT (document_id) DO UPDATE
		  SET template_code = EXCLUDED.template_code,
		      report_date   = EXCLUDED.report_date,
		      is_valid      = TRUE,
		      last_updated  = NOW()
		RETURNING id
	`, documentID, templateCode, reportDate).Scan(&id)
	return id, err
}

func (r *templateImportRepository) UpsertRowsBatch(
	ctx context.Context,
	tx *sql.Tx,
	rows []TemplateImportRow,
) error {
	if len(rows) == 0 {
		return nil
	}

	const colsPerRow = 9 // upload_id, document_id, template_code, sheet_code, row_number, report_date, data, raw_data, row_hash

	args := make([]any, 0, len(rows)*colsPerRow)
	placeholderSets := make([]string, 0, len(rows))

	for i, row := range rows {
		dataJSON, err := json.Marshal(row.Data)
		if err != nil {
			return fmt.Errorf("marshal data row %d: %w", row.RowNumber, err)
		}
		rawJSON, err := json.Marshal(row.RawData)
		if err != nil {
			return fmt.Errorf("marshal raw_data row %d: %w", row.RowNumber, err)
		}

		base := i * colsPerRow
		set := make([]string, colsPerRow)
		for j := range set {
			set[j] = fmt.Sprintf("$%d", base+j+1)
		}
		placeholderSets = append(placeholderSets, "("+strings.Join(set, ",")+")")

		args = append(args,
			row.UploadID,
			row.DocumentID,
			row.TemplateCode,
			row.SheetCode,
			row.RowNumber,
			row.ReportDate,
			dataJSON,
			rawJSON,
			row.RowHash,
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO import.template_row_data
			(upload_id, document_id, template_code, sheet_code, row_number, report_date, data, raw_data, row_hash)
		VALUES %s
		ON CONFLICT (row_hash) DO UPDATE SET
			upload_id     = EXCLUDED.upload_id,
			document_id   = EXCLUDED.document_id,
			template_code = EXCLUDED.template_code,
			sheet_code    = EXCLUDED.sheet_code,
			row_number    = EXCLUDED.row_number,
			report_date   = EXCLUDED.report_date,
			data          = EXCLUDED.data,
			raw_data      = EXCLUDED.raw_data,
			is_valid      = TRUE,
			last_updated  = NOW()
	`, strings.Join(placeholderSets, ","))

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *templateImportRepository) SoftDeleteBySheet(
	ctx context.Context,
	tx *sql.Tx,
	templateCode, sheetCode string,
	keepHashes []string,
) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE import.template_row_data
		SET is_valid = FALSE, last_updated = NOW()
		WHERE template_code = $1
		  AND sheet_code    = $2
		  AND is_valid      = TRUE
		  AND row_hash IS NOT NULL
		  AND NOT (row_hash = ANY($3))
	`, templateCode, sheetCode, pq.Array(keepHashes))
	return err
}

func (r *templateImportRepository) InvalidateByDocument(
	ctx context.Context,
	tx *sql.Tx,
	documentID uuid.UUID,
) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE import.template_row_data
		SET is_valid = FALSE, last_updated = NOW()
		WHERE document_id = $1 AND is_valid = TRUE
	`, documentID)
	return err
}

func (r *templateImportRepository) InvalidateUpload(
	ctx context.Context,
	tx *sql.Tx,
	documentID uuid.UUID,
) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE import.template_uploads
		SET is_valid = FALSE, last_updated = NOW()
		WHERE document_id = $1 AND is_valid = TRUE
	`, documentID)
	return err
}
