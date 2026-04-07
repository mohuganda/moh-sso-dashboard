package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type FacilityWeeklyMetricsRepository struct {
	db db.Store
}

func NewFacilityWeeklyMetricsRepository(store db.Store) *FacilityWeeklyMetricsRepository {
	return &FacilityWeeklyMetricsRepository{
		db: store,
	}
}

func (r *FacilityWeeklyMetricsRepository) CreateDiseaseMetric(
	ctx context.Context,
	arg db.CreateFacilityWeeklyDiseaseMetricParams,
) (db.FacilityWeeklyMetric, error) {
	return r.db.CreateFacilityWeeklyDiseaseMetric(ctx, arg)
}

func (r *FacilityWeeklyMetricsRepository) UpsertDiseaseMetric(
	ctx context.Context,
	arg db.UpsertFacilityWeeklyDiseaseMetricParams,
) (db.FacilityWeeklyMetric, error) {
	return r.db.UpsertFacilityWeeklyDiseaseMetric(ctx, arg)
}

func (r *FacilityWeeklyMetricsRepository) CreateIndicatorMetric(
	ctx context.Context,
	arg db.CreateFacilityWeeklyIndicatorMetricParams,
) (db.FacilityWeeklyMetric, error) {
	return r.db.CreateFacilityWeeklyIndicatorMetric(ctx, arg)
}

func (r *FacilityWeeklyMetricsRepository) UpsertIndicatorMetric(
	ctx context.Context,
	arg db.UpsertFacilityWeeklyIndicatorMetricParams,
) (db.FacilityWeeklyMetric, error) {
	return r.db.UpsertFacilityWeeklyIndicatorMetric(ctx, arg)
}

func (r *FacilityWeeklyMetricsRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (db.FacilityWeeklyMetric, error) {
	if id == uuid.Nil {
		return db.FacilityWeeklyMetric{}, errors.New("id is required")
	}

	return r.db.GetFacilityWeeklyMetricByID(ctx, id)
}

func (r *FacilityWeeklyMetricsRepository) GetBySourceRecordID(
	ctx context.Context,
	sourceRecordID string,
) (db.FacilityWeeklyMetric, error) {
	sourceRecordID = strings.TrimSpace(sourceRecordID)
	if sourceRecordID == "" {
		return db.FacilityWeeklyMetric{}, errors.New("source record id is required")
	}

	return r.db.GetFacilityMetricBySourceRecordID(ctx, sql.NullString{
		String: sourceRecordID,
		Valid:  true,
	})
}

func (r *FacilityWeeklyMetricsRepository) ListDiseaseMetricsByWeek(
	ctx context.Context,
	epiWeekID uuid.UUID,
) ([]db.ListFacilityWeeklyDiseaseMetricsByWeekRow, error) {
	if epiWeekID == uuid.Nil {
		return []db.ListFacilityWeeklyDiseaseMetricsByWeekRow{}, errors.New("epi week id is required")
	}

	return r.db.ListFacilityWeeklyDiseaseMetricsByWeek(ctx, epiWeekID)
}

func (r *FacilityWeeklyMetricsRepository) ListIndicatorMetricsByWeek(
	ctx context.Context,
	epiWeekID uuid.UUID,
) ([]db.ListFacilityWeeklyIndicatorMetricsByWeekRow, error) {
	if epiWeekID == uuid.Nil {
		return []db.ListFacilityWeeklyIndicatorMetricsByWeekRow{}, errors.New("epi week id is required")
	}

	return r.db.ListFacilityWeeklyIndicatorMetricsByWeek(ctx, epiWeekID)
}

func (r *FacilityWeeklyMetricsRepository) ListByFacility(
	ctx context.Context,
	facilityID uuid.UUID,
) ([]db.ListFacilityMetricsByFacilityRow, error) {
	if facilityID == uuid.Nil {
		return []db.ListFacilityMetricsByFacilityRow{}, errors.New("facility id is required")
	}

	return r.db.ListFacilityMetricsByFacility(ctx, facilityID)
}

func (r *FacilityWeeklyMetricsRepository) DeleteByWeek(
	ctx context.Context,
	epiWeekID uuid.UUID,
) error {
	if epiWeekID == uuid.Nil {
		return errors.New("epi week id is required")
	}

	return r.db.DeleteFacilityMetricsByWeek(ctx, epiWeekID)
}

func (r *FacilityWeeklyMetricsRepository) WithTx(
	ctx context.Context, fn func(q db.Querier) error,
) error {
	return r.db.ExecTx(ctx, fn)
}
