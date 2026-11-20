package core

import (
	"errors"
	"testing"
	"time"
)

func TestWorkerStatusRecoversAndStops(t *testing.T) {
	var status WorkerStatus
	at := time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)
	if got := status.Snapshot(); got.State != "starting" || got.LastSuccess != nil || got.LastFailure != nil {
		t.Fatalf("unexpected initial status: %+v", got)
	}
	status.RecordPoll(at, errors.New("queue unavailable"))
	status.RecordPoll(at, nil)
	got := status.Snapshot()
	if got.State != "ready" || got.LastSuccess == nil || got.LastFailure == nil ||
		!got.LastSuccess.Equal(at) || !got.LastFailure.Equal(at) {
		t.Fatalf("unexpected recovered status: %+v", got)
	}
	status.Stop()
	status.RecordPoll(at.Add(time.Minute), nil)
	got = status.Snapshot()
	if got.State != "stopped" || got.LastSuccess == nil || !got.LastSuccess.Equal(at) {
		t.Fatalf("unexpected stopped status: %+v", got)
	}
}
