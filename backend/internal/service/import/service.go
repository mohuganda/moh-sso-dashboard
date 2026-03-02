package service

import (
	"context"

	"github.com/google/uuid"
	documentRepo "github.com/moh-sso-dashboard/internal/repository/document"
	repository "github.com/moh-sso-dashboard/internal/repository/document"
	processRepo "github.com/moh-sso-dashboard/internal/repository/processes"
)

type Service struct {
	repository  documentRepo.DocumentRepository
	processRepo processRepo.ProcessRepository
	registry    *Registry
}

func NewService(repository repository.DocumentRepository, processRepo processRepo.ProcessRepository) *Service {
	reg := NewRegistry()

	s := &Service{
		repository:  repository,
		processRepo: processRepo,
		registry:    reg,
	}

	// Register processors
	reg.Register("CSV_IMPORT", NewCSVProcessor(repository, processRepo))
	reg.Register("EXCEL_IMPORT", NewExcelProcessor(repository))
	reg.Register("FHIR_IMPORT", NewFhirBundlerProcessor(repository))
	reg.Register("USER_BULK_IMPORT", NewUserBulkProcessor(repository))

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
