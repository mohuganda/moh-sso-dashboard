package postgres

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type EpiWeekRepository struct {
	db db.Store
}

func NewEpiWeekRepository(db db.Store) *EpiWeekRepository {
	return &EpiWeekRepository{db: db}
}

func (r *EpiWeekRepository) Create(ctx context.Context, arg db.CreateEpiWeekParams) (db.EpiWeek, error) {
	return r.db.CreateEpiWeek(ctx, arg)
}

func (r *EpiWeekRepository) GetByID(ctx context.Context, id uuid.UUID) (db.EpiWeek, error) {
	return r.db.GetEpiWeekByID(ctx, id)
}

func (r *EpiWeekRepository) GetByYearWeek(ctx context.Context, arg db.GetEpiWeekByYearWeekParams) (db.EpiWeek, error) {
	return r.db.GetEpiWeekByYearWeek(ctx, arg)
}

func (r *EpiWeekRepository) ListByYear(ctx context.Context, year int32) ([]db.EpiWeek, error) {
	return r.db.ListEpiWeeksByYear(ctx, year)
}

func (r *EpiWeekRepository) Upsert(ctx context.Context, arg db.UpsertEpiWeekParams) (db.EpiWeek, error) {
	return r.db.UpsertEpiWeek(ctx, arg)
}
