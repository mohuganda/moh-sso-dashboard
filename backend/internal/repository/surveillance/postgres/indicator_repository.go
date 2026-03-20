package postgres

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type IndicatorRepository struct {
	db db.Store
}

func NewIndicatorRepository(db db.Store) *IndicatorRepository {
	return &IndicatorRepository{db: db}
}

func (r *IndicatorRepository) Create(ctx context.Context, arg db.CreateIndicatorParams) (db.Indicator, error) {
	return r.db.CreateIndicator(ctx, arg)
}

func (r *IndicatorRepository) GetByID(ctx context.Context, id uuid.UUID) (db.Indicator, error) {
	return r.db.GetIndicatorByID(ctx, id)
}

func (r *IndicatorRepository) GetByName(ctx context.Context, name string) (db.Indicator, error) {
	return r.db.GetIndicatorByName(ctx, name)
}

func (r *IndicatorRepository) List(ctx context.Context) ([]db.Indicator, error) {
	return r.db.ListIndicators(ctx)
}

func (r *IndicatorRepository) ListActive(ctx context.Context) ([]db.Indicator, error) {
	return r.db.ListActiveIndicators(ctx)
}

func (r *IndicatorRepository) Upsert(ctx context.Context, arg db.UpsertIndicatorParams) (db.Indicator, error) {
	return r.db.UpsertIndicator(ctx, arg)
}

func (r *IndicatorRepository) SetActiveState(ctx context.Context, arg db.SetIndicatorActiveStateParams) (db.Indicator, error) {
	return r.db.SetIndicatorActiveState(ctx, arg)
}

func (r *IndicatorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.DeleteIndicator(ctx, id)
}
