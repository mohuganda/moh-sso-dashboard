package events

import (
	"context"
	"encoding/json"
	"time"
)

type Event struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	Source        string          `json:"source"`
	Subject       string          `json:"subject,omitempty"`
	CorrelationID string          `json:"correlation_id,omitempty"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Payload       json.RawMessage `json:"payload,omitempty"`
}

type Publisher interface {
	Publish(ctx context.Context, event Event) error
}

type Handler func(ctx context.Context, event Event) error
