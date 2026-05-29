package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/model"
	repository "github.com/moh-sso-dashboard/internal/repository/storage_locations"
	"github.com/moh-sso-dashboard/internal/utils"
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
	repo          repository.StorageLocationRepository
	notifications NotificationsService
}

func NewStorageLocationService(
	repo repository.StorageLocationRepository,
	notifications ...NotificationsService,
) StorageLocationService {
	var notificationSvc NotificationsService
	if len(notifications) > 0 {
		notificationSvc = notifications[0]
	}

	return &storageLocationService{
		repo:          repo,
		notifications: notificationSvc,
	}
}

func (s *storageLocationService) Create(
	ctx context.Context,
	input CreateStorageLocationInput,
) (db.StorageLocation, error) {
	if s == nil {
		return db.StorageLocation{}, errors.New("storage location service is nil")
	}

	if s.repo == nil {
		return db.StorageLocation{}, errors.New("storage location repository is nil")
	}

	input.Code = strings.TrimSpace(input.Code)
	input.Name = strings.TrimSpace(input.Name)
	input.Provider = strings.TrimSpace(input.Provider)
	input.BaseUri = strings.TrimSpace(input.BaseUri)

	if input.Code == "" {
		return db.StorageLocation{}, errors.New("storage location code is required")
	}

	if input.Name == "" {
		return db.StorageLocation{}, errors.New("storage location name is required")
	}

	if input.Provider == "" {
		return db.StorageLocation{}, errors.New("storage location provider is required")
	}

	// Check uniqueness.
	existing, err := s.repo.GetByCode(ctx, input.Code)
	if err == nil && existing.ID != uuid.Nil {
		return db.StorageLocation{}, fmt.Errorf("storage location code already exists: %s", input.Code)
	}

	location, err := s.repo.Create(ctx, db.CreateStorageLocationParams{
		Code:     input.Code,
		Name:     input.Name,
		Provider: input.Provider,
		BaseUri:  input.BaseUri,
		IsActive: input.IsActive,
	})
	if err != nil {
		return db.StorageLocation{}, fmt.Errorf("create storage location: %w", err)
	}

	s.notify(ctx, model.Notification{
		Type:       "STORAGE_LOCATION_CREATED",
		Title:      "Storage location created",
		Severity:   "info",
		Message:    "Storage location created",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"storage_location_id": location.ID.String(),
			"code":                location.Code,
			"name":                location.Name,
			"provider":            location.Provider,
			"base_uri":            location.BaseUri,
			"is_active":           location.IsActive,
		}),
	})

	return location, nil
}

func (s *storageLocationService) GetByID(
	ctx context.Context,
	id string,
) (db.StorageLocation, error) {
	if s == nil {
		return db.StorageLocation{}, errors.New("storage location service is nil")
	}

	if s.repo == nil {
		return db.StorageLocation{}, errors.New("storage location repository is nil")
	}

	parsedID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return db.StorageLocation{}, fmt.Errorf("invalid storage location id: %w", err)
	}

	location, err := s.repo.GetByID(ctx, parsedID)
	if err != nil {
		return db.StorageLocation{}, fmt.Errorf("get storage location by id: %w", err)
	}

	return location, nil
}

func (s *storageLocationService) GetByCode(
	ctx context.Context,
	code string,
) (db.StorageLocation, error) {
	if s == nil {
		return db.StorageLocation{}, errors.New("storage location service is nil")
	}

	if s.repo == nil {
		return db.StorageLocation{}, errors.New("storage location repository is nil")
	}

	code = strings.TrimSpace(code)
	if code == "" {
		return db.StorageLocation{}, errors.New("storage location code is required")
	}

	location, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return db.StorageLocation{}, fmt.Errorf("get storage location by code: %w", err)
	}

	return location, nil
}

func (s *storageLocationService) ListActive(
	ctx context.Context,
) ([]db.StorageLocation, error) {
	if s == nil {
		return nil, errors.New("storage location service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("storage location repository is nil")
	}

	locations, err := s.repo.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active storage locations: %w", err)
	}

	return locations, nil
}

func (s *storageLocationService) Update(
	ctx context.Context,
	id string,
	input UpdateStorageLocationInput,
) (db.StorageLocation, error) {
	if s == nil {
		return db.StorageLocation{}, errors.New("storage location service is nil")
	}

	if s.repo == nil {
		return db.StorageLocation{}, errors.New("storage location repository is nil")
	}

	parsedID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return db.StorageLocation{}, fmt.Errorf("invalid storage location id: %w", err)
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Provider = strings.TrimSpace(input.Provider)
	input.BaseUri = strings.TrimSpace(input.BaseUri)

	if input.Name == "" {
		return db.StorageLocation{}, errors.New("storage location name is required")
	}

	if input.Provider == "" {
		return db.StorageLocation{}, errors.New("storage location provider is required")
	}

	before, _ := s.repo.GetByID(ctx, parsedID)

	params := db.UpdateStorageLocationParams{
		ID:       parsedID,
		Name:     input.Name,
		Provider: input.Provider,
		BaseUri:  input.BaseUri,
		IsActive: input.IsActive,
	}

	location, err := s.repo.Update(ctx, params)
	if err != nil {
		return db.StorageLocation{}, fmt.Errorf("update storage location: %w", err)
	}

	s.notify(ctx, model.Notification{
		Type:       "STORAGE_LOCATION_UPDATED",
		Title:      "Storage location updated",
		Severity:   "info",
		Message:    "Storage location updated",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"storage_location_id": location.ID.String(),
			"code":                location.Code,
			"name":                location.Name,
			"provider":            location.Provider,
			"base_uri":            location.BaseUri,
			"is_active":           location.IsActive,
			"previous": map[string]any{
				"name":      before.Name,
				"provider":  before.Provider,
				"base_uri":  before.BaseUri,
				"is_active": before.IsActive,
			},
		}),
	})

	return location, nil
}

func (s *storageLocationService) Delete(
	ctx context.Context,
	id string,
) error {
	if s == nil {
		return errors.New("storage location service is nil")
	}

	if s.repo == nil {
		return errors.New("storage location repository is nil")
	}

	parsedID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return fmt.Errorf("invalid storage location id: %w", err)
	}

	location, _ := s.repo.GetByID(ctx, parsedID)

	if err := s.repo.Delete(ctx, parsedID); err != nil {
		return fmt.Errorf("delete storage location: %w", err)
	}

	if location.ID != uuid.Nil {
		s.notify(ctx, model.Notification{
			Type:       "STORAGE_LOCATION_DELETED",
			Title:      "Storage location deleted",
			Severity:   "warning",
			Message:    "Storage location deleted",
			TargetRole: "admin",
			Metadata: utils.MustJSON(map[string]any{
				"storage_location_id": location.ID.String(),
				"code":                location.Code,
				"name":                location.Name,
				"provider":            location.Provider,
				"base_uri":            location.BaseUri,
				"is_active":           location.IsActive,
			}),
		})
	}

	return nil
}

func (s *storageLocationService) notify(
	ctx context.Context,
	notification model.Notification,
) {
	if s == nil || s.notifications == nil {
		return
	}

	if strings.TrimSpace(notification.TargetRole) == "" {
		notification.TargetRole = "admin"
	}

	if _, err := s.notifications.Notify(ctx, notification); err != nil {
		fmt.Printf("storage location notification failed type=%s error=%v\n", notification.Type, err)
	}
}
