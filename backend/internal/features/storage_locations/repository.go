package storage_locations

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type StorageLocationRepository interface {
	Create(ctx context.Context, arg db.CreateStorageLocationParams) (db.StorageLocation, error)
	GetByID(ctx context.Context, id uuid.UUID) (db.StorageLocation, error)
	GetByCode(ctx context.Context, code string) (db.StorageLocation, error)
	ListActive(ctx context.Context) ([]db.StorageLocation, error)
	Update(ctx context.Context, arg db.UpdateStorageLocationParams) (db.StorageLocation, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Repository = StorageLocationRepository
