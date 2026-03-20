package interfaces

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type AlertRepository interface {
	Create(ctx context.Context, arg db.CreateAlertParams) (db.Alert, error)
	UpsertByExternalID(ctx context.Context, arg db.UpsertAlertByExternalIDParams) (db.Alert, error)
	GetByID(ctx context.Context, id uuid.UUID) (db.Alert, error)
	List(ctx context.Context) ([]db.ListAlertsRow, error)
	ListByDisease(ctx context.Context, diseaseID uuid.UUID) ([]db.ListAlertsByDiseaseRow, error)
	ListByDistrict(ctx context.Context, districtID uuid.UUID) ([]db.ListAlertsByDistrictRow, error)
	ListByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListAlertsByWeekRow, error)
	UpdateStatus(ctx context.Context, arg db.UpdateAlertStatusParams) (db.Alert, error)
}
