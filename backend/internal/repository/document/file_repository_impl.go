package document

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

type fileRepository struct{}

func NewFileRepository() FileRepository {
	return &fileRepository{}
}

func (r *fileRepository) CreateCustomFile(
	ctx context.Context,
	tx *sql.Tx,
	fileName, filePath, templateCode, documentID string,
) (int64, error) {
	var fileKey int64
	err := tx.QueryRowContext(ctx, `
		INSERT INTO import.custom_files
			(file_name, file_path, template_code, document_id, effective_start_date)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING file_key
	`, fileName, filePath, templateCode, documentID).Scan(&fileKey)
	if err != nil {
		return 0, err
	}
	return fileKey, nil
}

func (r *fileRepository) InsertCustomData(
	ctx context.Context,
	tx *sql.Tx,
	fileKey int64,
	data []byte,
) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO import.custom_data_files
			(file_data, file_key, effective_start_date)
		VALUES ($1, $2, NOW())
	`, data, fileKey)
	return err
}

func (r *fileRepository) InsertCustomDataBatch(
	ctx context.Context,
	tx *sql.Tx,
	fileKey int64,
	data [][]byte,
) error {
	if len(data) == 0 {
		return nil
	}

	var (
		args         = make([]any, 0, len(data)*2)
		valueStrings = make([]string, 0, len(data))
	)

	for i, row := range data {
		valueStrings = append(
			valueStrings,
			fmt.Sprintf("($%d, $%d, NOW())", i*2+1, i*2+2),
		)
		args = append(args, row, fileKey)
	}

	query := fmt.Sprintf(`
		INSERT INTO import.custom_data_files
			(file_data, file_key, effective_start_date)
		VALUES %s
	`, strings.Join(valueStrings, ","))

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *fileRepository) UpsertCustomDataBatch(
	ctx context.Context,
	tx *sql.Tx,
	fileKey int64,
	templateCode string,
	rows []UpsertRow,
) error {
	if len(rows) == 0 {
		return nil
	}

	// 4 params per row: file_data, file_key, template_code, row_hash
	args := make([]any, 0, len(rows)*4)
	valueStrings := make([]string, 0, len(rows))

	for i, row := range rows {
		base := i * 4
		valueStrings = append(
			valueStrings,
			fmt.Sprintf("($%d, $%d, $%d, $%d, NOW())", base+1, base+2, base+3, base+4),
		)
		args = append(args, row.Data, fileKey, templateCode, row.Hash)
	}

	query := fmt.Sprintf(`
		INSERT INTO import.custom_data_files
			(file_data, file_key, template_code, row_hash, effective_start_date)
		VALUES %s
		ON CONFLICT (template_code, row_hash)
		DO UPDATE SET
			file_data    = EXCLUDED.file_data,
			file_key     = EXCLUDED.file_key,
			last_updated = NOW(),
			is_valid     = TRUE
	`, strings.Join(valueStrings, ","))

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *fileRepository) SoftDeleteRemovedRows(
	ctx context.Context,
	tx *sql.Tx,
	templateCode string,
	keepHashes []string,
) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE import.custom_data_files
		SET is_valid = FALSE, last_updated = NOW()
		WHERE template_code = $1
		  AND is_valid = TRUE
		  AND row_hash IS NOT NULL
		  AND NOT (row_hash = ANY($2))
	`, templateCode, pq.Array(keepHashes))
	return err
}
