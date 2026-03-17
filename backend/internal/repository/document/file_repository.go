package document

import (
	"context"
	"database/sql"
)

type FileRepository interface {
	CreateCustomFile(
		ctx context.Context,
		tx *sql.Tx,
		fileName, filePath string,
	) (int64, error)

	InsertCustomData(
		ctx context.Context,
		tx *sql.Tx,
		fileKey int64,
		data []byte,
	) error

	InsertCustomDataBatch(
		ctx context.Context,
		tx *sql.Tx,
		fileKey int64,
		data [][]byte,
	) error
}
