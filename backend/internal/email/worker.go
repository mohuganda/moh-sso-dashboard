package email

import (
	"context"
	"log"
	"time"

	repository "github.com/moh-sso-dashboard/internal/repository/email"
)

type Worker struct {
	repo       repository.EmailRepository
	sender     smtpSender
	pollEvery  time.Duration
	batchSize  int32
	maxRetries int32
}

func NewWorker(repo repository.EmailRepository, sender smtpSender, pollEvery time.Duration, batchSize int32, maxRetries int32) *Worker {
	return &Worker{
		repo:       repo,
		sender:     sender,
		pollEvery:  pollEvery,
		batchSize:  batchSize,
		maxRetries: maxRetries,
	}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.pollEvery)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.process(ctx)
		}
	}
}

func (w *Worker) process(ctx context.Context) {
	items, err := w.repo.ClaimNextBatch(ctx, w.batchSize)
	if err != nil {
		log.Printf("email worker: claim batch failed: %v", err)
		return
	}

	for _, item := range items {
		err := w.sender.Send(ctx, item.Message)
		if err == nil {
			if markErr := w.repo.MarkSent(ctx, item.ID); markErr != nil {
				log.Printf("email worker: mark sent failed: %v", markErr)
			}
			continue
		}

		attempts := item.Attempts + 1
		if attempts >= item.MaxAttempts || attempts >= w.maxRetries {
			if markErr := w.repo.MarkFailed(ctx, item.ID, attempts, err.Error()); markErr != nil {
				log.Printf("email worker: mark failed failed: %v", markErr)
			}
			continue
		}

		if markErr := w.repo.MarkRetry(ctx, item.ID, attempts, err.Error()); markErr != nil {
			log.Printf("email worker: mark retry failed: %v", markErr)
		}
	}
}
