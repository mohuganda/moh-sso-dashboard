package postgres

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type AlertRepository struct {
	db db.Store
}

func NewAlertRepository(db db.Store) *AlertRepository {
	return &AlertRepository{db: db}
}

func (r *AlertRepository) Create(ctx context.Context, arg db.CreateAlertParams) (db.Alert, error) {
	return r.db.CreateAlert(ctx, arg)
}

func (r *AlertRepository) UpsertByExternalID(ctx context.Context, arg db.UpsertAlertByExternalIDParams) (db.Alert, error) {
	return r.db.UpsertAlertByExternalID(ctx, arg)
}

func (r *AlertRepository) GetByID(ctx context.Context, id uuid.UUID) (db.Alert, error) {
	return r.db.GetAlertByID(ctx, id)
}

func (r *AlertRepository) List(ctx context.Context) ([]db.ListAlertsRow, error) {
	return r.db.ListAlerts(ctx)
}

func (r *AlertRepository) ListByDisease(ctx context.Context, diseaseID uuid.UUID) ([]db.ListAlertsByDiseaseRow, error) {
	return r.db.ListAlertsByDisease(ctx, diseaseID)
}

func (r *AlertRepository) ListByDistrict(ctx context.Context, districtID uuid.NullUUID) ([]db.ListAlertsByDistrictRow, error) {
	return r.db.ListAlertsByDistrict(ctx, districtID)
}

func (r *AlertRepository) ListByWeek(ctx context.Context, epiWeekID uuid.NullUUID) ([]db.ListAlertsByWeekRow, error) {
	return r.db.ListAlertsByWeek(ctx, epiWeekID)
}

func (r *AlertRepository) UpdateStatus(ctx context.Context, arg db.UpdateAlertStatusParams) (db.Alert, error) {
	return r.db.UpdateAlertStatus(ctx, arg)
}
