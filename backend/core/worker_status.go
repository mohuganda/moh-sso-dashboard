package core

import (
	"sync"
	"time"
)

// WorkerSnapshot describes the local worker's polling lifecycle. LastSuccess
// and LastFailure refer to polling the queue, including successful empty polls.
type WorkerSnapshot struct {
	State       string     `json:"state"`
	LastSuccess *time.Time `json:"lastSuccess,omitempty"`
	LastFailure *time.Time `json:"lastFailure,omitempty"`
}

type WorkerStatus struct {
	mu          sync.Mutex
	state       string
	lastSuccess time.Time
	lastFailure time.Time
}

// RecordPoll records a completed attempt to claim work.
func (s *WorkerStatus) RecordPoll(at time.Time, err error) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state == "stopped" {
		return
	}
	if err != nil {
		s.lastFailure = at.UTC()
		s.state = "degraded"
		return
	}
	s.lastSuccess = at.UTC()
	s.state = "ready"
}

func (s *WorkerStatus) Stop() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = "stopped"
}

func (s *WorkerStatus) Snapshot() WorkerSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := WorkerSnapshot{State: s.state}
	if result.State == "" {
		result.State = "starting"
	}
	if !s.lastSuccess.IsZero() {
		success := s.lastSuccess
		result.LastSuccess = &success
	}
	if !s.lastFailure.IsZero() {
		failure := s.lastFailure
		result.LastFailure = &failure
	}
	return result
}
