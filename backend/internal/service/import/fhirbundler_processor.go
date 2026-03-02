package service

import (
	"context"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	repository "github.com/moh-sso-dashboard/internal/repository/document"
)

type FhirBundlerProcessor struct {
	repository repository.DocumentRepository
}

func NewFhirBundlerProcessor(repository repository.DocumentRepository) *FhirBundlerProcessor {
	return &FhirBundlerProcessor{repository: repository}
}

func (c *FhirBundlerProcessor) Process(ctx context.Context, p db.Process) error {
	_, err := c.repository.GetDocument(ctx, p.DocumentID)
	if err != nil {
		return err
	}

	return nil
}
