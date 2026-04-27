package worker

import (
	"context"
	"time"

	logger "github.com/moh-sso-dashboard/internal/log"
)

func recoverWorkerPanic(log *logger.Logger, workerName string) {
	if r := recover(); r != nil {
		if log != nil {
			log.Error(
				"worker panic recovered",
				"worker", workerName,
				"panic", r,
			)
		}
	}
}

func sleepWithContext(ctx context.Context, delay time.Duration) {
	if delay <= 0 {
		return
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return
	case <-timer.C:
		return
	}
}
