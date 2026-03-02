package service

import (
	"context"
	"database/sql"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/model"
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
	logger         *logger.Logger
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
	reg.Register(model.ProcessTypeCSVImport, NewCSVProcessor(documentRepo, processRepo, fileRepository, storage, remote))

	reg.Register(model.ProcessTypeExcelImport, NewExcelProcessor(documentRepo, storage))

	reg.Register(model.ProcessTypeFHIRImport, NewFhirBundlerProcessor(documentRepo, storage))

	reg.Register(model.ProcessTypeUserBulkImport, NewUserBulkProcessor(documentRepo, storage))

	return s
}

func (s *Service) Execute(ctx context.Context, processID uuid.UUID) error {

	start := time.Now()

	s.logger.Info("process execution started | process_id=", processID)

	// Panic protection (important for background jobs)
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error(
				"process panicked | process_id=", processID,
				" | panic=", r,
				" | stack=", string(debug.Stack()),
			)
			_ = s.processRepo.Fail(ctx, processID, "internal panic")
		}
	}()

	// 1️⃣ Load process
	proc, err := s.processRepo.GetProcessByID(ctx, processID)
	if err != nil {
		s.logger.Error(
			"failed to load process | process_id=", processID,
			" | error=", err,
		)
		return err
	}

	s.logger.Info(
		"process loaded | process_id=", proc.ID,
		" | document_id=", proc.DocumentID,
		" | type=", proc.ProcessType,
		" | attempts=", proc.Attempts,
	)

	// 2️⃣ Resolve processor
	processor, ok := s.registry.Get(model.ProcessType(proc.ProcessType))
	if !ok {
		s.logger.Error(
			"unsupported process type | process_id=", proc.ID,
			" | type=", proc.ProcessType,
		)

		if err := s.processRepo.Fail(ctx, proc.ID, "unsupported process type"); err != nil {
			s.logger.Error(
				"failed to mark process as FAILED | process_id=", proc.ID,
				" | error=", err,
			)
		}

		return fmt.Errorf("unsupported process type: %s", proc.ProcessType)
	}

	// 3️⃣ Execute processor
	s.logger.Info(
		"executing processor | process_id=", proc.ID,
		" | type=", proc.ProcessType,
	)

	err = processor.Process(ctx, proc)
	if err != nil {

		s.logger.Error(
			"processor execution failed | process_id=", proc.ID,
			" | document_id=", proc.DocumentID,
			" | type=", proc.ProcessType,
			" | error=", err,
		)

		if failErr := s.processRepo.Fail(ctx, processID, err.Error()); failErr != nil {
			s.logger.Error(
				"failed to update FAILED status | process_id=", processID,
				" | original_error=", err,
				" | update_error=", failErr,
			)
		}

		return err
	}

	// 4️⃣ Mark complete
	if err := s.processRepo.Complete(ctx, processID); err != nil {
		s.logger.Error(
			"failed to mark process as COMPLETE | process_id=", processID,
			" | error=", err,
		)
		return err
	}

	s.logger.Info(
		"process completed successfully | process_id=", processID,
		" | duration_ms=", time.Since(start).Milliseconds(),
	)

	return nil
}
