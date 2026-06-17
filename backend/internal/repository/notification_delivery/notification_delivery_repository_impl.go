package notification_delivery

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
)

type notificationDeliveryRepository struct {
	store  db.Store
	logger *logger.Logger
}

func NewNotificationDeliveryRepository(
	store db.Store,
	logger logger.Logger,
) NotificationDeliveryRepository {

	return &notificationDeliveryRepository{
		store:  store,
		logger: &logger,
	}
}

func (r *notificationDeliveryRepository) Create(
	ctx context.Context,
	arg db.CreateNotificationDeliveryParams,
) (db.NotificationDelivery, error) {
	delivery, err := r.store.CreateNotificationDelivery(ctx, arg)
	if err != nil {
		r.logger.Error(
			"failed to create notification delivery",
			"error", err,
			"notification_id", arg.NotificationID,
			"channel", arg.Channel,
		)

		return db.NotificationDelivery{}, fmt.Errorf("create notification delivery: %w", err)
	}

	r.logger.Info(
		"notification delivery created",
		"delivery_id", delivery.ID,
		"notification_id", delivery.NotificationID,
		"channel", delivery.Channel,
		"status", delivery.Status,
	)

	return delivery, nil
}

func (r *notificationDeliveryRepository) ListByNotificationID(
	ctx context.Context,
	notificationID uuid.UUID,
) ([]db.NotificationDelivery, error) {
	deliveries, err := r.store.ListNotificationDeliveriesByNotificationID(ctx, notificationID)
	if err != nil {
		return nil, fmt.Errorf("list notification deliveries by notification id: %w", err)
	}

	return deliveries, nil
}

func (r *notificationDeliveryRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (db.NotificationDelivery, error) {
	delivery, err := r.store.GetNotificationDeliveryByID(ctx, id)
	if err != nil {
		return db.NotificationDelivery{}, fmt.Errorf("get notification delivery by id: %w", err)
	}

	return delivery, nil
}

func (r *notificationDeliveryRepository) ClaimPending(
	ctx context.Context,
	channel string,
	limit int32,
) ([]db.NotificationDelivery, error) {
	deliveries, err := r.store.ClaimPendingNotificationDeliveries(
		ctx,
		db.ClaimPendingNotificationDeliveriesParams{
			Channel: channel,
			Limit:   limit,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("claim pending notification deliveries: %w", err)
	}

	return deliveries, nil
}

func (r *notificationDeliveryRepository) MarkSent(
	ctx context.Context,
	id uuid.UUID,
) error {
	if err := r.store.MarkNotificationDeliverySent(ctx, id); err != nil {
		return fmt.Errorf("mark notification delivery sent: %w", err)
	}

	return nil
}

func (r *notificationDeliveryRepository) MarkRetry(
	ctx context.Context,
	arg db.MarkNotificationDeliveryRetryParams,
) error {
	if err := r.store.MarkNotificationDeliveryRetry(ctx, arg); err != nil {
		return fmt.Errorf("mark notification delivery retry: %w", err)
	}

	return nil
}

func (r *notificationDeliveryRepository) MarkFailed(
	ctx context.Context,
	arg db.MarkNotificationDeliveryFailedParams,
) error {
	if err := r.store.MarkNotificationDeliveryFailed(ctx, arg); err != nil {
		return fmt.Errorf("mark notification delivery failed: %w", err)
	}

	return nil
}

func (r *notificationDeliveryRepository) Cancel(
	ctx context.Context,
	id uuid.UUID,
) error {
	if err := r.store.CancelNotificationDelivery(ctx, id); err != nil {
		return fmt.Errorf("cancel notification delivery: %w", err)
	}

	return nil
}

func (r *notificationDeliveryRepository) CountPendingByChannel(
	ctx context.Context,
	channel string,
) (int64, error) {
	count, err := r.store.CountPendingNotificationDeliveriesByChannel(ctx, channel)
	if err != nil {
		return 0, fmt.Errorf("count pending notification deliveries by channel: %w", err)
	}

	return count, nil
}

func (r *notificationDeliveryRepository) List(
	ctx context.Context,
	arg db.ListNotificationDeliveriesParams,
) ([]db.NotificationDelivery, error) {
	deliveries, err := r.store.ListNotificationDeliveries(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("list notification deliveries: %w", err)
	}

	return deliveries, nil
}

func (r *notificationDeliveryRepository) Count(
	ctx context.Context,
	arg db.CountNotificationDeliveriesParams,
) (int64, error) {
	count, err := r.store.CountNotificationDeliveries(ctx, arg)
	if err != nil {
		return 0, fmt.Errorf("count notification deliveries: %w", err)
	}

	return count, nil
}
