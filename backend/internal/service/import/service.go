package service

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	documentRepo "github.com/moh-sso-dashboard/internal/repository/document"
	processRepo "github.com/moh-sso-dashboard/internal/repository/processes"
	"github.com/moh-sso-dashboard/internal/storage"
)

type Service struct {
	documentRepo documentRepo.DocumentRepository
	processRepo  processRepo.ProcessRepository
	registry     *Registry
	storage      storage.Storage
	db           db.Store
}

func NewService(documentRepo documentRepo.DocumentRepository, processRepo processRepo.ProcessRepository, storage storage.Storage, db db.Store) *Service {
	reg := NewRegistry()

	s := &Service{
		documentRepo: documentRepo,
		processRepo:  processRepo,
		registry:     reg,
		storage:      storage,
		db:           db,
	}

	// Register processors
	reg.Register("CSV_IMPORT", NewCSVProcessor(documentRepo, processRepo, storage, db))
	reg.Register("EXCEL_IMPORT", NewExcelProcessor(documentRepo, storage, db))
	reg.Register("FHIR_IMPORT", NewFhirBundlerProcessor(documentRepo, storage, db))
	reg.Register("USER_BULK_IMPORT", NewUserBulkProcessor(documentRepo, storage, db))

	return s
}

func (s *Service) Execute(ctx context.Context, processID uuid.UUID, storage storage.Storage, db db.Store) error {

	// 1️⃣ Load process
	proc, err := s.processRepo.GetProcessByID(ctx, processID)
	if err != nil {
		return err
	}

	// 2️⃣ Resolve processor
	processor, ok := s.registry.Get(proc.ProcessType)
	if !ok {
		return s.processRepo.Fail(ctx, proc.ID, "unsupported process type")
	}

	// 3️⃣ Execute
	err = processor.Process(ctx, proc)
	if err != nil {
		_ = s.processRepo.Fail(ctx, processID, err.Error())
		return err
	}

	// 4️⃣ Mark complete
	return s.processRepo.Complete(ctx, processID)
}
