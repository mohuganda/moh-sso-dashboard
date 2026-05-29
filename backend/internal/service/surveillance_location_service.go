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

type SurveillanceLocationService struct {
	log           *logger.Logger
	regionRepo    interfaces.RegionRepository
	districtRepo  interfaces.DistrictRepository
	subCountyRepo interfaces.SubCountyRepository
	notifications NotificationsService
}

func NewSurveillanceLocationService(
	log *logger.Logger,
	regionRepo interfaces.RegionRepository,
	districtRepo interfaces.DistrictRepository,
	subCountyRepo interfaces.SubCountyRepository,
	notifications ...NotificationsService,
) *SurveillanceLocationService {
	var notificationSvc NotificationsService
	if len(notifications) > 0 {
		notificationSvc = notifications[0]
	}

	return &SurveillanceLocationService{
		log:           log,
		regionRepo:    regionRepo,
		districtRepo:  districtRepo,
		subCountyRepo: subCountyRepo,
		notifications: notificationSvc,
	}
}

// ============================================================
// Regions
// ============================================================

func (s *SurveillanceLocationService) ListRegions(ctx context.Context) ([]db.Region, error) {
	if s == nil {
		return nil, errors.New("surveillance location service is nil")
	}

	if s.regionRepo == nil {
		return nil, errors.New("region repository is nil")
	}

	if s.log != nil {
		s.log.Debug(ctx, "listing regions")
	}

	items, err := s.regionRepo.List(ctx)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to list regions", "error", err)
		}

		return []db.Region{}, fmt.Errorf("list regions: %w", err)
	}

	return items, nil
}

func (s *SurveillanceLocationService) GetRegionByID(
	ctx context.Context,
	id uuid.UUID,
) (db.Region, error) {
	if s == nil {
		return db.Region{}, errors.New("surveillance location service is nil")
	}

	if s.regionRepo == nil {
		return db.Region{}, errors.New("region repository is nil")
	}

	if id == uuid.Nil {
		return db.Region{}, errors.New("region id is required")
	}

	if s.log != nil {
		s.log.Debug(ctx, "getting region by id", "region_id", id)
	}

	item, err := s.regionRepo.GetByID(ctx, id)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to get region by id", "region_id", id, "error", err)
		}

		return db.Region{}, fmt.Errorf("get region by id: %w", err)
	}

	return item, nil
}

func (s *SurveillanceLocationService) GetRegionByName(
	ctx context.Context,
	name string,
) (db.Region, error) {
	if s == nil {
		return db.Region{}, errors.New("surveillance location service is nil")
	}

	if s.regionRepo == nil {
		return db.Region{}, errors.New("region repository is nil")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return db.Region{}, errors.New("region name is required")
	}

	if s.log != nil {
		s.log.Debug(ctx, "getting region by name", "name", name)
	}

	item, err := s.regionRepo.GetByName(ctx, name)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to get region by name", "name", name, "error", err)
		}

		return db.Region{}, fmt.Errorf("get region by name: %w", err)
	}

	return item, nil
}

func (s *SurveillanceLocationService) UpsertRegion(
	ctx context.Context,
	arg db.UpsertRegionParams,
) (db.Region, error) {
	if s == nil {
		return db.Region{}, errors.New("surveillance location service is nil")
	}

	if s.regionRepo == nil {
		return db.Region{}, errors.New("region repository is nil")
	}

	arg.Name = strings.TrimSpace(arg.Name)
	if arg.Name == "" {
		return db.Region{}, errors.New("region name is required")
	}

	if s.log != nil {
		s.log.Info(ctx, "upserting region", "name", arg.Name)
	}

	item, err := s.regionRepo.Upsert(ctx, arg)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to upsert region", "name", arg.Name, "error", err)
		}

		return db.Region{}, fmt.Errorf("upsert region: %w", err)
	}

	s.notify(ctx, model.Notification{
		Type:       "SURVEILLANCE_REGION_UPSERTED",
		Title:      "Region updated",
		Severity:   "info",
		Message:    "Surveillance region created or updated",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"region_id": item.ID.String(),
			"name":      item.Name,
		}),
	})

	return item, nil
}

// ============================================================
// Districts
// ============================================================

func (s *SurveillanceLocationService) ListDistricts(
	ctx context.Context,
) ([]db.ListDistrictsRow, error) {
	if s == nil {
		return nil, errors.New("surveillance location service is nil")
	}

	if s.districtRepo == nil {
		return nil, errors.New("district repository is nil")
	}

	if s.log != nil {
		s.log.Debug(ctx, "listing districts")
	}

	items, err := s.districtRepo.List(ctx)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to list districts", "error", err)
		}

		return []db.ListDistrictsRow{}, fmt.Errorf("list districts: %w", err)
	}

	return items, nil
}

func (s *SurveillanceLocationService) ListDistrictsByRegion(
	ctx context.Context,
	regionID uuid.UUID,
) ([]db.District, error) {
	if s == nil {
		return nil, errors.New("surveillance location service is nil")
	}

	if s.districtRepo == nil {
		return nil, errors.New("district repository is nil")
	}

	if regionID == uuid.Nil {
		return nil, errors.New("region id is required")
	}

	if s.log != nil {
		s.log.Debug(ctx, "listing districts by region", "region_id", regionID)
	}

	items, err := s.districtRepo.ListByRegion(ctx, regionID)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to list districts by region", "region_id", regionID, "error", err)
		}

		return []db.District{}, fmt.Errorf("list districts by region: %w", err)
	}

	return items, nil
}

func (s *SurveillanceLocationService) GetDistrictByID(
	ctx context.Context,
	id uuid.UUID,
) (db.District, error) {
	if s == nil {
		return db.District{}, errors.New("surveillance location service is nil")
	}

	if s.districtRepo == nil {
		return db.District{}, errors.New("district repository is nil")
	}

	if id == uuid.Nil {
		return db.District{}, errors.New("district id is required")
	}

	if s.log != nil {
		s.log.Debug(ctx, "getting district by id", "district_id", id)
	}

	item, err := s.districtRepo.GetByID(ctx, id)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to get district by id", "district_id", id, "error", err)
		}

		return db.District{}, fmt.Errorf("get district by id: %w", err)
	}

	return item, nil
}

func (s *SurveillanceLocationService) GetDistrictByName(
	ctx context.Context,
	name string,
) (db.District, error) {
	if s == nil {
		return db.District{}, errors.New("surveillance location service is nil")
	}

	if s.districtRepo == nil {
		return db.District{}, errors.New("district repository is nil")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return db.District{}, errors.New("district name is required")
	}

	if s.log != nil {
		s.log.Debug(ctx, "getting district by name", "name", name)
	}

	item, err := s.districtRepo.GetByName(ctx, name)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to get district by name", "name", name, "error", err)
		}

		return db.District{}, fmt.Errorf("get district by name: %w", err)
	}

	return item, nil
}

func (s *SurveillanceLocationService) UpsertDistrict(
	ctx context.Context,
	arg db.UpsertDistrictParams,
) (db.District, error) {
	if s == nil {
		return db.District{}, errors.New("surveillance location service is nil")
	}

	if s.districtRepo == nil {
		return db.District{}, errors.New("district repository is nil")
	}

	arg.Name = strings.TrimSpace(arg.Name)
	if arg.Name == "" {
		return db.District{}, errors.New("district name is required")
	}

	if s.log != nil {
		s.log.Info(ctx, "upserting district", "name", arg.Name, "region_id", arg.RegionID)
	}

	item, err := s.districtRepo.Upsert(ctx, arg)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to upsert district", "name", arg.Name, "error", err)
		}

		return db.District{}, fmt.Errorf("upsert district: %w", err)
	}

	s.notify(ctx, model.Notification{
		Type:       "SURVEILLANCE_DISTRICT_UPSERTED",
		Title:      "District updated",
		Severity:   "info",
		Message:    "Surveillance district created or updated",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"district_id": item.ID.String(),
			"name":        item.Name,
		}),
	})

	return item, nil
}

// ============================================================
// Sub-counties
// ============================================================

func (s *SurveillanceLocationService) ListSubcountiesByDistrict(
	ctx context.Context,
	districtID uuid.UUID,
) ([]db.SubCounty, error) {
	if s == nil {
		return nil, errors.New("surveillance location service is nil")
	}

	if s.subCountyRepo == nil {
		return nil, errors.New("subcounty repository is nil")
	}

	if districtID == uuid.Nil {
		return []db.SubCounty{}, errors.New("district id is required")
	}

	if s.log != nil {
		s.log.Debug(ctx, "listing subcounties by district", "district_id", districtID)
	}

	items, err := s.subCountyRepo.ListByDistrict(ctx, districtID)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to list subcounties by district", "district_id", districtID, "error", err)
		}

		return []db.SubCounty{}, fmt.Errorf("list subcounties by district: %w", err)
	}

	return items, nil
}

func (s *SurveillanceLocationService) GetSubcountyByID(
	ctx context.Context,
	id uuid.UUID,
) (db.SubCounty, error) {
	if s == nil {
		return db.SubCounty{}, errors.New("surveillance location service is nil")
	}

	if s.subCountyRepo == nil {
		return db.SubCounty{}, errors.New("subcounty repository is nil")
	}

	if id == uuid.Nil {
		return db.SubCounty{}, errors.New("subcounty id is required")
	}

	if s.log != nil {
		s.log.Debug(ctx, "getting subcounty by id", "id", id)
	}

	item, err := s.subCountyRepo.GetByID(ctx, id)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to get subcounty by id", "id", id, "error", err)
		}

		return db.SubCounty{}, fmt.Errorf("get subcounty by id: %w", err)
	}

	return item, nil
}

func (s *SurveillanceLocationService) UpsertSubcounty(
	ctx context.Context,
	arg db.UpsertSubCountyParams,
) (db.SubCounty, error) {
	if s == nil {
		return db.SubCounty{}, errors.New("surveillance location service is nil")
	}

	if s.subCountyRepo == nil {
		return db.SubCounty{}, errors.New("subcounty repository is nil")
	}

	arg.Name = strings.TrimSpace(arg.Name)
	if arg.Name == "" {
		return db.SubCounty{}, errors.New("subcounty name is required")
	}

	if arg.DistrictID == uuid.Nil {
		return db.SubCounty{}, errors.New("district id is required")
	}

	if s.log != nil {
		s.log.Info(ctx, "upserting subcounty", "name", arg.Name, "district_id", arg.DistrictID)
	}

	item, err := s.subCountyRepo.Upsert(ctx, arg)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to upsert subcounty", "name", arg.Name, "district_id", arg.DistrictID, "error", err)
		}

		return db.SubCounty{}, fmt.Errorf("upsert subcounty: %w", err)
	}

	s.notify(ctx, model.Notification{
		Type:       "SURVEILLANCE_SUBCOUNTY_UPSERTED",
		Title:      "Sub-county updated",
		Severity:   "info",
		Message:    "Surveillance sub-county created or updated",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"subcounty_id": item.ID.String(),
			"name":         item.Name,
			"district_id":  item.DistrictID.String(),
		}),
	})

	return item, nil
}

func (s *SurveillanceLocationService) DeleteSubcounty(
	ctx context.Context,
	id uuid.UUID,
) error {
	if s == nil {
		return errors.New("surveillance location service is nil")
	}

	if s.subCountyRepo == nil {
		return errors.New("subcounty repository is nil")
	}

	if id == uuid.Nil {
		return errors.New("subcounty id is required")
	}

	if s.log != nil {
		s.log.Info(ctx, "deleting subcounty", "id", id)
	}

	item, _ := s.subCountyRepo.GetByID(ctx, id)

	if err := s.subCountyRepo.Delete(ctx, id); err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to delete subcounty", "id", id, "error", err)
		}

		return fmt.Errorf("delete subcounty: %w", err)
	}

	if item.ID != uuid.Nil {
		s.notify(ctx, model.Notification{
			Type:       "SURVEILLANCE_SUBCOUNTY_DELETED",
			Title:      "Sub-county deleted",
			Severity:   "warning",
			Message:    "Surveillance sub-county deleted",
			TargetRole: "admin",
			Metadata: utils.MustJSON(map[string]any{
				"subcounty_id": item.ID.String(),
				"name":         item.Name,
				"district_id":  item.DistrictID.String(),
			}),
		})
	}

	return nil
}

func (s *SurveillanceLocationService) notify(
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
				"surveillance location notification failed",
				"type", notification.Type,
				"error", err,
			)
		}
	}
}
