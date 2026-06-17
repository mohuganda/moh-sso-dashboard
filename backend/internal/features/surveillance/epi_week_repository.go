package surveillance

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type EpiWeekRepository interface {
	Create(ctx context.Context, arg db.CreateEpiWeekParams) (db.EpiWeek, error)
	GetByID(ctx context.Context, id uuid.UUID) (db.EpiWeek, error)
	GetByYearWeek(ctx context.Context, arg db.GetEpiWeekByYearWeekParams) (db.EpiWeek, error)
	ListByYear(ctx context.Context, year int32) ([]db.EpiWeek, error)
	Upsert(ctx context.Context, arg db.UpsertEpiWeekParams) (db.EpiWeek, error)
}
