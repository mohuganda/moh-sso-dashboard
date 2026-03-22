package service

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
)

type SurveillanceFacilityService struct {
	log          *logger.Logger
	facilityRepo interfaces.FacilityRepository
}

func NewSurveillanceFacilityService(
	log *logger.Logger,
	facilityRepo interfaces.FacilityRepository,
) *SurveillanceFacilityService {
	return &SurveillanceFacilityService{
		log:          log,
		facilityRepo: facilityRepo,
	}
}

func (s *SurveillanceFacilityService) ListFacilities(ctx context.Context) ([]db.ListFacilitiesRow, error) {
	s.log.Debug(ctx, "listing facilities")

	items, err := s.facilityRepo.List(ctx)
	if err != nil {
		s.log.Error(ctx, "failed to list facilities", "error", err)
		return []db.ListFacilitiesRow{}, err
	}

	return items, nil
}

func (s *SurveillanceFacilityService) ListFacilitiesByDistrict(ctx context.Context, districtID uuid.UUID) ([]db.Facility, error) {
	s.log.Debug(ctx, "listing facilities by district", "district_id", districtID)

	items, err := s.facilityRepo.ListByDistrict(ctx, districtID)
	if err != nil {
		s.log.Error(ctx, "failed to list facilities by district", "district_id", districtID, "error", err)
		return []db.Facility{}, err
	}

	return items, nil
}

func (s *SurveillanceFacilityService) GetFacilityByID(ctx context.Context, id uuid.UUID) (db.Facility, error) {
	s.log.Debug(ctx, "getting facility by id", "facility_id", id)

	item, err := s.facilityRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error(ctx, "failed to get facility by id", "facility_id", id, "error", err)
		return db.Facility{}, err
	}

	return item, nil
}

func (s *SurveillanceFacilityService) GetFacilityByExternalID(ctx context.Context, externalID string) (db.Facility, error) {
	s.log.Debug(ctx, "getting facility by external id", "external_id", externalID)

	item, err := s.facilityRepo.GetByExternalID(ctx, externalID)
	if err != nil {
		s.log.Error(ctx, "failed to get facility by external id", "external_id", externalID, "error", err)
		return db.Facility{}, err
	}

	return item, nil
}

func (s *SurveillanceFacilityService) UpsertFacilityByExternalID(ctx context.Context, arg db.UpsertFacilityByExternalIDParams) (db.Facility, error) {
	s.log.Info(ctx, "upserting facility by external id", "external_id", arg.ExternalID)

	item, err := s.facilityRepo.UpsertByExternalID(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to upsert facility by external id", "external_id", arg.ExternalID, "error", err)
		return db.Facility{}, err
	}

	return item, nil
}

func (s *SurveillanceFacilityService) UpsertFacilityByNameDistrict(ctx context.Context, arg db.UpsertFacilityByNameDistrictParams) (db.Facility, error) {
	s.log.Info(ctx, "upserting facility by name and district", "name", arg.Name)

	item, err := s.facilityRepo.UpsertByNameDistrict(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to upsert facility by name and district", "name", arg.Name, "error", err)
		return db.Facility{}, err
	}

	return item, nil
}
