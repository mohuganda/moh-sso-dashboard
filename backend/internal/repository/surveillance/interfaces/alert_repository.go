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
	List(ctx context.Context) ([]db.Alert, error)
	ListByDisease(ctx context.Context, diseaseID uuid.UUID) ([]db.Alert, error)
	ListByDistrict(ctx context.Context, districtID uuid.NullUUID) ([]db.Alert, error)
	ListByWeek(ctx context.Context, epiWeekID uuid.NullUUID) ([]db.Alert, error)
	UpdateStatus(ctx context.Context, arg db.UpdateAlertStatusParams) (db.Alert, error)
}
