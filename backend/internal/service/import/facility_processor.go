package service

import (
	"context"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	repository "github.com/moh-sso-dashboard/internal/repository/document"
	"github.com/moh-sso-dashboard/internal/storage"
)

type FacilityProcessor struct {
	repository repository.DocumentRepository
	storage    storage.Storage
	db         db.Store
}

func NewFacilityProcessor(repository repository.DocumentRepository, storage storage.Storage, db db.Store) *FacilityProcessor {
	return &FacilityProcessor{repository: repository, storage: storage, db: db}
}

func (c *FacilityProcessor) Process(ctx context.Context, p db.Process, storage storage.Storage, db db.Store) error {
	_, err := c.repository.GetDocument(ctx, p.DocumentID)
	if err != nil {
		return err
	}

	return nil
}
