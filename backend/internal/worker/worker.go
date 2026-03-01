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

func (w *Worker) Start(ctx context.Context) {

	log.Println("📦 Document worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Worker shutting down")
			return
		default:
			w.processNext(ctx)
		}
	}
}

func (w *Worker) idle() {
	time.Sleep(w.pollDelay)
}

func (w *Worker) processNext(ctx context.Context) {

	proc, err := w.processRepo.ClaimNextPending(ctx)
	if err != nil {

		// No rows available
		// 1️⃣ No pending jobs (normal situation)
		if errors.Is(err, sql.ErrNoRows) {
			w.idle()
			return
		}

		// Unexpected DB error
		log.Printf("❌ claim error: %v\n", err)
		time.Sleep(w.pollDelay)
		return
	}

	log.Printf("🚀 Processing %s (%s)\n", proc.ID, proc.ProcessType)

	err = w.importSvc.Execute(ctx, proc.ID)
	if err != nil {
		log.Printf("❌ execution failed: %v\n", err)
		// Fail already handled inside Execute
		return
	}

	log.Printf("✅ Completed %s\n", proc.ID)
}
