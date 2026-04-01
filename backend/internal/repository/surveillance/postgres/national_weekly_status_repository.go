package postgres

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type NationalWeeklyStatusRepository struct {
	db db.Store
}

func NewNationalWeeklyStatusRepository(db db.Store) *NationalWeeklyStatusRepository {
	return &NationalWeeklyStatusRepository{db: db}
}

func (r *NationalWeeklyStatusRepository) CreateDiseaseStatus(ctx context.Context, arg db.CreateNationalWeeklyDiseaseStatusParams) (db.NationalWeeklyStatus, error) {
	return r.db.CreateNationalWeeklyDiseaseStatus(ctx, arg)
}

func (r *NationalWeeklyStatusRepository) UpsertDiseaseStatus(ctx context.Context, arg db.UpsertNationalWeeklyDiseaseStatusParams) (db.NationalWeeklyStatus, error) {
	return r.db.UpsertNationalWeeklyDiseaseStatus(ctx, arg)
}

func (r *NationalWeeklyStatusRepository) CreateIndicatorStatus(ctx context.Context, arg db.CreateNationalWeeklyIndicatorStatusParams) (db.NationalWeeklyStatus, error) {
	return r.db.CreateNationalWeeklyIndicatorStatus(ctx, arg)
}

func (r *NationalWeeklyStatusRepository) UpsertIndicatorStatus(ctx context.Context, arg db.UpsertNationalWeeklyIndicatorStatusParams) (db.NationalWeeklyStatus, error) {
	return r.db.UpsertNationalWeeklyIndicatorStatus(ctx, arg)
}

func (r *NationalWeeklyStatusRepository) ListDiseaseStatusesByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListNationalWeeklyDiseaseStatusesByWeekRow, error) {
	return r.db.ListNationalWeeklyDiseaseStatusesByWeek(ctx, epiWeekID)
}

func (r *NationalWeeklyStatusRepository) ListIndicatorStatusesByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListNationalWeeklyIndicatorStatusesByWeekRow, error) {
	return r.db.ListNationalWeeklyIndicatorStatusesByWeek(ctx, epiWeekID)
}

func (r *NationalWeeklyStatusRepository) DeleteByWeek(ctx context.Context, epiWeekID uuid.UUID) error {
	return r.db.DeleteNationalStatusesByWeek(ctx, epiWeekID)
}

func (r *NationalWeeklyStatusRepository) WithTx(ctx context.Context, fn func(q db.Querier) error) error {
	return r.db.ExecTx(ctx, fn)
}
