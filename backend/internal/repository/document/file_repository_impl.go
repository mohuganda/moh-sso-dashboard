package document

import (
	"context"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type fileRepository struct{}

func NewFileRepository() FileRepository {
	return &fileRepository{}
}

func (r *fileRepository) CreateCustomFile(
	ctx context.Context,
	q db.DBTX,
	fileName, filePath string,
) (int64, error) {

	var fileKey int64

	err := q.QueryRowContext(ctx, `
		INSERT INTO custom_files 
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
	q db.DBTX,
	fileKey int64,
	data []byte,
) error {

	_, err := q.ExecContext(ctx, `
		INSERT INTO custom_data_files
			(file_data, file_key, effective_start_date)
		VALUES ($1, $2, NOW())
	`, data, fileKey)

	return err
}
