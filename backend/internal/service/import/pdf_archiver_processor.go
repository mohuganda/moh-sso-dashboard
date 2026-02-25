package service

import (
	"context"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	repository "github.com/moh-sso-dashboard/internal/repository/document"
)

type PdfArchiverProcessor struct {
	repository repository.DocumentRepository
}

func NewPdfArchiverProcessor(repository repository.DocumentRepository) *PdfArchiverProcessor {
	return &PdfArchiverProcessor{repository: repository}
}

func (c *PdfArchiverProcessor) Process(ctx context.Context, p db.Process) error {
	_, err := c.repository.GetDocument(ctx, p.DocumentID)
	if err != nil {
		return err
	}

	return nil
}
