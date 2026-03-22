package service

import (
	"context"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
)

type SurveillanceLookupService struct {
	log        *logger.Logger
	lookupRepo interfaces.LookupRepository
}

func NewSurveillanceLookupService(
	log *logger.Logger,
	lookupRepo interfaces.LookupRepository,
) *SurveillanceLookupService {
	return &SurveillanceLookupService{
		log:        log,
		lookupRepo: lookupRepo,
	}
}

func (s *SurveillanceLookupService) ResolveDiseaseByName(ctx context.Context, name string) (db.Disease, error) {
	return s.lookupRepo.ResolveDiseaseByName(ctx, name)
}

func (s *SurveillanceLookupService) ResolveIndicatorByName(ctx context.Context, name string) (db.Indicator, error) {
	return s.lookupRepo.ResolveIndicatorByName(ctx, name)
}

func (s *SurveillanceLookupService) ResolveWeekByYearWeek(ctx context.Context, arg db.ResolveWeekByYearWeekParams) (db.EpiWeek, error) {
	return s.lookupRepo.ResolveWeekByYearWeek(ctx, arg)
}

func (s *SurveillanceLookupService) ResolveDistrictByName(ctx context.Context, name string) (db.District, error) {
	return s.lookupRepo.ResolveDistrictByName(ctx, name)
}

func (s *SurveillanceLookupService) ResolveRegionByName(ctx context.Context, name string) (db.Region, error) {
	return s.lookupRepo.ResolveRegionByName(ctx, name)
}

func (s *SurveillanceLookupService) ResolveSubCountyByNameAndDistrict(ctx context.Context, arg db.ResolveSubCountyByNameAndDistrictParams) (db.SubCounty, error) {
	return s.lookupRepo.ResolveSubCountyByNameAndDistrict(ctx, arg)
}

func (s *SurveillanceLookupService) ResolveFacilityByExternalID(ctx context.Context, externalID string) (db.Facility, error) {
	return s.lookupRepo.ResolveFacilityByExternalID(ctx, externalID)
}

func (s *SurveillanceLookupService) ResolveFacilityByNameAndDistrict(ctx context.Context, arg db.ResolveFacilityByNameAndDistrictParams) (db.Facility, error) {
	return s.lookupRepo.ResolveFacilityByNameAndDistrict(ctx, arg)
}
