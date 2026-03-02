package service

import (
	"context"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	repository "github.com/moh-sso-dashboard/internal/repository/document"
)

type ExcelProcessor struct {
	repository repository.DocumentRepository
}

func NewExcelProcessor(repository repository.DocumentRepository) *ExcelProcessor {
	return &ExcelProcessor{repository: repository}
}

func (c *ExcelProcessor) Process(ctx context.Context, p db.Process) error {
	_, err := c.repository.GetDocument(ctx, p.DocumentID)
	if err != nil {
		return err
	}

	return nil
}
