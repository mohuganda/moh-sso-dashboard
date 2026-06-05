package surveillance

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type IndicatorRepository interface {
	Create(ctx context.Context, arg db.CreateIndicatorParams) (db.Indicator, error)
	GetByID(ctx context.Context, id uuid.UUID) (db.Indicator, error)
	GetByName(ctx context.Context, name string) (db.Indicator, error)
	List(ctx context.Context) ([]db.Indicator, error)
	ListActive(ctx context.Context) ([]db.Indicator, error)
	Upsert(ctx context.Context, arg db.UpsertIndicatorParams) (db.Indicator, error)
	SetActiveState(ctx context.Context, arg db.SetIndicatorActiveStateParams) (db.Indicator, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
