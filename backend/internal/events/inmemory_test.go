package events

import (
	"context"
	"testing"

	"github.com/moh-sso-dashboard/internal/observability"
)

func TestInMemoryPublisherPublishesToSubscribers(t *testing.T) {
	publisher := NewInMemoryPublisher()
	ctx := observability.WithCorrelationID(context.Background(), "corr-123")

	var received Event
	publisher.Subscribe("announcement.published", func(_ context.Context, event Event) error {
		received = event
		return nil
	})

	err := publisher.Publish(ctx, Event{
		Type:    "announcement.published",
		Source:  "announcements",
		Subject: "announcement-1",
	})
	if err != nil {
		t.Fatal(err)
	}

	if received.ID == "" {
		t.Fatal("expected generated event id")
	}
	if received.CorrelationID != "corr-123" {
		t.Fatalf("expected correlation id corr-123, got %q", received.CorrelationID)
	}
	if received.OccurredAt.IsZero() {
		t.Fatal("expected occurred_at to be populated")
	}
}
