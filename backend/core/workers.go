package core

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type WorkerTask struct {
	Name string
	Run  func(context.Context) error
}

// StartWorkerGroup cancels and joins every started task before infrastructure
// cleanup. The stop callback is safe to retry after a shutdown deadline.
func StartWorkerGroup(ctx context.Context, tasks []WorkerTask, logError func(...interface{})) Stop {
	ctx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var failures []error
	for _, task := range tasks {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := task.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
				mu.Lock()
				failures = append(failures, fmt.Errorf("%s: %w", task.Name, err))
				mu.Unlock()
				if logError != nil {
					logError("Background worker stopped: ", task.Name, err)
				}
			}
		}()
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	return func(shutdown context.Context) error {
		cancel()
		select {
		case <-done:
			mu.Lock()
			defer mu.Unlock()
			return errors.Join(failures...)
		case <-shutdown.Done():
			return shutdown.Err()
		}
	}
}
