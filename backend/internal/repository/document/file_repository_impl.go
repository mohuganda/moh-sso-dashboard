package document

import (
	"context"
	"database/sql"
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
