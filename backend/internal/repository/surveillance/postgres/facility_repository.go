package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type FacilityRepository struct {
	db db.Store
}

func NewFacilityRepository(db db.Store) *FacilityRepository {
	return &FacilityRepository{db: db}
}

func (r *FacilityRepository) Create(ctx context.Context, arg db.CreateFacilityParams) (db.Facility, error) {
	return r.db.CreateFacility(ctx, arg)
}

func (r *FacilityRepository) GetByID(ctx context.Context, id uuid.UUID) (db.Facility, error) {
	return r.db.GetFacilityByID(ctx, id)
}

func (r *FacilityRepository) GetByExternalID(ctx context.Context, externalID string) (db.Facility, error) {

	if externalID == "" {
		return db.Facility{}, errors.New("external id is required")
	}
	param := sql.NullString{
		String: externalID,
		Valid:  true,
	}

	facility, err := r.db.GetFacilityByExternalID(ctx, param)
	if err != nil {
		return db.Facility{}, err
	}

	return facility, nil

}

func (r *FacilityRepository) GetByNameAndDistrict(ctx context.Context, arg db.GetFacilityByNameAndDistrictParams) (db.Facility, error) {
	return r.db.GetFacilityByNameAndDistrict(ctx, arg)
}

func (r *FacilityRepository) List(ctx context.Context) ([]db.ListFacilitiesRow, error) {
	return r.db.ListFacilities(ctx)
}

func (r *FacilityRepository) ListByDistrict(ctx context.Context, districtID uuid.UUID) ([]db.Facility, error) {
	if districtID == uuid.Nil {
		return []db.Facility{}, errors.New("district id is required")
	}

	param := uuid.NullUUID{
		UUID:  districtID,
		Valid: true,
	}

	facility, err := r.db.ListFacilitiesByDistrict(ctx, param)
	if err != nil {
		return []db.Facility{}, err
	}

	return facility, nil
}

func (r *FacilityRepository) UpsertByExternalID(ctx context.Context, arg db.UpsertFacilityByExternalIDParams) (db.Facility, error) {
	return r.db.UpsertFacilityByExternalID(ctx, arg)
}

func (r *FacilityRepository) UpsertByNameDistrict(ctx context.Context, arg db.UpsertFacilityByNameDistrictParams) (db.Facility, error) {
	return r.db.UpsertFacilityByNameDistrict(ctx, arg)
}

func (r *FacilityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.DeleteFacility(ctx, id)
}
