package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/model"
	"github.com/sqlc-dev/pqtype"
)

type EmailQueueService interface {
	Queue(ctx context.Context, msg model.Message) error
}

type NotificationEmailDeliveryWorker struct {
	store      db.Store
	emailSvc   EmailQueueService
	pollDelay  time.Duration
	batchSize  int32
	maxRetries int32
	logger     *logger.Logger
}

func NewNotificationEmailDeliveryWorker(
	store db.Store,
	emailSvc EmailQueueService,
	pollDelay time.Duration,
	batchSize int32,
	maxRetries int32,
	logger *logger.Logger,
) (*NotificationEmailDeliveryWorker, error) {
	if store == nil {
		return nil, errors.New("store is required")
	}

	if emailSvc == nil {
		return nil, errors.New("email service is required")
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

	return &NotificationEmailDeliveryWorker{
		store:      store,
		emailSvc:   emailSvc,
		pollDelay:  pollDelay,
		batchSize:  batchSize,
		maxRetries: maxRetries,
		logger:     logger,
	}, nil
}

func (w *NotificationEmailDeliveryWorker) Start(ctx context.Context) error {
	if w == nil {
		return errors.New("notification email delivery worker is nil")
	}

	w.logger.Info(
		"notification email delivery worker started",
		"poll_delay", w.pollDelay.String(),
		"batch_size", w.batchSize,
	)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("notification email delivery worker stopped")
			return ctx.Err()
		default:
		}

		if err := w.process(ctx); err != nil {
			w.logger.Error(
				"notification email delivery worker error",
				"error", err,
			)
		}

		sleepWithContext(ctx, w.pollDelay)
	}
}

func (w *NotificationEmailDeliveryWorker) process(ctx context.Context) error {
	items, err := w.store.ClaimPendingNotificationDeliveries(
		ctx,
		db.ClaimPendingNotificationDeliveriesParams{
			Channel: string(model.NotificationChannelEmail),
			Limit:   w.batchSize,
		},
	)
	if err != nil {
		return fmt.Errorf("claim email notification deliveries: %w", err)
	}

	if len(items) == 0 {
		return nil
	}

	for _, item := range items {
		if err := w.processOne(ctx, item); err != nil {
			w.logger.Error(
				"failed to process notification email delivery",
				"error", err,
				"delivery_id", item.ID,
				"notification_id", item.NotificationID,
			)
		}
	}

	return nil
}

func (w *NotificationEmailDeliveryWorker) processOne(
	ctx context.Context,
	item db.NotificationDelivery,
) error {
	msg, err := notificationDeliveryToEmailMessage(item)
	if err != nil {
		return w.failOrRetry(ctx, item, err)
	}

	if err := w.emailSvc.Queue(ctx, msg); err != nil {
		return w.failOrRetry(ctx, item, err)
	}

	if err := w.store.MarkNotificationDeliverySent(ctx, item.ID); err != nil {
		return fmt.Errorf("mark notification email delivery sent: %w", err)
	}

	w.logger.Info(
		"notification email delivery queued",
		"delivery_id", item.ID,
		"notification_id", item.NotificationID,
	)

	return nil
}

func (w *NotificationEmailDeliveryWorker) failOrRetry(
	ctx context.Context,
	item db.NotificationDelivery,
	cause error,
) error {
	nextAttempts := item.Attempts + 1

	if nextAttempts >= item.MaxAttempts {
		if err := w.store.MarkNotificationDeliveryFailed(
			ctx,
			db.MarkNotificationDeliveryFailedParams{
				ID:        item.ID,
				LastError: sql.NullString{String: cause.Error(), Valid: true},
			},
		); err != nil {
			return fmt.Errorf("mark notification delivery failed: %w", err)
		}

		return cause
	}

	delay := retryDelay(nextAttempts)

	if err := w.store.MarkNotificationDeliveryRetry(
		ctx,
		db.MarkNotificationDeliveryRetryParams{
			ID:        item.ID,
			LastError: sql.NullString{String: cause.Error(), Valid: true},
			Column3:   delay.String(),
		},
	); err != nil {
		return fmt.Errorf("mark notification delivery retry: %w", err)
	}

	return cause
}

func retryDelay(attempts int32) time.Duration {
	switch {
	case attempts <= 1:
		return 30 * time.Second
	case attempts == 2:
		return 2 * time.Minute
	default:
		return 5 * time.Minute
	}
}

type emailRecipient struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type emailPayload struct {
	Subject     string             `json:"subject"`
	TextBody    string             `json:"text_body"`
	HTMLBody    string             `json:"html_body"`
	Headers     map[string]string  `json:"headers"`
	Attachments []model.Attachment `json:"attachments"`
}

func notificationDeliveryToEmailMessage(
	item db.NotificationDelivery,
) (model.Message, error) {
	var recipient emailRecipient
	if err := unmarshalNullRawMessage(item.Recipient, &recipient); err != nil {
		return model.Message{}, fmt.Errorf("decode recipient: %w", err)
	}

	if strings.TrimSpace(recipient.Email) == "" {
		return model.Message{}, errors.New("email recipient is required")
	}

	var payload emailPayload
	if err := unmarshalNullRawMessage(item.Payload, &payload); err != nil {
		return model.Message{}, fmt.Errorf("decode payload: %w", err)
	}

	var templateData map[string]any
	if err := unmarshalNullRawMessage(item.TemplateData, &templateData); err != nil {
		return model.Message{}, fmt.Errorf("decode template data: %w", err)
	}

	templateName := ""
	if item.TemplateName.Valid {
		templateName = item.TemplateName.String
	}

	subject := strings.TrimSpace(payload.Subject)
	if subject == "" {
		subject = "Notification"
	}

	textBody := strings.TrimSpace(payload.TextBody)
	if textBody == "" {
		textBody = "You have a new notification."
	}

	msg := model.Message{
		To: []model.Address{
			{
				Name:  strings.TrimSpace(recipient.Name),
				Email: strings.TrimSpace(recipient.Email),
			},
		},
		Subject:      subject,
		TextBody:     textBody,
		HTMLBody:     payload.HTMLBody,
		TemplateName: templateName,
		TemplateData: templateData,
		Headers:      payload.Headers,
		Attachments:  payload.Attachments,
		Metadata: map[string]string{
			"source":           "notification-delivery-worker",
			"notification_id":  item.NotificationID.String(),
			"delivery_id":      item.ID.String(),
			"delivery_channel": item.Channel,
		},
	}

	return msg, nil
}

func unmarshalNullRawMessage(value pqtype.NullRawMessage, target any) error {
	if !value.Valid || len(value.RawMessage) == 0 {
		return nil
	}

	return json.Unmarshal(value.RawMessage, target)
}
