package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
	"github.com/moh-sso-dashboard/internal/utils"
)

type SurveillanceFacilityService struct {
	log           *logger.Logger
	facilityRepo  interfaces.FacilityRepository
	notifications NotificationsService
}

func NewSurveillanceFacilityService(
	log *logger.Logger,
	facilityRepo interfaces.FacilityRepository,
	notifications ...NotificationsService,
) *SurveillanceFacilityService {
	var notificationSvc NotificationsService
	if len(notifications) > 0 {
		notificationSvc = notifications[0]
	}

	return &SurveillanceFacilityService{
		log:           log,
		facilityRepo:  facilityRepo,
		notifications: notificationSvc,
	}
}

func (s *SurveillanceFacilityService) ListFacilities(
	ctx context.Context,
) ([]db.ListFacilitiesRow, error) {
	if s == nil {
		return nil, errors.New("surveillance facility service is nil")
	}

	if s.facilityRepo == nil {
		return nil, errors.New("facility repository is nil")
	}

	if s.log != nil {
		s.log.Debug(ctx, "listing facilities")
	}

	items, err := s.facilityRepo.List(ctx)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to list facilities", "error", err)
		}

		return []db.ListFacilitiesRow{}, fmt.Errorf("list facilities: %w", err)
	}

	return items, nil
}

func (s *SurveillanceFacilityService) ListFacilitiesByDistrict(
	ctx context.Context,
	districtID uuid.UUID,
) ([]db.Facility, error) {
	if s == nil {
		return nil, errors.New("surveillance facility service is nil")
	}

	if s.facilityRepo == nil {
		return nil, errors.New("facility repository is nil")
	}

	if districtID == uuid.Nil {
		return nil, errors.New("district id is required")
	}

	if s.log != nil {
		s.log.Debug(ctx, "listing facilities by district", "district_id", districtID)
	}

	items, err := s.facilityRepo.ListByDistrict(ctx, districtID)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to list facilities by district", "district_id", districtID, "error", err)
		}

		return []db.Facility{}, fmt.Errorf("list facilities by district: %w", err)
	}

	return items, nil
}

func (s *SurveillanceFacilityService) GetFacilityByID(
	ctx context.Context,
	id uuid.UUID,
) (db.Facility, error) {
	if s == nil {
		return db.Facility{}, errors.New("surveillance facility service is nil")
	}

	if s.facilityRepo == nil {
		return db.Facility{}, errors.New("facility repository is nil")
	}

	if id == uuid.Nil {
		return db.Facility{}, errors.New("facility id is required")
	}

	if s.log != nil {
		s.log.Debug(ctx, "getting facility by id", "facility_id", id)
	}

	item, err := s.facilityRepo.GetByID(ctx, id)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to get facility by id", "facility_id", id, "error", err)
		}

		return db.Facility{}, fmt.Errorf("get facility by id: %w", err)
	}

	return item, nil
}

func (s *SurveillanceFacilityService) GetFacilityByExternalID(
	ctx context.Context,
	externalID string,
) (db.Facility, error) {
	if s == nil {
		return db.Facility{}, errors.New("surveillance facility service is nil")
	}

	if s.facilityRepo == nil {
		return db.Facility{}, errors.New("facility repository is nil")
	}

	externalID = strings.TrimSpace(externalID)
	if externalID == "" {
		return db.Facility{}, errors.New("facility external id is required")
	}

	if s.log != nil {
		s.log.Debug(ctx, "getting facility by external id", "external_id", externalID)
	}

	item, err := s.facilityRepo.GetByExternalID(ctx, externalID)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to get facility by external id", "external_id", externalID, "error", err)
		}

		return db.Facility{}, fmt.Errorf("get facility by external id: %w", err)
	}

	return item, nil
}

func (s *SurveillanceFacilityService) UpsertFacilityByExternalID(
	ctx context.Context,
	arg db.UpsertFacilityByExternalIDParams,
) (db.Facility, error) {
	if s == nil {
		return db.Facility{}, errors.New("surveillance facility service is nil")
	}

	if s.facilityRepo == nil {
		return db.Facility{}, errors.New("facility repository is nil")
	}

	arg.Name = strings.TrimSpace(arg.Name)

	if arg.Name == "" {
		return db.Facility{}, errors.New("facility name is required")
	}

	if s.log != nil {
		s.log.Info(ctx, "upserting facility by external id", "external_id", arg.ExternalID, "name", arg.Name)
	}

	item, err := s.facilityRepo.UpsertByExternalID(ctx, arg)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to upsert facility by external id", "external_id", arg.ExternalID, "error", err)
		}

		return db.Facility{}, fmt.Errorf("upsert facility by external id: %w", err)
	}

	s.notify(ctx, model.Notification{
		Type:       "SURVEILLANCE_FACILITY_UPSERTED",
		Title:      "Facility updated",
		Severity:   "info",
		Message:    "Surveillance facility created or updated",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"facility_id": item.ID.String(),
			"name":        item.Name,
			"external_id": nullStringValue(item.ExternalID),
			"region_id":   nullUUIDString(item.RegionID),
			"sub_county_id": nullUUIDString(
				item.SubCountyID,
			),
		}),
	})

	return item, nil
}

func (s *SurveillanceFacilityService) UpsertFacilityByNameDistrict(
	ctx context.Context,
	arg db.UpsertFacilityByNameDistrictParams,
) (db.Facility, error) {
	if s == nil {
		return db.Facility{}, errors.New("surveillance facility service is nil")
	}

	if s.facilityRepo == nil {
		return db.Facility{}, errors.New("facility repository is nil")
	}

	arg.Name = strings.TrimSpace(arg.Name)

	if arg.Name == "" {
		return db.Facility{}, errors.New("facility name is required")
	}

	if s.log != nil {
		s.log.Info(ctx, "upserting facility by name and district", "name", arg.Name, "district_id", arg.DistrictID)
	}

	item, err := s.facilityRepo.UpsertByNameDistrict(ctx, arg)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to upsert facility by name and district", "name", arg.Name, "error", err)
		}

		return db.Facility{}, fmt.Errorf("upsert facility by name and district: %w", err)
	}

	s.notify(ctx, model.Notification{
		Type:       "SURVEILLANCE_FACILITY_UPSERTED",
		Title:      "Facility updated",
		Severity:   "info",
		Message:    "Surveillance facility created or updated",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"facility_id": item.ID.String(),
			"name":        item.Name,
			"external_id": nullStringValue(item.ExternalID),
			"district_id": item.DistrictID,
			"region_id":   nullUUIDString(item.RegionID),
			"sub_county_id": nullUUIDString(
				item.SubCountyID,
			),
		}),
	})

	return item, nil
}

func (s *SurveillanceFacilityService) notify(
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
		if s.log != nil {
			s.log.Error(
				ctx,
				"surveillance facility notification failed",
				"type", notification.Type,
				"error", err,
			)
		}
	}
}

func nullUUIDString(value uuid.NullUUID) string {
	if !value.Valid {
		return ""
	}

	return value.UUID.String()
}
