package document

import (
	"context"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type FileRepository interface {
	CreateCustomFile(
		ctx context.Context,
		q db.DBTX,
		fileName, filePath string,
	) (int64, error)

	InsertCustomData(
		ctx context.Context,
		q db.DBTX,
		fileKey int64,
		data []byte,
	) error
}
