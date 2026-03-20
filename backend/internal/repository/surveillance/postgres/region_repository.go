package postgres

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type RegionRepository struct {
	db db.Store
}

func NewRegionRepository(db db.Store) *RegionRepository {
	return &RegionRepository{db: db}
}

func (r *RegionRepository) Create(ctx context.Context, arg db.CreateRegionParams) (db.Region, error) {
	return r.db.CreateRegion(ctx, arg)
}

func (r *RegionRepository) GetByID(ctx context.Context, id uuid.UUID) (db.Region, error) {
	return r.db.GetRegionByID(ctx, id)
}

func (r *RegionRepository) GetByName(ctx context.Context, name string) (db.Region, error) {
	return r.db.GetRegionByName(ctx, name)
}

func (r *RegionRepository) List(ctx context.Context) ([]db.Region, error) {
	return r.db.ListRegions(ctx)
}

func (r *RegionRepository) Upsert(ctx context.Context, arg db.UpsertRegionParams) (db.Region, error) {
	return r.db.UpsertRegion(ctx, arg)
}

func (r *RegionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.DeleteRegion(ctx, id)
}
