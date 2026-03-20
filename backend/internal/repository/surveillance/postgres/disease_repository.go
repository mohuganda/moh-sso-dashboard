package postgres

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type DiseaseRepository struct {
	db db.Store
}

func NewDiseaseRepository(db db.Store) *DiseaseRepository {
	return &DiseaseRepository{db: db}
}

func (r *DiseaseRepository) Create(ctx context.Context, arg db.CreateDiseaseParams) (db.Disease, error) {
	return r.db.CreateDisease(ctx, arg)
}

func (r *DiseaseRepository) GetByID(ctx context.Context, id uuid.UUID) (db.Disease, error) {
	return r.db.GetDiseaseByID(ctx, id)
}

func (r *DiseaseRepository) GetByName(ctx context.Context, name string) (db.Disease, error) {
	return r.db.GetDiseaseByName(ctx, name)
}

func (r *DiseaseRepository) List(ctx context.Context) ([]db.Disease, error) {
	return r.db.ListDiseases(ctx)
}

func (r *DiseaseRepository) ListActive(ctx context.Context) ([]db.Disease, error) {
	return r.db.ListActiveDiseases(ctx)
}

func (r *DiseaseRepository) Upsert(ctx context.Context, arg db.UpsertDiseaseParams) (db.Disease, error) {
	return r.db.UpsertDisease(ctx, arg)
}

func (r *DiseaseRepository) SetActiveState(ctx context.Context, arg db.SetDiseaseActiveStateParams) (db.Disease, error) {
	return r.db.SetDiseaseActiveState(ctx, arg)
}

func (r *DiseaseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.DeleteDisease(ctx, id)
}
