package service

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"

	repository "github.com/moh-sso-dashboard/internal/repository/storage_locations"
)

type CreateStorageLocationInput db.CreateStorageLocationParams

type UpdateStorageLocationInput db.UpdateStorageLocationParams

type StorageLocationService interface {
	Create(ctx context.Context, input CreateStorageLocationInput) (db.StorageLocation, error)
	GetByID(ctx context.Context, id string) (db.StorageLocation, error)
	GetByCode(ctx context.Context, code string) (db.StorageLocation, error)
	ListActive(ctx context.Context) ([]db.StorageLocation, error)
	Update(ctx context.Context, id string, input UpdateStorageLocationInput) (db.StorageLocation, error)
	Delete(ctx context.Context, id string) error
}

type storageLocationService struct {
	repo repository.StorageLocationRepository
}

func NewStorageLocationService(repo repository.StorageLocationRepository) StorageLocationService {
	return &storageLocationService{repo: repo}
}

func (s *storageLocationService) Create(
	ctx context.Context,
	input CreateStorageLocationInput,
) (db.StorageLocation, error) {

	// Check uniqueness
	_, err := s.repo.GetByCode(ctx, input.Code)
	if err == nil {
		return db.StorageLocation{}, err
	}

	// Create
	return s.repo.Create(ctx, db.CreateStorageLocationParams{
		Code:     input.Code,
		Name:     input.Name,
		Provider: input.Provider,
		BaseUri:  input.BaseUri,
		IsActive: input.IsActive,
	})
}

func (s *storageLocationService) GetByID(
	ctx context.Context,
	id string,
) (db.StorageLocation, error) {

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return db.StorageLocation{}, err
	}

	location, err := s.repo.GetByID(ctx, parsedID)
	if err != nil {
		return db.StorageLocation{}, err
	}

	return location, nil
}

func (s *storageLocationService) GetByCode(
	ctx context.Context,
	code string,
) (db.StorageLocation, error) {

	return s.repo.GetByCode(ctx, code)
}

func (s *storageLocationService) ListActive(
	ctx context.Context,
) ([]db.StorageLocation, error) {

	return s.repo.ListActive(ctx)
}

func (s *storageLocationService) Update(
	ctx context.Context,
	id string,
	input UpdateStorageLocationInput,
) (db.StorageLocation, error) {

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return db.StorageLocation{}, err
	}

	params := db.UpdateStorageLocationParams{
		ID:       parsedID,
		Name:     input.Name,
		Provider: input.Provider,
		BaseUri:  input.BaseUri,
		IsActive: input.IsActive,
	}

	return s.repo.Update(ctx, params)
}

func (s *storageLocationService) Delete(
	ctx context.Context,
	id string,
) error {

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, parsedID)
}
