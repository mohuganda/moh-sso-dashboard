package documents

import (
	"context"
	"database/sql"
)

type UpsertRow struct {
	Data []byte
	Hash string
}

type FileRepository interface {
	CreateCustomFile(
		ctx context.Context,
		tx *sql.Tx,
		fileName, filePath, templateCode, documentID string,
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

	UpsertCustomDataBatch(
		ctx context.Context,
		tx *sql.Tx,
		fileKey int64,
		templateCode string,
		rows []UpsertRow,
	) error

	SoftDeleteRemovedRows(
		ctx context.Context,
		tx *sql.Tx,
		templateCode string,
		keepHashes []string,
	) error
}
