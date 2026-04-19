package email

import (
	"context"
	"database/sql"

	"github.com/moh-sso-dashboard/internal/model"
)

type EmailRepository interface {
	Enqueue(ctx context.Context, msg model.Message) (*OutboxMessage, error)
	GetByID(ctx context.Context, id string) (*OutboxMessage, error)
	List(ctx context.Context, limit, offset int32) ([]OutboxMessage, error)
	ListByStatus(ctx context.Context, status string, limit, offset int32) ([]OutboxMessage, error)
	ClaimNextBatch(ctx context.Context, limit int32) ([]OutboxMessage, error)
	MarkProcessing(ctx context.Context, id string) error
	MarkSent(ctx context.Context, id string) error
	MarkRetry(ctx context.Context, id string, attempts int32, lastErr string) error
	MarkFailed(ctx context.Context, id string, attempts int32, lastErr string) error
	ResetStuckJobs(ctx context.Context) (int64, error)
	DeleteByID(ctx context.Context, id string) error
	DeleteSentOlderThan(context.Context, sql.NullTime) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
}
