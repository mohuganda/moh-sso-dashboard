package interfaces

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type SubCountyRepository interface {
	Create(ctx context.Context, arg db.CreateSubCountyParams) (db.SubCounty, error)
	GetByID(ctx context.Context, id uuid.UUID) (db.SubCounty, error)
	GetByNameAndDistrict(ctx context.Context, arg db.GetSubCountyByNameAndDistrictParams) (db.SubCounty, error)
	ListByDistrict(ctx context.Context, districtID uuid.UUID) ([]db.SubCounty, error)
	Upsert(ctx context.Context, arg db.UpsertSubCountyParams) (db.SubCounty, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
