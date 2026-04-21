package email

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/config"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/model"
)

type emailRepository struct {
	db     db.Store
	config *config.Config
	logger *logger.Logger
}

func NewEmailRepository(
	cfg *config.Config,
	store db.Store,
	log logger.Logger) EmailRepository {
	return &emailRepository{
		db:     store,
		config: cfg,
		logger: &log,
	}
}

func (r *emailRepository) Enqueue(ctx context.Context, msg model.Message) (*OutboxMessage, error) {
	if err := validateMessage(msg); err != nil {
		return nil, err
	}

	params, err := toCreateEmailOutboxParams(msg)
	if err != nil {
		return nil, fmt.Errorf("build create email outbox params: %w", err)
	}

	row, err := r.db.CreateEmailOutbox(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create email outbox: %w", err)
	}

	out, err := mapEmailOutboxRow(row)
	if err != nil {
		return nil, fmt.Errorf("map created email outbox: %w", err)
	}

	return &out, nil
}

func (r *emailRepository) GetByID(ctx context.Context, id string) (*OutboxMessage, error) {
	row, err := r.db.GetEmailOutboxByID(ctx, parseUUID(id))
	if err != nil {
		return nil, fmt.Errorf("get email outbox by id: %w", err)
	}

	out, err := mapEmailOutboxRow(row)
	if err != nil {
		return nil, fmt.Errorf("map email outbox by id: %w", err)
	}

	return &out, nil
}

func (r *emailRepository) List(ctx context.Context, limit, offset int32) ([]OutboxMessage, error) {
	rows, err := r.db.ListEmailOutbox(ctx, db.ListEmailOutboxParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list email outbox: %w", err)
	}

	return mapEmailOutboxRows(rows)
}

func (r *emailRepository) ListByStatus(ctx context.Context, status string, limit, offset int32) ([]OutboxMessage, error) {
	rows, err := r.db.ListEmailOutboxByStatus(ctx, db.ListEmailOutboxByStatusParams{
		Status: status,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list email outbox by status: %w", err)
	}

	return mapEmailOutboxRows(rows)
}

func (r *emailRepository) ClaimNextBatch(ctx context.Context, limit int32) ([]OutboxMessage, error) {
	rows, err := r.db.ClaimEmailOutboxBatch(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("claim email outbox batch: %w", err)
	}

	return mapEmailOutboxRows(rows)
}

func (r *emailRepository) MarkProcessing(ctx context.Context, id string) error {
	err := r.db.MarkEmailOutboxProcessing(ctx, parseUUID(id))
	if err != nil {
		return fmt.Errorf("mark email outbox processing: %w", err)
	}
	return nil
}

func (r *emailRepository) MarkSent(ctx context.Context, id string) error {
	err := r.db.MarkEmailOutboxSent(ctx, parseUUID(id))
	if err != nil {
		return fmt.Errorf("mark email outbox sent: %w", err)
	}
	return nil
}

func (r *emailRepository) MarkRetry(ctx context.Context, id string, attempts int32, lastErr string) error {
	err := r.db.MarkEmailOutboxRetry(ctx, db.MarkEmailOutboxRetryParams{
		ID:        parseUUID(id),
		Attempts:  attempts,
		LastError: sql.NullString{String: lastErr, Valid: lastErr != ""},
	})
	if err != nil {
		return fmt.Errorf("mark email outbox retry: %w", err)
	}
	return nil
}

func (r *emailRepository) MarkFailed(ctx context.Context, id string, attempts int32, lastErr string) error {
	err := r.db.MarkEmailOutboxFailed(ctx, db.MarkEmailOutboxFailedParams{
		ID:        parseUUID(id),
		Attempts:  attempts,
		LastError: sql.NullString{String: lastErr, Valid: lastErr != ""},
	})
	if err != nil {
		return fmt.Errorf("mark email outbox failed: %w", err)
	}
	return nil
}

func (r *emailRepository) ResetStuckJobs(ctx context.Context) (int64, error) {
	n, err := r.db.ResetStuckEmailOutboxJobs(ctx)
	if err != nil {
		return 0, fmt.Errorf("reset stuck email outbox jobs: %w", err)
	}
	return n, nil
}

func (r *emailRepository) DeleteByID(ctx context.Context, id string) error {
	err := r.db.DeleteEmailOutboxByID(ctx, parseUUID(id))
	if err != nil {
		return fmt.Errorf("delete email outbox by id: %w", err)
	}
	return nil
}

func (r *emailRepository) DeleteSentOlderThan(ctx context.Context, sentBefore sql.NullTime) (int64, error) {
	n, err := r.db.DeleteSentEmailOutboxOlderThan(ctx, sentBefore)
	if err != nil {
		return 0, fmt.Errorf("delete sent email outbox older than: %w", err)
	}
	return n, nil
}

func (r *emailRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	n, err := r.db.CountEmailOutboxByStatus(ctx, status)
	if err != nil {
		return 0, fmt.Errorf("count email outbox by status: %w", err)
	}
	return n, nil
}

func parseUUID(v string) uuid.UUID {
	id, err := uuid.Parse(v)
	if err != nil {
		return uuid.Nil
	}
	return id
}

func validateMessage(msg model.Message) error {
	if len(msg.To) == 0 {
		return errors.New("at least one recipient is required")
	}

	for _, to := range msg.To {
		if strings.TrimSpace(to.Email) == "" {
			return errors.New("recipient email cannot be empty")
		}
	}

	if strings.TrimSpace(msg.Subject) == "" {
		return errors.New("subject is required")
	}

	if strings.TrimSpace(msg.TextBody) == "" {
		return errors.New("body is required")
	}

	return nil
}
