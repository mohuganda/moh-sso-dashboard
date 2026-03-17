package document

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type fileRepository struct{}

func NewFileRepository() FileRepository {
	return &fileRepository{}
}

func (r *fileRepository) CreateCustomFile(
	ctx context.Context,
	tx *sql.Tx,
	fileName, filePath string,
) (int64, error) {
	var fileKey int64

	err := tx.QueryRowContext(ctx, `
		INSERT INTO import.custom_files 
			(file_name, file_path, effective_start_date)
		VALUES ($1, $2, NOW())
		RETURNING file_key
	`, fileName, filePath).Scan(&fileKey)
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
		// ($1, $2, NOW()), ($3, $4, NOW()), ...
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
