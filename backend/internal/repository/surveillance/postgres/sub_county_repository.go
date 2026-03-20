package postgres

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type SubCountyRepository struct {
	db db.Store
}

func NewSubCountyRepository(db db.Store) *SubCountyRepository {
	return &SubCountyRepository{db: db}
}

func (r *SubCountyRepository) Create(ctx context.Context, arg db.CreateSubCountyParams) (db.SubCounty, error) {
	return r.db.CreateSubCounty(ctx, arg)
}

func (r *SubCountyRepository) GetByID(ctx context.Context, id uuid.UUID) (db.SubCounty, error) {
	return r.db.GetSubCountyByID(ctx, id)
}

func (r *SubCountyRepository) GetByNameAndDistrict(ctx context.Context, arg db.GetSubCountyByNameAndDistrictParams) (db.SubCounty, error) {
	return r.db.GetSubCountyByNameAndDistrict(ctx, arg)
}

func (r *SubCountyRepository) ListByDistrict(ctx context.Context, districtID uuid.UUID) ([]db.SubCounty, error) {
	return r.db.ListSubCountiesByDistrict(ctx, districtID)
}

func (r *SubCountyRepository) Upsert(ctx context.Context, arg db.UpsertSubCountyParams) (db.SubCounty, error) {
	return r.db.UpsertSubCounty(ctx, arg)
}

func (r *SubCountyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.DeleteSubCounty(ctx, id)
}
