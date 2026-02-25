package processes

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/config"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/model"
)

type processRepository struct {
	db     db.Store
	config *config.Config
	logger *logger.Logger
}

func NewProcessRepository(
	cfg *config.Config,
	store db.Store,
	log logger.Logger,
) ProcessRepository {
	return &processRepository{
		db:     store,
		config: cfg,
		logger: &log,
	}
}

func (r *processRepository) CreateProcess(
	ctx context.Context,
	arg db.CreateProcessParams,
) (db.Process, error) {

	proc, err := r.db.CreateProcess(ctx, arg)
	if err != nil {
		return db.Process{}, err
	}

	return proc, nil
}

func (r *processRepository) GetProcessByID(
	ctx context.Context,
	id uuid.UUID,
) (db.Process, error) {

	proc, err := r.db.GetProcessByID(ctx, id)
	if err != nil {
		return db.Process{}, err
	}

	return proc, nil
}

func (r *processRepository) ListProcesses(
	ctx context.Context,
	filters model.ProcessFilters,
	page model.Pagination,
) ([]db.Process, error) {

	limit := page.Limit
	if limit <= 0 {
		limit = 25
	}

	procs, err := r.db.ListProcesses(ctx, db.ListProcessesParams{
		Limit:  limit,
		Offset: page.Offset,
	})
	if err != nil {
		return nil, err
	}

	return procs, nil
}

func (r *processRepository) ClaimNextPending(
	ctx context.Context,
) (db.Process, error) {

	proc, err := r.db.ClaimNextPendingProcess(ctx)
	if err != nil {
		return db.Process{}, err
	}

	return proc, nil
}

func (r *processRepository) UpdateProgress(
	ctx context.Context,
	id uuid.UUID,
	progress int32,
	message *string,
) error {

	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}

	err := r.db.UpdateProcessProgress(ctx, db.UpdateProcessProgressParams{
		ID:       id,
		Progress: progress,
		Message: sql.NullString{
			String: *message,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *processRepository) Complete(
	ctx context.Context,
	id uuid.UUID,
) error {

	err := r.db.CompleteProcess(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *processRepository) Fail(
	ctx context.Context,
	id uuid.UUID,
	errMsg string,
) error {

	err := r.db.FailProcess(ctx, db.FailProcessParams{
		ID: id,
		Error: sql.NullString{
			String: errMsg,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *processRepository) Cancel(
	ctx context.Context,
	id uuid.UUID,
) error {

	err := r.db.CancelProcess(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
