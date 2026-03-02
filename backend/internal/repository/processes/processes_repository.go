package processes

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/model"
)

type ProcessRepository interface {
	CreateProcess(ctx context.Context, arg db.CreateProcessParams) (db.Process, error)
	GetProcessByID(ctx context.Context, id uuid.UUID) (db.Process, error)
	ListProcesses(ctx context.Context, filters model.ProcessFilters, page model.Pagination) ([]db.Process, error)
	ClaimNextPending(ctx context.Context) (db.Process, error)
	UpdateProgress(ctx context.Context, id uuid.UUID, progress int32, message *string) error
	Complete(ctx context.Context, id uuid.UUID) error
	Fail(ctx context.Context, id uuid.UUID, errMsg string) error
	Cancel(ctx context.Context, id uuid.UUID) error
}
