package service

import (
	"context"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	repository "github.com/moh-sso-dashboard/internal/repository/document"
	"github.com/moh-sso-dashboard/internal/storage"
)

type FhirBundlerProcessor struct {
	repository repository.DocumentRepository
	storage    storage.Storage
	db         db.Store
}

func NewFhirBundlerProcessor(repository repository.DocumentRepository, storage storage.Storage, db db.Store) *FhirBundlerProcessor {
	return &FhirBundlerProcessor{repository: repository, storage: storage, db: db}
}

func (c *FhirBundlerProcessor) Process(ctx context.Context, p db.Process) error {
	_, err := c.repository.GetDocument(ctx, p.DocumentID)
	if err != nil {
		return err
	}

	return nil
}
