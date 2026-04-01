package postgres

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type RegionWeeklyStatusRepository struct {
	db db.Store
}

func NewRegionWeeklyStatusRepository(db db.Store) *RegionWeeklyStatusRepository {
	return &RegionWeeklyStatusRepository{db: db}
}

func (r *RegionWeeklyStatusRepository) CreateDiseaseStatus(ctx context.Context, arg db.CreateRegionWeeklyDiseaseStatusParams) (db.RegionWeeklyStatus, error) {
	return r.db.CreateRegionWeeklyDiseaseStatus(ctx, arg)
}

func (r *RegionWeeklyStatusRepository) UpsertDiseaseStatus(ctx context.Context, arg db.UpsertRegionWeeklyDiseaseStatusParams) (db.RegionWeeklyStatus, error) {
	return r.db.UpsertRegionWeeklyDiseaseStatus(ctx, arg)
}

func (r *RegionWeeklyStatusRepository) CreateIndicatorStatus(ctx context.Context, arg db.CreateRegionWeeklyIndicatorStatusParams) (db.RegionWeeklyStatus, error) {
	return r.db.CreateRegionWeeklyIndicatorStatus(ctx, arg)
}

func (r *RegionWeeklyStatusRepository) UpsertIndicatorStatus(ctx context.Context, arg db.UpsertRegionWeeklyIndicatorStatusParams) (db.RegionWeeklyStatus, error) {
	return r.db.UpsertRegionWeeklyIndicatorStatus(ctx, arg)
}

func (r *RegionWeeklyStatusRepository) ListDiseaseStatusesByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListRegionWeeklyDiseaseStatusesByWeekRow, error) {
	return r.db.ListRegionWeeklyDiseaseStatusesByWeek(ctx, epiWeekID)
}

func (r *RegionWeeklyStatusRepository) ListIndicatorStatusesByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListRegionWeeklyIndicatorStatusesByWeekRow, error) {
	return r.db.ListRegionWeeklyIndicatorStatusesByWeek(ctx, epiWeekID)
}

func (r *RegionWeeklyStatusRepository) DeleteByWeek(ctx context.Context, epiWeekID uuid.UUID) error {
	return r.db.DeleteRegionStatusesByWeek(ctx, epiWeekID)
}

func (r *RegionWeeklyStatusRepository) WithTx(ctx context.Context, fn func(q db.Querier) error) error {
	return r.db.ExecTx(ctx, fn)
}
