package notification_delivery

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/sqlc-dev/pqtype"
)

type MarkSentWithProviderParams struct {
	ID                uuid.UUID
	Provider          sql.NullString
	ProviderMessageID sql.NullString
	ProviderStatus    sql.NullString
	ProviderResponse  pqtype.NullRawMessage
}

type NotificationDeliveryRepository interface {
	Create(ctx context.Context, arg db.CreateNotificationDeliveryParams) (db.NotificationDelivery, error)

	ListByNotificationID(ctx context.Context, notificationID uuid.UUID) ([]db.NotificationDelivery, error)
	GetByID(ctx context.Context, id uuid.UUID) (db.NotificationDelivery, error)

	ClaimPending(ctx context.Context, channel string, limit int32) ([]db.NotificationDelivery, error)

	MarkSent(ctx context.Context, id uuid.UUID) error
	MarkSentWithProvider(ctx context.Context, arg MarkSentWithProviderParams) error
	MarkRetry(ctx context.Context, arg db.MarkNotificationDeliveryRetryParams) error
	MarkFailed(ctx context.Context, arg db.MarkNotificationDeliveryFailedParams) error
	Cancel(ctx context.Context, id uuid.UUID) error

	CountPendingByChannel(ctx context.Context, channel string) (int64, error)
	List(ctx context.Context, arg db.ListNotificationDeliveriesParams) ([]db.NotificationDelivery, error)
	Count(ctx context.Context, arg db.CountNotificationDeliveriesParams) (int64, error)
	ListMetrics(ctx context.Context) ([]db.ListNotificationDeliveryMetricsRow, error)
}
