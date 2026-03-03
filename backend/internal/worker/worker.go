package worker

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	processRepository "github.com/moh-sso-dashboard/internal/repository/processes"
	importService "github.com/moh-sso-dashboard/internal/service/import"
	"github.com/moh-sso-dashboard/internal/storage"
)

type Worker struct {
	processRepo processRepository.ProcessRepository
	importSvc   *importService.Service
	pollDelay   time.Duration
	storage     storage.Storage
}

func NewWorker(
	processRepo processRepository.ProcessRepository,
	importSvc *importService.Service,
	pollDelay time.Duration,
	storage storage.Storage,
) *Worker {
	return &Worker{
		processRepo: processRepo,
		importSvc:   importSvc,
		pollDelay:   pollDelay,
		storage:     storage,
	}
}

func (w *Worker) Start(ctx context.Context) error {

	log.Println("📦 Document worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Worker shutting down")
			return nil
		default:
			if err := w.processNext(ctx); err != nil {
				return err
			}
			w.idle()
		}
	}
}

func (w *Worker) idle() {
	time.Sleep(w.pollDelay)
}

func (w *Worker) processNext(ctx context.Context) error {

	defer func() {
		if r := recover(); r != nil {
			log.Printf("🔥 worker panic recovered: %v", r)
		}
	}()

	// Stop immediately if shutting down
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	proc, err := w.processRepo.ClaimNextPending(ctx)
	if err != nil {

		// No pending jobs (normal idle state)
		if errors.Is(err, sql.ErrNoRows) {
			w.idle()
			return nil
		}

		// Unexpected DB error
		log.Printf("❌ claim error: %v", err)
		time.Sleep(w.pollDelay)
		return err
	}

	log.Printf("🚀 Processing %s (%s)", proc.ID, proc.ProcessType)

	if err := w.importSvc.Execute(ctx, proc.ID); err != nil {
		log.Printf("❌ execution failed for %s: %v", proc.ID, err)
		// Execution already marks failure internally
		return err
	}

	log.Printf("✅ Completed %s", proc.ID)
	return nil
}
