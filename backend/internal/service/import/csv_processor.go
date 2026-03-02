package service

import (
	"context"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	documentRepository "github.com/moh-sso-dashboard/internal/repository/document"
	processRepository "github.com/moh-sso-dashboard/internal/repository/processes"
)

type CSVProcessor struct {
	documentRepository documentRepository.DocumentRepository
	processRepository  processRepository.ProcessRepository
}

func NewCSVProcessor(documentRepository documentRepository.DocumentRepository, processRepository processRepository.ProcessRepository) *CSVProcessor {
	return &CSVProcessor{documentRepository: documentRepository, processRepository: processRepository}
}

func (c *CSVProcessor) Process(ctx context.Context, p db.Process) error {
	_, err := c.documentRepository.GetDocument(ctx, p.DocumentID)
	if err != nil {
		return err
	}

	// Update progress
	msg := "Parsing CSV"
	_ = c.processRepository.UpdateProgress(ctx, p.ID, 10, &msg)

	// TODO:
	// - Open file from storage
	// - Parse CSV
	// - Transform rows
	// - Insert into DB
	// - Update progress incrementally

	msg = "Finished processing"
	_ = c.processRepository.UpdateProgress(ctx, p.ID, 90, &msg)

	return nil
}
