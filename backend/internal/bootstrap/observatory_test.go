package bootstrap

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func waitForObservatory(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.After(3 * time.Second)
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for !condition() {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for observatory state")
		case <-ticker.C:
		}
	}
}

func TestObservatoryRetriesWithoutBlockingAndRecoversAfterOutage(t *testing.T) {
	var reachable atomic.Bool
	var initializationCalls atomic.Int32
	var pingCalls atomic.Int32
	monitor := startObservatoryMonitor(context.Background(), time.Millisecond,
		func(context.Context) error {
			pingCalls.Add(1)
			if !reachable.Load() {
				return errors.New("private database address")
			}
			return nil
		},
		func(context.Context) error { initializationCalls.Add(1); return nil })
	defer monitor.Close()
	waitForObservatory(t, func() bool { return pingCalls.Load() > 1 })
	if monitor.Ready() || initializationCalls.Load() != 0 {
		t.Fatal("offline Observatory must not initialize or block the caller")
	}
	reachable.Store(true)
	waitForObservatory(t, monitor.Ready)
	if initializationCalls.Load() != 1 {
		t.Fatal("requests enabled before initialization")
	}
	// A reconnect validates migrations again before requests can resume.
	reachable.Store(false)
	waitForObservatory(t, func() bool { return !monitor.Ready() })
	reachable.Store(true)
	waitForObservatory(t, monitor.Ready)
	if initializationCalls.Load() != 2 {
		t.Fatal("recovery did not verify migrations")
	}
}

func TestObservatoryMigrationFailureStaysUnavailableUntilRetrySucceeds(t *testing.T) {
	var allowMigration atomic.Bool
	var calls atomic.Int32
	monitor := startObservatoryMonitor(context.Background(), time.Millisecond,
		func(context.Context) error { return nil },
		func(context.Context) error {
			calls.Add(1)
			if !allowMigration.Load() {
				return errors.New("migration failed")
			}
			return nil
		})
	waitForObservatory(t, func() bool { return calls.Load() > 1 })
	if monitor.Ready() {
		t.Fatal("failed migrations must not enable requests")
	}
	allowMigration.Store(true)
	waitForObservatory(t, monitor.Ready)
	monitor.Close()
	if monitor.Ready() {
		t.Fatal("shutdown retained ready state")
	}
	callsAtShutdown := calls.Load()
	time.Sleep(5 * time.Millisecond)
	if calls.Load() != callsAtShutdown {
		t.Fatal("retry continued after shutdown")
	}
}

func TestObservatoryPingIsCanceledAtShutdown(t *testing.T) {
	started := make(chan struct{})
	monitor := startObservatoryMonitor(context.Background(), time.Hour,
		func(ctx context.Context) error { close(started); <-ctx.Done(); return ctx.Err() },
		func(context.Context) error { t.Error("must not migrate after canceled ping"); return nil })
	<-started
	done := make(chan struct{})
	go func() { monitor.Close(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("shutdown failed to cancel ping")
	}
}
