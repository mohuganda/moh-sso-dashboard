package storage_locations

import (
	"context"

	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/config"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
)

type storageLocationRepository struct {
	db     db.Store
	config *config.Config
	logger *logger.Logger
}

func NewStorageRepositoryRepository(
	cfg *config.Config,
	store db.Store,
	log logger.Logger,
) StorageLocationRepository {
	return &storageLocationRepository{
		db:     store,
		config: cfg,
		logger: &log,
	}
}

func (r *storageLocationRepository) Create(
	ctx context.Context,
	arg db.CreateStorageLocationParams,
) (db.StorageLocation, error) {

	location, err := r.db.CreateStorageLocation(ctx, arg)
	if err != nil {
		return db.StorageLocation{}, err
	}

	return location, nil
}

func (r *storageLocationRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (db.StorageLocation, error) {

	location, err := r.db.GetStorageLocationByID(ctx, id)
	if err != nil {
		return db.StorageLocation{}, err
	}

	return location, nil
}

func (r *storageLocationRepository) GetByCode(
	ctx context.Context,
	code string,
) (db.StorageLocation, error) {

	location, err := r.db.GetStorageLocationByCode(ctx, code)
	if err != nil {
		return db.StorageLocation{}, err
	}

	return location, nil
}

func (r *storageLocationRepository) ListActive(
	ctx context.Context,
) ([]db.StorageLocation, error) {

	locations, err := r.db.ListActiveStorageLocations(ctx)
	if err != nil {
		return nil, err
	}

	return locations, nil
}

func (r *storageLocationRepository) Update(
	ctx context.Context,
	arg db.UpdateStorageLocationParams,
) (db.StorageLocation, error) {

	location, err := r.db.UpdateStorageLocation(ctx, arg)
	if err != nil {
		return db.StorageLocation{}, err
	}

	return location, nil
}

func (r *storageLocationRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {

	err := r.db.DeleteStorageLocation(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
