package interfaces

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type RegionWeeklyStatusRepository interface {
	CreateDiseaseStatus(ctx context.Context, arg db.CreateRegionWeeklyDiseaseStatusParams) (db.RegionWeeklyStatus, error)
	UpsertDiseaseStatus(ctx context.Context, arg db.UpsertRegionWeeklyDiseaseStatusParams) (db.RegionWeeklyStatus, error)

	CreateIndicatorStatus(ctx context.Context, arg db.CreateRegionWeeklyIndicatorStatusParams) (db.RegionWeeklyStatus, error)
	UpsertIndicatorStatus(ctx context.Context, arg db.UpsertRegionWeeklyIndicatorStatusParams) (db.RegionWeeklyStatus, error)

	ListDiseaseStatusesByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.RegionWeeklyStatus, error)
	ListIndicatorStatusesByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.RegionWeeklyStatus, error)

	DeleteByWeek(ctx context.Context, epiWeekID uuid.UUID) error
}
