package worker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/model"
	notificationDeliveryRepo "github.com/moh-sso-dashboard/internal/repository/notification_delivery"
	"github.com/moh-sso-dashboard/internal/service"
)

type NotificationSMSDeliveryWorker struct {
	notificationDelivery notificationDeliveryRepo.NotificationDeliveryRepository
	smsSvc               service.SMSService
	pollDelay            time.Duration
	batchSize            int32
	maxRetries           int32
	logger               *logger.Logger
}

func NewNotificationSMSDeliveryWorker(
	notificationDelivery notificationDeliveryRepo.NotificationDeliveryRepository,
	smsSvc service.SMSService,
	pollDelay time.Duration,
	batchSize int32,
	maxRetries int32,
	logger *logger.Logger,
) (*NotificationSMSDeliveryWorker, error) {
	if notificationDelivery == nil {
		return nil, errors.New("notification delivery repository is required")
	}
	if smsSvc == nil {
		return nil, errors.New("sms service is required")
	}
	if logger == nil {
		return nil, errors.New("logger is required")
	}
	if pollDelay <= 0 {
		pollDelay = 5 * time.Second
	}
	if batchSize <= 0 {
		batchSize = 20
	}
	if maxRetries <= 0 {
		maxRetries = 3
	}

	return &NotificationSMSDeliveryWorker{
		notificationDelivery: notificationDelivery,
		smsSvc:               smsSvc,
		pollDelay:            pollDelay,
		batchSize:            batchSize,
		maxRetries:           maxRetries,
		logger:               logger,
	}, nil
}

func (w *NotificationSMSDeliveryWorker) Start(ctx context.Context) error {
	if w == nil {
		return errors.New("notification sms delivery worker is nil")
	}

	w.logger.Info(
		"notification sms delivery worker started",
		"poll_delay", w.pollDelay.String(),
		"batch_size", w.batchSize,
		"max_retries", w.maxRetries,
	)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("notification sms delivery worker stopped")
			return ctx.Err()
		default:
		}

		if err := w.process(ctx); err != nil {
			w.logger.Error("notification sms delivery worker error", "error", err)
		}

		sleepWithContext(ctx, w.pollDelay)
	}
}

func (w *NotificationSMSDeliveryWorker) process(ctx context.Context) error {
	items, err := w.notificationDelivery.ClaimPending(
		ctx,
		string(model.NotificationChannelSMS),
		w.batchSize,
	)
	if err != nil {
		return fmt.Errorf("claim sms notification deliveries: %w", err)
	}

	for _, item := range items {
		if err := w.processOne(ctx, item); err != nil {
			w.logger.Error(
				"failed to process notification sms delivery",
				"error", err,
				"delivery_id", item.ID,
				"notification_id", item.NotificationID,
			)
		}
	}

	return nil
}

func (w *NotificationSMSDeliveryWorker) processOne(ctx context.Context, item db.NotificationDelivery) error {
	msg, err := notificationDeliveryToSMSMessage(item)
	if err != nil {
		return w.failOrRetry(ctx, item, err)
	}

	result, err := w.smsSvc.Send(ctx, msg)
	if err != nil {
		return w.failOrRetry(ctx, item, err)
	}

	if err := w.notificationDelivery.MarkSent(ctx, item.ID); err != nil {
		return fmt.Errorf("mark notification sms delivery sent: %w", err)
	}

	w.logger.Info(
		"notification sms delivery sent",
		"delivery_id", item.ID,
		"notification_id", item.NotificationID,
		"provider", result.Provider,
		"message_id", result.MessageID,
		"status", result.Status,
	)

	return nil
}

func (w *NotificationSMSDeliveryWorker) failOrRetry(
	ctx context.Context,
	item db.NotificationDelivery,
	cause error,
) error {
	nextAttempts := item.Attempts + 1
	if nextAttempts >= item.MaxAttempts {
		if err := w.notificationDelivery.MarkFailed(
			ctx,
			db.MarkNotificationDeliveryFailedParams{
				ID: item.ID,
				LastError: sql.NullString{
					String: cause.Error(),
					Valid:  true,
				},
			},
		); err != nil {
			return fmt.Errorf("mark notification sms delivery failed: %w", err)
		}

		return cause
	}

	delay := retryDelay(nextAttempts)
	if err := w.notificationDelivery.MarkRetry(
		ctx,
		db.MarkNotificationDeliveryRetryParams{
			ID: item.ID,
			LastError: sql.NullString{
				String: cause.Error(),
				Valid:  true,
			},
			Column3: delay.String(),
		},
	); err != nil {
		return fmt.Errorf("mark notification sms delivery retry: %w", err)
	}

	return cause
}

type smsRecipient struct {
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	PhoneNumber string `json:"phone_number"`
	To          string `json:"to"`
}

type smsPayload struct {
	Body    string `json:"body"`
	Message string `json:"message"`
}

func notificationDeliveryToSMSMessage(item db.NotificationDelivery) (service.SMSMessage, error) {
	var recipient smsRecipient
	if err := unmarshalNullRawMessage(item.Recipient, &recipient); err != nil {
		return service.SMSMessage{}, fmt.Errorf("decode recipient: %w", err)
	}

	to := firstNonEmpty(recipient.Phone, recipient.PhoneNumber, recipient.To)
	if strings.TrimSpace(to) == "" {
		return service.SMSMessage{}, errors.New("sms recipient phone is required")
	}

	var payload smsPayload
	if err := unmarshalNullRawMessage(item.Payload, &payload); err != nil {
		return service.SMSMessage{}, fmt.Errorf("decode payload: %w", err)
	}

	body := firstNonEmpty(payload.Body, payload.Message)
	if strings.TrimSpace(body) == "" {
		return service.SMSMessage{}, errors.New("sms body is required")
	}

	return service.SMSMessage{
		To:   to,
		Body: body,
		Metadata: map[string]string{
			"source":           "notification-delivery-worker",
			"notification_id":  item.NotificationID.String(),
			"delivery_id":      item.ID.String(),
			"delivery_channel": item.Channel,
		},
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
