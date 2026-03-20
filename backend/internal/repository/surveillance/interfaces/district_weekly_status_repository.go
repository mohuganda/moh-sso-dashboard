package interfaces

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type DistrictWeeklyStatusRepository interface {
	CreateDiseaseStatus(ctx context.Context, arg db.CreateDistrictWeeklyDiseaseStatusParams) (db.DistrictWeeklyStatus, error)
	UpsertDiseaseStatus(ctx context.Context, arg db.UpsertDistrictWeeklyDiseaseStatusParams) (db.DistrictWeeklyStatus, error)
	CreateIndicatorStatus(ctx context.Context, arg db.CreateDistrictWeeklyIndicatorStatusParams) (db.DistrictWeeklyStatus, error)
	UpsertIndicatorStatus(ctx context.Context, arg db.UpsertDistrictWeeklyIndicatorStatusParams) (db.DistrictWeeklyStatus, error)
	ListDiseaseStatusesByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListDistrictWeeklyDiseaseStatusesByWeekRow, error)
	ListIndicatorStatusesByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListDistrictWeeklyIndicatorStatusesByWeekRow, error)
	ListByDistrictAndWeek(ctx context.Context, arg db.ListDistrictStatusesByDistrictAndWeekParams) ([]db.ListDistrictStatusesByDistrictAndWeekRow, error)
	DeleteByWeek(ctx context.Context, epiWeekID uuid.UUID) error
}
