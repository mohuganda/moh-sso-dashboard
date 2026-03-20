package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type DistrictRepository struct {
	db db.Store
}

func NewDistrictRepository(db db.Store) *DistrictRepository {
	return &DistrictRepository{db: db}
}

func (r *DistrictRepository) Create(ctx context.Context, arg db.CreateDistrictParams) (db.District, error) {
	return r.db.CreateDistrict(ctx, arg)
}

func (r *DistrictRepository) GetByID(ctx context.Context, id uuid.UUID) (db.District, error) {
	return r.db.GetDistrictByID(ctx, id)
}

func (r *DistrictRepository) GetByName(ctx context.Context, name string) (db.District, error) {
	return r.db.GetDistrictByName(ctx, name)
}

func (r *DistrictRepository) List(ctx context.Context) ([]db.ListDistrictsRow, error) {
	return r.db.ListDistricts(ctx)
}

func (r *DistrictRepository) ListByRegion(ctx context.Context, regionID uuid.UUID) ([]db.District, error) {

	if regionID == uuid.Nil {
		return []db.District{}, errors.New("district id is required")
	}

	param := uuid.NullUUID{
		UUID:  regionID,
		Valid: true,
	}

	districts, err := r.db.ListDistrictsByRegion(ctx, param)
	if err != nil {
		return []db.District{}, err
	}

	return districts, nil
}

func (r *DistrictRepository) Upsert(ctx context.Context, arg db.UpsertDistrictParams) (db.District, error) {
	return r.db.UpsertDistrict(ctx, arg)
}

func (r *DistrictRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.DeleteDistrict(ctx, id)
}
