package events

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/moh-sso-dashboard/internal/observability"
)

type InMemoryPublisher struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

func NewInMemoryPublisher() *InMemoryPublisher {
	return &InMemoryPublisher{
		handlers: map[string][]Handler{},
	}
}

func (p *InMemoryPublisher) Subscribe(eventType string, handler Handler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handlers[eventType] = append(p.handlers[eventType], handler)
}

func (p *InMemoryPublisher) Publish(ctx context.Context, event Event) error {
	if event.ID == "" {
		event.ID = uuid.NewString()
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now().UTC()
	}
	if event.CorrelationID == "" {
		event.CorrelationID = observability.CorrelationIDFromContext(ctx)
	}

	p.mu.RLock()
	handlers := append([]Handler(nil), p.handlers[event.Type]...)
	p.mu.RUnlock()

	log.Info().
		Str("event_id", event.ID).
		Str("event_type", event.Type).
		Str("source", event.Source).
		Str("subject", event.Subject).
		Str("correlation_id", event.CorrelationID).
		Int("handlers", len(handlers)).
		Msg("domain event published")

	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

type NoopPublisher struct{}

func (NoopPublisher) Publish(context.Context, Event) error {
	return nil
}
