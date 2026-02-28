package service

import (
	"context"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	repository "github.com/moh-sso-dashboard/internal/repository/document"
	"github.com/moh-sso-dashboard/internal/storage"
)

type UserBulkProcessor struct {
	repository repository.DocumentRepository
	storage    storage.Storage
	db         db.Store
}

func NewUserBulkProcessor(repository repository.DocumentRepository, storage storage.Storage, db db.Store) *UserBulkProcessor {
	return &UserBulkProcessor{repository: repository, storage: storage, db: db}
}

func (c *UserBulkProcessor) Process(ctx context.Context, p db.Process) error {
	_, err := c.repository.GetDocument(ctx, p.DocumentID)
	if err != nil {
		return err
	}

	return nil
}
