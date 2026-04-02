package postgres

import (
	"context"
	"errors"

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
	alert, err := r.db.CreateAlert(ctx, arg)
	if err != nil {
		return db.Alert{}, err
	}
	return alert, nil
}

func (r *AlertRepository) UpsertByExternalID(ctx context.Context, arg db.UpsertAlertByExternalIDParams) (db.Alert, error) {
	alert, err := r.db.UpsertAlertByExternalID(ctx, arg)
	if err != nil {
		return db.Alert{}, err
	}
	return alert, nil
}

func (r *AlertRepository) GetByID(ctx context.Context, id uuid.UUID) (db.Alert, error) {
	if id == uuid.Nil {
		return db.Alert{}, errors.New("alert id is required")
	}

	alert, err := r.db.GetAlertByID(ctx, id)
	if err != nil {
		return db.Alert{}, err
	}

	return alert, nil
}

func (r *AlertRepository) List(ctx context.Context) ([]db.ListAlertsRow, error) {
	return r.db.ListAlerts(ctx)
}

func (r *AlertRepository) ListByDisease(ctx context.Context, diseaseID uuid.UUID) ([]db.ListAlertsByDiseaseRow, error) {

	if diseaseID == uuid.Nil {
		return []db.ListAlertsByDiseaseRow{}, errors.New("disease id is required")
	}

	alerts, err := r.db.ListAlertsByDisease(ctx, diseaseID)

	if err != nil {
		return []db.ListAlertsByDiseaseRow{}, err
	}

	return alerts, nil
}

func (r *AlertRepository) ListByDistrict(
	ctx context.Context,
	districtID uuid.UUID,
) ([]db.ListAlertsByDistrictRow, error) {
	if districtID == uuid.Nil {
		return []db.ListAlertsByDistrictRow{}, errors.New("district id is required")
	}

	param := uuid.NullUUID{
		UUID:  districtID,
		Valid: true,
	}

	alerts, err := r.db.ListAlertsByDistrict(ctx, param)
	if err != nil {
		return []db.ListAlertsByDistrictRow{}, err
	}

	return alerts, nil
}

func (r *AlertRepository) ListByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListAlertsByWeekRow, error) {
	if epiWeekID == uuid.Nil {
		return []db.ListAlertsByWeekRow{}, errors.New("epi_week id is required")
	}
	param := uuid.NullUUID{
		UUID:  epiWeekID,
		Valid: true,
	}
	alerts, err := r.db.ListAlertsByWeek(ctx, param)
	if err != nil {
		return []db.ListAlertsByWeekRow{}, err
	}

	return alerts, nil
}

func (r *AlertRepository) UpdateStatus(ctx context.Context, arg db.UpdateAlertStatusParams) (db.Alert, error) {
	return r.db.UpdateAlertStatus(ctx, arg)
}

func (r *AlertRepository) WithTx(ctx context.Context, fn func(q db.Querier) error) error {
	return r.db.ExecTx(ctx, fn)
}
