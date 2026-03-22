package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
)

type SurveillanceLocationService struct {
	log           *logger.Logger
	regionRepo    interfaces.RegionRepository
	districtRepo  interfaces.DistrictRepository
	subCountyRepo interfaces.SubCountyRepository
}

func NewSurveillanceLocationService(
	log *logger.Logger,
	regionRepo interfaces.RegionRepository,
	districtRepo interfaces.DistrictRepository,
	subCountyRepo interfaces.SubCountyRepository,

) *SurveillanceLocationService {
	return &SurveillanceLocationService{
		log:           log,
		regionRepo:    regionRepo,
		districtRepo:  districtRepo,
		subCountyRepo: subCountyRepo,
	}
}

// regions
func (s *SurveillanceLocationService) ListRegions(ctx context.Context) ([]db.Region, error) {
	s.log.Debug(ctx, "listing regions")

	items, err := s.regionRepo.List(ctx)
	if err != nil {
		s.log.Error(ctx, "failed to list regions", "error", err)
		return []db.Region{}, err
	}

	return items, nil
}

func (s *SurveillanceLocationService) GetRegionByID(ctx context.Context, id uuid.UUID) (db.Region, error) {
	s.log.Debug(ctx, "getting region by id", "region_id", id)

	item, err := s.regionRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error(ctx, "failed to get region by id", "region_id", id, "error", err)
		return db.Region{}, err
	}

	return item, nil
}

func (s *SurveillanceLocationService) GetRegionByName(ctx context.Context, name string) (db.Region, error) {
	s.log.Debug(ctx, "getting region by name", "name", name)

	item, err := s.regionRepo.GetByName(ctx, name)
	if err != nil {
		s.log.Error(ctx, "failed to get region by name", "name", name, "error", err)
		return db.Region{}, err
	}

	return item, nil
}

func (s *SurveillanceLocationService) UpsertRegion(ctx context.Context, arg db.UpsertRegionParams) (db.Region, error) {
	s.log.Info(ctx, "upserting region", "name", arg.Name)

	item, err := s.regionRepo.Upsert(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to upsert region", "name", arg.Name, "error", err)
		return db.Region{}, err
	}

	return item, nil
}

// districts
func (s *SurveillanceLocationService) ListDistricts(ctx context.Context) ([]db.ListDistrictsRow, error) {
	s.log.Debug(ctx, "listing districts")

	items, err := s.districtRepo.List(ctx)
	if err != nil {
		s.log.Error(ctx, "failed to list districts", "error", err)
		return []db.ListDistrictsRow{}, err
	}

	return items, nil
}

func (s *SurveillanceLocationService) ListDistrictsByRegion(ctx context.Context, regionID uuid.UUID) ([]db.District, error) {
	s.log.Debug(ctx, "listing districts by region", "region_id", regionID)

	items, err := s.districtRepo.ListByRegion(ctx, regionID)
	if err != nil {
		s.log.Error(ctx, "failed to list districts by region", "region_id", regionID, "error", err)
		return []db.District{}, err
	}

	return items, nil
}

func (s *SurveillanceLocationService) GetDistrictByID(ctx context.Context, id uuid.UUID) (db.District, error) {
	s.log.Debug(ctx, "getting district by id", "district_id", id)

	item, err := s.districtRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error(ctx, "failed to get district by id", "district_id", id, "error", err)
		return db.District{}, err
	}

	return item, nil
}

func (s *SurveillanceLocationService) GetDistrictByName(ctx context.Context, name string) (db.District, error) {
	s.log.Debug(ctx, "getting district by name", "name", name)

	item, err := s.districtRepo.GetByName(ctx, name)
	if err != nil {
		s.log.Error(ctx, "failed to get district by name", "name", name, "error", err)
		return db.District{}, err
	}

	return item, nil
}

func (s *SurveillanceLocationService) UpsertDistrict(ctx context.Context, arg db.UpsertDistrictParams) (db.District, error) {
	s.log.Info(ctx, "upserting district", "name", arg.Name)

	item, err := s.districtRepo.Upsert(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to upsert district", "name", arg.Name, "error", err)
		return db.District{}, err
	}

	return item, nil
}

// sub-county
func (s *SurveillanceLocationService) ListSubcountiesByDistrict(ctx context.Context, districtID uuid.UUID) ([]db.SubCounty, error) {
	s.log.Debug(ctx, "listing subcounties by district", "district_id", districtID)

	if districtID == uuid.Nil {
		return []db.SubCounty{}, errors.New("district id is required")
	}

	items, err := s.subCountyRepo.ListByDistrict(ctx, districtID)
	if err != nil {
		s.log.Error(ctx, "failed to list subcounties by district", "district_id", districtID, "error", err)
		return []db.SubCounty{}, err
	}

	return items, nil
}

func (s *SurveillanceLocationService) GetSubcountyByID(ctx context.Context, id uuid.UUID) (db.SubCounty, error) {
	s.log.Debug(ctx, "getting subcounty by id", "id", id)

	if id == uuid.Nil {
		return db.SubCounty{}, errors.New("subcounty id is required")
	}

	item, err := s.subCountyRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error(ctx, "failed to get subcounty by id", "id", id, "error", err)
		return db.SubCounty{}, err
	}

	return item, nil
}

func (s *SurveillanceLocationService) UpsertSubcounty(ctx context.Context, arg db.UpsertSubCountyParams) (db.SubCounty, error) {
	s.log.Info(ctx, "upserting subcounty", "name", arg.Name, "district_id", arg.DistrictID)

	item, err := s.subCountyRepo.Upsert(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to upsert subcounty", "name", arg.Name, "district_id", arg.DistrictID, "error", err)
		return db.SubCounty{}, err
	}

	return item, nil
}

func (s *SurveillanceLocationService) DeleteSubcounty(ctx context.Context, id uuid.UUID) error {
	s.log.Info(ctx, "deleting subcounty", "id", id)

	if id == uuid.Nil {
		return errors.New("subcounty id is required")
	}

	if err := s.subCountyRepo.Delete(ctx, id); err != nil {
		s.log.Error(ctx, "failed to delete subcounty", "id", id, "error", err)
		return err
	}

	return nil
}
