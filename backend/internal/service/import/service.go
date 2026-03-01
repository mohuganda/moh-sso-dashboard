package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	documentRepo "github.com/moh-sso-dashboard/internal/repository/document"
	processRepo "github.com/moh-sso-dashboard/internal/repository/processes"
	"github.com/moh-sso-dashboard/internal/storage"
)

type Service struct {
	documentRepo   documentRepo.DocumentRepository
	processRepo    processRepo.ProcessRepository
	fileRepository documentRepo.FileRepository
	registry       *Registry
	storage        storage.Storage
	remoteDB       *sql.DB
}

func NewService(documentRepo documentRepo.DocumentRepository, processRepo processRepo.ProcessRepository, fileRepository documentRepo.FileRepository, storage storage.Storage, remote *sql.DB) *Service {
	reg := NewRegistry()

	s := &Service{
		documentRepo:   documentRepo,
		processRepo:    processRepo,
		fileRepository: fileRepository,
		registry:       reg,
		storage:        storage,
		remoteDB:       remote,
	}

	// Register processors
	reg.Register("CSV_IMPORT", NewCSVProcessor(documentRepo, processRepo, fileRepository, storage, remote))
	reg.Register("EXCEL_IMPORT", NewExcelProcessor(documentRepo, storage))
	reg.Register("FHIR_IMPORT", NewFhirBundlerProcessor(documentRepo, storage))
	reg.Register("USER_BULK_IMPORT", NewUserBulkProcessor(documentRepo, storage))

	return s
}

func (s *Service) Execute(ctx context.Context, processID uuid.UUID) error {

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
