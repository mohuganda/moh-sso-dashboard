package data_quality

import (
	"context"
	"database/sql"
	"time"
)

// ScheduleWorker executes scans that were queued to run later.
type ScheduleWorker struct {
	store    *DQAStore
	dwhDB    *sql.DB
	interval time.Duration
	log      func(args ...any)
}

func NewScheduleWorker(store *DQAStore, dwhDB *sql.DB, interval time.Duration, log func(args ...any)) *ScheduleWorker {
	if interval <= 0 {
		interval = time.Minute
	}
	return &ScheduleWorker{store: store, dwhDB: dwhDB, interval: interval, log: log}
}

func (w *ScheduleWorker) Start(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			w.drain(ctx)
		}
	}
}

// drain runs every due scan, one at a time, so a backlog cannot start a dozen
// warehouse scans at once.
func (w *ScheduleWorker) drain(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		scheduled, found, err := w.store.claimDueRun(ctx, time.Now())
		if err != nil {
			w.logf("dqa: claiming a scheduled run failed: ", err)
			return
		}
		if !found {
			return
		}

		result, runErr := w.store.RunTable(ctx, w.dwhDB, scheduled.TableID, scheduled.Scope, scheduled.CreatedBy)
		var runID int64
		if runErr == nil {
			runID = result.RunID
		} else {
			w.logf("dqa: scheduled run failed: ", runErr)
		}
		if err := w.store.finishScheduledRun(ctx, scheduled.ID, runID, runErr); err != nil {
			w.logf("dqa: recording a scheduled run result failed: ", err)
		}
	}
}

func (w *ScheduleWorker) logf(args ...any) {
	if w.log != nil {
		w.log(args...)
	}
}
