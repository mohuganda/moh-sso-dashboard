package bootstrap

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

// The pool remains shared with repositories throughout outages. Readiness is
// published only after a successful connection and completed migrations.
type observatoryMonitor struct {
	ready  atomic.Bool
	cancel context.CancelFunc
	done   chan struct{}
}

func (m *observatoryMonitor) Ready() bool { return m.ready.Load() }
func (m *observatoryMonitor) Close()      { m.cancel(); <-m.done }

func startObservatoryMonitor(parent context.Context, interval time.Duration,
	ping func(context.Context) error, initialize func(context.Context) error) *observatoryMonitor {
	ctx, cancel := context.WithCancel(parent)
	monitor := &observatoryMonitor{cancel: cancel, done: make(chan struct{})}
	go func() {
		defer close(monitor.done)
		defer monitor.ready.Store(false)
		initialized := false
		first := true
		for {
			if ctx.Err() != nil {
				return
			}
			checkCtx, stopCheck := context.WithTimeout(ctx, 3*time.Second)
			err := ping(checkCtx)
			stopCheck()
			phase := "connection"
			if err == nil && !initialized {
				phase = "migrations"
				initCtx, stopInit := context.WithTimeout(ctx, 30*time.Second)
				err = initialize(initCtx)
				stopInit()
				initialized = err == nil
			}
			if err != nil {
				initialized = false
			}
			available := err == nil && ctx.Err() == nil
			changed := monitor.ready.Swap(available) != available
			if first || changed {
				if available {
					log.Info().Msg("Observatory is available; surveillance requests are enabled")
				} else {
					// Database errors may contain connection metadata. Report only the phase.
					log.Warn().Str("phase", phase).Msg("Observatory unavailable; portal remains available and connection will retry")
				}
			}
			first = false
			timer := time.NewTimer(interval)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
	}()
	return monitor
}
