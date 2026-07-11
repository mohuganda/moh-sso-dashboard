package worker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	logger "github.com/moh-sso-dashboard/internal/log"
	processRepository "github.com/moh-sso-dashboard/internal/repository/processes"
	importService "github.com/moh-sso-dashboard/internal/service/import"
	"github.com/moh-sso-dashboard/internal/storage"
)

type DocumentWorker struct {
	processRepo processRepository.ProcessRepository
	importSvc   *importService.Service
	pollDelay   time.Duration
	storage     storage.Storage
	logger      *logger.Logger
}

func NewDocumentWorker(
	processRepo processRepository.ProcessRepository,
	importSvc *importService.Service,
	pollDelay time.Duration,
	storage storage.Storage,
	logger *logger.Logger,
) (*DocumentWorker, error) {
	if processRepo == nil {
		return nil, errors.New("process repository is required")
	}

	if importSvc == nil {
		return nil, errors.New("import service is required")
	}

	if storage == nil {
		return nil, errors.New("storage service is required")
	}

	if logger == nil {
		return nil, errors.New("logger is required")
	}

	if pollDelay <= 0 {
		pollDelay = 5 * time.Second
	}

	return &DocumentWorker{
		processRepo: processRepo,
		importSvc:   importSvc,
		pollDelay:   pollDelay,
		storage:     storage,
		logger:      logger,
	}, nil
}

func (w *DocumentWorker) Start(ctx context.Context) error {
	if w == nil {
		return errors.New("document worker is nil")
	}

	w.logger.Info("document worker started")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("document worker shutting down")
			return nil

		default:
			if err := w.processNext(ctx); err != nil {
				w.logger.Error("document worker processing failed", "error", err)
			}

			w.idle(ctx)
		}
	}
}

func (w *DocumentWorker) processNext(ctx context.Context) error {
	defer recoverWorkerPanic(w.logger, "document worker")

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	proc, err := w.processRepo.ClaimNextPending(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return fmt.Errorf("claim next pending process: %w", err)
	}

	w.logger.Info(
		"processing document job",
		"process_id", proc.ID,
		"process_type", proc.ProcessType,
	)

	if err := w.importSvc.Execute(ctx, proc.ID); err != nil {
		return fmt.Errorf("execute document job %s: %w", proc.ID, err)
	}

	w.logger.Info(
		"document job completed",
		"process_id", proc.ID,
	)

	return nil
}

func (w *DocumentWorker) idle(ctx context.Context) {
	sleepWithContext(ctx, w.pollDelay)
}
