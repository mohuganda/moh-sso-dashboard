package interfaces

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type NationalWeeklyStatusRepository interface {
	CreateDiseaseStatus(ctx context.Context, arg db.CreateNationalWeeklyDiseaseStatusParams) (db.NationalWeeklyStatus, error)
	UpsertDiseaseStatus(ctx context.Context, arg db.UpsertNationalWeeklyDiseaseStatusParams) (db.NationalWeeklyStatus, error)
	CreateIndicatorStatus(ctx context.Context, arg db.CreateNationalWeeklyIndicatorStatusParams) (db.NationalWeeklyStatus, error)
	UpsertIndicatorStatus(ctx context.Context, arg db.UpsertNationalWeeklyIndicatorStatusParams) (db.NationalWeeklyStatus, error)
	ListDiseaseStatusesByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListNationalWeeklyDiseaseStatusesByWeekRow, error)
	ListIndicatorStatusesByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListNationalWeeklyIndicatorStatusesByWeekRow, error)
	DeleteByWeek(ctx context.Context, epiWeekID uuid.UUID) error
}
