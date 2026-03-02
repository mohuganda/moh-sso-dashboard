package service

import (
	"context"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	repository "github.com/moh-sso-dashboard/internal/repository/document"
)

type UserBulkProcessor struct {
	repository repository.DocumentRepository
}

func NewUserBulkProcessor(repository repository.DocumentRepository) *UserBulkProcessor {
	return &UserBulkProcessor{repository: repository}
}

func (c *UserBulkProcessor) Process(ctx context.Context, p db.Process) error {
	_, err := c.repository.GetDocument(ctx, p.DocumentID)
	if err != nil {
		return err
	}

	return nil
}
