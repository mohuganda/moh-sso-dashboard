package service

import (
	"context"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	repository "github.com/moh-sso-dashboard/internal/features/documents"
	"github.com/moh-sso-dashboard/internal/storage"
)

type UserBulkProcessor struct {
	repository repository.DocumentRepository
	storage    storage.Storage
}

func NewUserBulkProcessor(repository repository.DocumentRepository, storage storage.Storage) *UserBulkProcessor {
	return &UserBulkProcessor{repository: repository, storage: storage}
}

func (c *UserBulkProcessor) Process(ctx context.Context, p db.Process) error {
	_, err := c.repository.GetDocument(ctx, p.DocumentID)
	if err != nil {
		return err
	}

	return nil
}
