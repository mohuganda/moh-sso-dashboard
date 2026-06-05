package storage_locations

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/config"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/model"
	sharedservice "github.com/moh-sso-dashboard/internal/service"
	"github.com/moh-sso-dashboard/internal/utils"
)

type CreateStorageLocationInput db.CreateStorageLocationParams

type UpdateStorageLocationInput db.UpdateStorageLocationParams

type Service interface {
	Create(ctx context.Context, input CreateStorageLocationInput) (db.StorageLocation, error)
	GetByID(ctx context.Context, id string) (db.StorageLocation, error)
	GetByCode(ctx context.Context, code string) (db.StorageLocation, error)
	ListActive(ctx context.Context) ([]db.StorageLocation, error)
	Update(ctx context.Context, id string, input UpdateStorageLocationInput) (db.StorageLocation, error)
	Delete(ctx context.Context, id string) error
}

type storageLocationService struct {
	repo          Repository
	notifications sharedservice.NotificationsService
	cfg           *config.Config
}

func NewService(
	repo Repository,
	notifications sharedservice.NotificationsService,
	cfg ...*config.Config,
) Service {
	var appConfig *config.Config
	if len(cfg) > 0 {
		appConfig = cfg[0]
	}

	return &storageLocationService{
		repo:          repo,
		notifications: notifications,
		cfg:           appConfig,
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

	// In-app only.
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

	notification := model.Notification{
		Type:       "STORAGE_LOCATION_UPDATED",
		Title:      "Storage location updated",
		Severity:   "warning",
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
	}

	s.attachAdminEmailDelivery(
		&notification,
		"storage-location-updated",
		"Storage location updated",
		fmt.Sprintf("Storage location %s was updated.", location.Name),
		map[string]any{
			"Name":      s.systemAdminName(),
			"Platform":  s.platformName(),
			"Code":      location.Code,
			"Provider":  location.Provider,
			"ActionURL": s.adminStorageURL(),
			"Details": fmt.Sprintf(
				"Storage Location ID: %s\nCode: %s\nName: %s\nProvider: %s\nBase URI: %s\nIs Active: %v\n\nPrevious Name: %s\nPrevious Provider: %s\nPrevious Base URI: %s\nPrevious Active: %v",
				location.ID.String(),
				location.Code,
				location.Name,
				location.Provider,
				location.BaseUri,
				location.IsActive,
				before.Name,
				before.Provider,
				before.BaseUri,
				before.IsActive,
			),
		},
	)

	s.notify(ctx, notification)

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
		notification := model.Notification{
			Type:       "STORAGE_LOCATION_DELETED",
			Title:      "Storage location deleted",
			Severity:   "critical",
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
		}

		s.attachAdminEmailDelivery(
			&notification,
			"storage-location-deleted",
			"Storage location deleted",
			fmt.Sprintf("Storage location %s was deleted.", location.Name),
			map[string]any{
				"Name":      s.systemAdminName(),
				"Platform":  s.platformName(),
				"Code":      location.Code,
				"Provider":  location.Provider,
				"ActionURL": s.adminStorageURL(),
				"Details": fmt.Sprintf(
					"Storage Location ID: %s\nCode: %s\nName: %s\nProvider: %s\nBase URI: %s\nWas Active: %v",
					location.ID.String(),
					location.Code,
					location.Name,
					location.Provider,
					location.BaseUri,
					location.IsActive,
				),
			},
		)

		s.notify(ctx, notification)
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

	_, _ = s.notifications.Notify(ctx, notification)
}

func (s *storageLocationService) attachAdminEmailDelivery(
	notification *model.Notification,
	templateName string,
	subject string,
	textBody string,
	templateData map[string]any,
) {
	if notification == nil {
		return
	}

	adminEmail := strings.TrimSpace(s.systemAdminEmail())
	if adminEmail == "" {
		return
	}

	if templateData == nil {
		templateData = map[string]any{}
	}

	if _, ok := templateData["Name"]; !ok {
		templateData["Name"] = s.systemAdminName()
	}

	if _, ok := templateData["Platform"]; !ok {
		templateData["Platform"] = s.platformName()
	}

	if _, ok := templateData["ActionURL"]; !ok {
		templateData["ActionURL"] = s.adminStorageURL()
	}

	notification.Deliveries = []model.NotificationDeliveryRequest{
		{
			Channel: model.NotificationChannelInApp,
			Recipient: map[string]any{
				"target_role": notification.TargetRole,
			},
			Payload: map[string]any{
				"title":    notification.Title,
				"message":  notification.Message,
				"type":     notification.Type,
				"severity": notification.Severity,
			},
			MaxAttempts: 1,
		},
		{
			Channel: model.NotificationChannelEmail,
			Recipient: map[string]any{
				"name":  s.systemAdminName(),
				"email": adminEmail,
			},
			TemplateName: templateName,
			TemplateData: templateData,
			Payload: map[string]any{
				"subject":   subject,
				"text_body": textBody,
			},
			MaxAttempts: 5,
		},
	}
}

func (s *storageLocationService) platformName() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.PlatformName) != "" {
		return strings.TrimSpace(s.cfg.Notification.PlatformName)
	}

	return "MOH Integrated Health Portal"
}

func (s *storageLocationService) systemAdminName() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.SystemAdminName) != "" {
		return strings.TrimSpace(s.cfg.Notification.SystemAdminName)
	}

	return "System Administrator"
}

func (s *storageLocationService) systemAdminEmail() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.SystemAdminEmail) != "" {
		return strings.TrimSpace(s.cfg.Notification.SystemAdminEmail)
	}

	return ""
}

func (s *storageLocationService) adminDashboardURL() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.AdminDashboardURL) != "" {
		return strings.TrimSpace(s.cfg.Notification.AdminDashboardURL)
	}

	return "http://localhost:3000/admin/home"
}

func (s *storageLocationService) adminStorageURL() string {
	base := strings.TrimRight(s.adminDashboardURL(), "/")

	if strings.HasSuffix(base, "/admin/home") {
		return strings.TrimSuffix(base, "/admin/home") + "/admin/storage-locations"
	}

	return base + "/storage-locations"
}
