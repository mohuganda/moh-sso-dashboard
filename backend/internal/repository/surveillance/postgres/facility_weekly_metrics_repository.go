package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type FacilityWeeklyMetricsRepository struct {
	db db.Store
}

func NewFacilityWeeklyMetricsRepository(db db.Store) *FacilityWeeklyMetricsRepository {
	return &FacilityWeeklyMetricsRepository{db: db}
}

func (r *FacilityWeeklyMetricsRepository) CreateDiseaseMetric(ctx context.Context, arg db.CreateFacilityWeeklyDiseaseMetricParams) (db.FacilityWeeklyMetric, error) {
	return r.db.CreateFacilityWeeklyDiseaseMetric(ctx, arg)
}

func (r *FacilityWeeklyMetricsRepository) UpsertDiseaseMetric(ctx context.Context, arg db.UpsertFacilityWeeklyDiseaseMetricParams) (db.FacilityWeeklyMetric, error) {
	return r.db.UpsertFacilityWeeklyDiseaseMetric(ctx, arg)
}

func (r *FacilityWeeklyMetricsRepository) CreateIndicatorMetric(ctx context.Context, arg db.CreateFacilityWeeklyIndicatorMetricParams) (db.FacilityWeeklyMetric, error) {
	return r.db.CreateFacilityWeeklyIndicatorMetric(ctx, arg)
}

func (r *FacilityWeeklyMetricsRepository) UpsertIndicatorMetric(ctx context.Context, arg db.UpsertFacilityWeeklyIndicatorMetricParams) (db.FacilityWeeklyMetric, error) {
	return r.db.UpsertFacilityWeeklyIndicatorMetric(ctx, arg)
}

func (r *FacilityWeeklyMetricsRepository) GetByID(ctx context.Context, id uuid.UUID) (db.FacilityWeeklyMetric, error) {
	return r.db.GetFacilityWeeklyMetricByID(ctx, id)
}

func (r *FacilityWeeklyMetricsRepository) GetBySourceRecordID(ctx context.Context, sourceRecordID string) (db.FacilityWeeklyMetric, error) {

	if sourceRecordID == "" {
		return db.FacilityWeeklyMetric{}, errors.New("external id is required")
	}
	param := sql.NullString{
		String: sourceRecordID,
		Valid:  true,
	}

	facility, err := r.db.GetFacilityMetricBySourceRecordID(ctx, param)
	if err != nil {
		return db.FacilityWeeklyMetric{}, err
	}

	return facility, nil

}

func (r *FacilityWeeklyMetricsRepository) ListDiseaseMetricsByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListFacilityWeeklyDiseaseMetricsByWeekRow, error) {
	return r.db.ListFacilityWeeklyDiseaseMetricsByWeek(ctx, epiWeekID)
}

func (r *FacilityWeeklyMetricsRepository) ListIndicatorMetricsByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListFacilityWeeklyIndicatorMetricsByWeekRow, error) {
	return r.db.ListFacilityWeeklyIndicatorMetricsByWeek(ctx, epiWeekID)
}

func (r *FacilityWeeklyMetricsRepository) ListByFacility(ctx context.Context, facilityID uuid.UUID) ([]db.ListFacilityMetricsByFacilityRow, error) {
	return r.db.ListFacilityMetricsByFacility(ctx, facilityID)
}

func (r *FacilityWeeklyMetricsRepository) DeleteByWeek(ctx context.Context, epiWeekID uuid.UUID) error {
	return r.db.DeleteFacilityMetricsByWeek(ctx, epiWeekID)
}

func (r *FacilityWeeklyMetricsRepository) WithTx(ctx context.Context, fn func(q db.Querier) error) error {
	return r.db.ExecTx(ctx, fn)
}
