package notification_delivery

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type NotificationDeliveryRepository interface {
	Create(ctx context.Context, arg db.CreateNotificationDeliveryParams) (db.NotificationDelivery, error)

	ListByNotificationID(ctx context.Context, notificationID uuid.UUID) ([]db.NotificationDelivery, error)
	GetByID(ctx context.Context, id uuid.UUID) (db.NotificationDelivery, error)

	ClaimPending(ctx context.Context, channel string, limit int32) ([]db.NotificationDelivery, error)

	MarkSent(ctx context.Context, id uuid.UUID) error
	MarkRetry(ctx context.Context, arg db.MarkNotificationDeliveryRetryParams) error
	MarkFailed(ctx context.Context, arg db.MarkNotificationDeliveryFailedParams) error
	Cancel(ctx context.Context, id uuid.UUID) error

	CountPendingByChannel(ctx context.Context, channel string) (int64, error)
	List(ctx context.Context, arg db.ListNotificationDeliveriesParams) ([]db.NotificationDelivery, error)
	Count(ctx context.Context, arg db.CountNotificationDeliveriesParams) (int64, error)
}
