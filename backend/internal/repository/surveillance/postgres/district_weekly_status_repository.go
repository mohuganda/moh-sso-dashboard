package postgres

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type DistrictWeeklyStatusRepository struct {
	db db.Store
}

func NewDistrictWeeklyStatusRepository(db db.Store) *DistrictWeeklyStatusRepository {
	return &DistrictWeeklyStatusRepository{db: db}
}

func (r *DistrictWeeklyStatusRepository) CreateDiseaseStatus(ctx context.Context, arg db.CreateDistrictWeeklyDiseaseStatusParams) (db.DistrictWeeklyStatus, error) {
	return r.db.CreateDistrictWeeklyDiseaseStatus(ctx, arg)
}

func (r *DistrictWeeklyStatusRepository) UpsertDiseaseStatus(ctx context.Context, arg db.UpsertDistrictWeeklyDiseaseStatusParams) (db.DistrictWeeklyStatus, error) {
	return r.db.UpsertDistrictWeeklyDiseaseStatus(ctx, arg)
}

func (r *DistrictWeeklyStatusRepository) CreateIndicatorStatus(ctx context.Context, arg db.CreateDistrictWeeklyIndicatorStatusParams) (db.DistrictWeeklyStatus, error) {
	return r.db.CreateDistrictWeeklyIndicatorStatus(ctx, arg)
}

func (r *DistrictWeeklyStatusRepository) UpsertIndicatorStatus(ctx context.Context, arg db.UpsertDistrictWeeklyIndicatorStatusParams) (db.DistrictWeeklyStatus, error) {
	return r.db.UpsertDistrictWeeklyIndicatorStatus(ctx, arg)
}

func (r *DistrictWeeklyStatusRepository) ListDiseaseStatusesByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListDistrictWeeklyDiseaseStatusesByWeekRow, error) {
	return r.db.ListDistrictWeeklyDiseaseStatusesByWeek(ctx, epiWeekID)
}

func (r *DistrictWeeklyStatusRepository) ListIndicatorStatusesByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListDistrictWeeklyIndicatorStatusesByWeekRow, error) {
	return r.db.ListDistrictWeeklyIndicatorStatusesByWeek(ctx, epiWeekID)
}

func (r *DistrictWeeklyStatusRepository) ListByDistrictAndWeek(ctx context.Context, arg db.ListDistrictStatusesByDistrictAndWeekParams) ([]db.ListDistrictStatusesByDistrictAndWeekRow, error) {
	return r.db.ListDistrictStatusesByDistrictAndWeek(ctx, arg)
}

func (r *DistrictWeeklyStatusRepository) DeleteByWeek(ctx context.Context, epiWeekID uuid.UUID) error {
	return r.db.DeleteDistrictStatusesByWeek(ctx, epiWeekID)
}

func (r *DistrictWeeklyStatusRepository) WithTx(ctx context.Context, fn func(q db.Querier) error) error {
	return r.db.ExecTx(ctx, fn)
}
