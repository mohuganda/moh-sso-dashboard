package service

import (
	"context"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	repository "github.com/moh-sso-dashboard/internal/repository/document"
)

type FacilityProcessor struct {
	repository repository.DocumentRepository
}

func NewFacilityProcessor(repository repository.DocumentRepository) *FacilityProcessor {
	return &FacilityProcessor{repository: repository}
}

func (c *FacilityProcessor) Process(ctx context.Context, p db.Process) error {
	_, err := c.repository.GetDocument(ctx, p.DocumentID)
	if err != nil {
		return err
	}

	return nil
}
