package interfaces

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type FacilityWeeklyMetricsRepository interface {
	CreateDiseaseMetric(ctx context.Context, arg db.CreateFacilityWeeklyDiseaseMetricParams) (db.FacilityWeeklyMetric, error)
	UpsertDiseaseMetric(ctx context.Context, arg db.UpsertFacilityWeeklyDiseaseMetricParams) (db.FacilityWeeklyMetric, error)
	CreateIndicatorMetric(ctx context.Context, arg db.CreateFacilityWeeklyIndicatorMetricParams) (db.FacilityWeeklyMetric, error)
	UpsertIndicatorMetric(ctx context.Context, arg db.UpsertFacilityWeeklyIndicatorMetricParams) (db.FacilityWeeklyMetric, error)
	GetByID(ctx context.Context, id uuid.UUID) (db.FacilityWeeklyMetric, error)
	GetBySourceRecordID(ctx context.Context, sourceRecordID string) (db.FacilityWeeklyMetric, error)

	ListDiseaseMetricsByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListFacilityWeeklyDiseaseMetricsByWeekRow, error)
	ListIndicatorMetricsByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListFacilityWeeklyIndicatorMetricsByWeekRow, error)
	ListByFacility(ctx context.Context, facilityID uuid.UUID) ([]db.ListFacilityMetricsByFacilityRow, error)

	ListDiseaseTrend(ctx context.Context, arg db.ListFacilityDiseaseMetricsTrendParams) ([]db.ListFacilityDiseaseMetricsTrendRow, error)
	ListIndicatorTrend(ctx context.Context, arg db.ListFacilityIndicatorMetricsTrendParams) ([]db.ListFacilityIndicatorMetricsTrendRow, error)

	DeleteByWeek(ctx context.Context, epiWeekID uuid.UUID) error
	WithTx(ctx context.Context, fn func(q db.Querier) error) error
}
