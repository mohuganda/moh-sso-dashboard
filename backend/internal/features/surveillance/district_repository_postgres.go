package surveillance

import (
	"context"
	"errors"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type postgresDistrictRepository struct {
	db db.Store
}

func NewDistrictRepository(db db.Store) DistrictRepository {
	return &postgresDistrictRepository{db: db}
}

func (r *postgresDistrictRepository) Create(ctx context.Context, arg db.CreateDistrictParams) (db.District, error) {
	return r.db.CreateDistrict(ctx, arg)
}

func (r *postgresDistrictRepository) GetByID(ctx context.Context, id uuid.UUID) (db.District, error) {
	return r.db.GetDistrictByID(ctx, id)
}

func (r *postgresDistrictRepository) GetByName(ctx context.Context, name string) (db.District, error) {
	return r.db.GetDistrictByName(ctx, name)
}

func (r *postgresDistrictRepository) List(ctx context.Context) ([]db.ListDistrictsRow, error) {
	return r.db.ListDistricts(ctx)
}

func (r *postgresDistrictRepository) ListByRegion(ctx context.Context, regionID uuid.UUID) ([]db.District, error) {

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

func (r *postgresDistrictRepository) Upsert(ctx context.Context, arg db.UpsertDistrictParams) (db.District, error) {
	return r.db.UpsertDistrict(ctx, arg)
}

func (r *postgresDistrictRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.DeleteDistrict(ctx, id)
}
