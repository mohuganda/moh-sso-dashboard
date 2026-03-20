package service

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
)

type SurveillanceService struct {
	log *logger.Logger

	diseaseRepo        interfaces.DiseaseRepository
	indicatorRepo      interfaces.IndicatorRepository
	regionRepo         interfaces.RegionRepository
	districtRepo       interfaces.DistrictRepository
	subCountyRepo      interfaces.SubCountyRepository
	facilityRepo       interfaces.FacilityRepository
	epiWeekRepo        interfaces.EpiWeekRepository
	metricsRepo        interfaces.FacilityWeeklyMetricsRepository
	districtStatusRepo interfaces.DistrictWeeklyStatusRepository
	regionStatusRepo   interfaces.RegionWeeklyStatusRepository
	nationalStatusRepo interfaces.NationalWeeklyStatusRepository
	dashboardRepo      interfaces.DashboardRepository
	alertRepo          interfaces.AlertRepository
	lookupRepo         interfaces.LookupRepository
}

func NewSurveillanceService(
	log *logger.Logger,
	diseaseRepo interfaces.DiseaseRepository,
	indicatorRepo interfaces.IndicatorRepository,
	regionRepo interfaces.RegionRepository,
	districtRepo interfaces.DistrictRepository,
	subCountyRepo interfaces.SubCountyRepository,
	facilityRepo interfaces.FacilityRepository,
	epiWeekRepo interfaces.EpiWeekRepository,
	metricsRepo interfaces.FacilityWeeklyMetricsRepository,
	districtStatusRepo interfaces.DistrictWeeklyStatusRepository,
	regionStatusRepo interfaces.RegionWeeklyStatusRepository,
	nationalStatusRepo interfaces.NationalWeeklyStatusRepository,
	dashboardRepo interfaces.DashboardRepository,
	alertRepo interfaces.AlertRepository,
	lookupRepo interfaces.LookupRepository,
) *SurveillanceService {
	return &SurveillanceService{
		log:                log,
		diseaseRepo:        diseaseRepo,
		indicatorRepo:      indicatorRepo,
		regionRepo:         regionRepo,
		districtRepo:       districtRepo,
		subCountyRepo:      subCountyRepo,
		facilityRepo:       facilityRepo,
		epiWeekRepo:        epiWeekRepo,
		metricsRepo:        metricsRepo,
		districtStatusRepo: districtStatusRepo,
		regionStatusRepo:   regionStatusRepo,
		nationalStatusRepo: nationalStatusRepo,
		dashboardRepo:      dashboardRepo,
		alertRepo:          alertRepo,
		lookupRepo:         lookupRepo,
	}
}

func (s *SurveillanceService) ListDiseases(ctx context.Context) ([]db.Disease, error) {
	s.log.Debug(ctx, "listing diseases")

	items, err := s.diseaseRepo.List(ctx)
	if err != nil {
		s.log.Error(ctx, "failed to list diseases", "error", err)
		return []db.Disease{}, err
	}

	return items, nil
}

func (s *SurveillanceService) ListActiveDiseases(ctx context.Context) ([]db.Disease, error) {
	s.log.Debug(ctx, "listing active diseases")

	items, err := s.diseaseRepo.ListActive(ctx)
	if err != nil {
		s.log.Error(ctx, "failed to list active diseases", "error", err)
		return []db.Disease{}, err
	}

	return items, nil
}

func (s *SurveillanceService) GetDiseaseByID(ctx context.Context, id uuid.UUID) (db.Disease, error) {
	s.log.Debug(ctx, "getting disease by id", "disease_id", id)

	item, err := s.diseaseRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error(ctx, "failed to get disease by id", "disease_id", id, "error", err)
		return db.Disease{}, err
	}

	return item, nil
}

func (s *SurveillanceService) GetDiseaseByName(ctx context.Context, name string) (db.Disease, error) {
	s.log.Debug(ctx, "getting disease by name", "name", name)

	item, err := s.diseaseRepo.GetByName(ctx, name)
	if err != nil {
		s.log.Error(ctx, "failed to get disease by name", "name", name, "error", err)
		return db.Disease{}, err
	}

	return item, nil
}

func (s *SurveillanceService) CreateDisease(ctx context.Context, arg db.CreateDiseaseParams) (db.Disease, error) {
	s.log.Info(ctx, "creating disease", "name", arg.Name)

	item, err := s.diseaseRepo.Create(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to create disease", "name", arg.Name, "error", err)
		return db.Disease{}, err
	}

	return item, nil
}

func (s *SurveillanceService) UpsertDisease(ctx context.Context, arg db.UpsertDiseaseParams) (db.Disease, error) {
	s.log.Info(ctx, "upserting disease", "name", arg.Name)

	item, err := s.diseaseRepo.Upsert(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to upsert disease", "name", arg.Name, "error", err)
		return db.Disease{}, err
	}

	return item, nil
}

func (s *SurveillanceService) SetDiseaseActiveState(ctx context.Context, arg db.SetDiseaseActiveStateParams) (db.Disease, error) {
	s.log.Info(ctx, "setting disease active state", "disease_id", arg.ID)

	item, err := s.diseaseRepo.SetActiveState(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to set disease active state", "disease_id", arg.ID, "error", err)
		return db.Disease{}, err
	}

	return item, nil
}

func (s *SurveillanceService) DeleteDisease(ctx context.Context, id uuid.UUID) error {
	s.log.Info(ctx, "deleting disease", "disease_id", id)

	if err := s.diseaseRepo.Delete(ctx, id); err != nil {
		s.log.Error(ctx, "failed to delete disease", "disease_id", id, "error", err)
		return err
	}

	return nil
}

func (s *SurveillanceService) ListIndicators(ctx context.Context) ([]db.Indicator, error) {
	s.log.Debug(ctx, "listing indicators")

	items, err := s.indicatorRepo.List(ctx)
	if err != nil {
		s.log.Error(ctx, "failed to list indicators", "error", err)
		return []db.Indicator{}, err
	}

	return items, nil
}

func (s *SurveillanceService) ListActiveIndicators(ctx context.Context) ([]db.Indicator, error) {
	s.log.Debug(ctx, "listing active indicators")

	items, err := s.indicatorRepo.ListActive(ctx)
	if err != nil {
		s.log.Error(ctx, "failed to list active indicators", "error", err)
		return []db.Indicator{}, err
	}

	return items, nil
}

func (s *SurveillanceService) GetIndicatorByID(ctx context.Context, id uuid.UUID) (db.Indicator, error) {
	s.log.Debug(ctx, "getting indicator by id", "indicator_id", id)

	item, err := s.indicatorRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error(ctx, "failed to get indicator by id", "indicator_id", id, "error", err)
		return db.Indicator{}, err
	}

	return item, nil
}

func (s *SurveillanceService) GetIndicatorByName(ctx context.Context, name string) (db.Indicator, error) {
	s.log.Debug(ctx, "getting indicator by name", "name", name)

	item, err := s.indicatorRepo.GetByName(ctx, name)
	if err != nil {
		s.log.Error(ctx, "failed to get indicator by name", "name", name, "error", err)
		return db.Indicator{}, err
	}

	return item, nil
}

func (s *SurveillanceService) CreateIndicator(ctx context.Context, arg db.CreateIndicatorParams) (db.Indicator, error) {
	s.log.Info(ctx, "creating indicator", "name", arg.Name)

	item, err := s.indicatorRepo.Create(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to create indicator", "name", arg.Name, "error", err)
		return db.Indicator{}, err
	}

	return item, nil
}

func (s *SurveillanceService) UpsertIndicator(ctx context.Context, arg db.UpsertIndicatorParams) (db.Indicator, error) {
	s.log.Info(ctx, "upserting indicator", "name", arg.Name)

	item, err := s.indicatorRepo.Upsert(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to upsert indicator", "name", arg.Name, "error", err)
		return db.Indicator{}, err
	}

	return item, nil
}

func (s *SurveillanceService) SetIndicatorActiveState(ctx context.Context, arg db.SetIndicatorActiveStateParams) (db.Indicator, error) {
	s.log.Info(ctx, "setting indicator active state", "indicator_id", arg.ID)

	item, err := s.indicatorRepo.SetActiveState(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to set indicator active state", "indicator_id", arg.ID, "error", err)
		return db.Indicator{}, err
	}

	return item, nil
}

func (s *SurveillanceService) DeleteIndicator(ctx context.Context, id uuid.UUID) error {
	s.log.Info(ctx, "deleting indicator", "indicator_id", id)

	if err := s.indicatorRepo.Delete(ctx, id); err != nil {
		s.log.Error(ctx, "failed to delete indicator", "indicator_id", id, "error", err)
		return err
	}

	return nil
}

func (s *SurveillanceService) ListRegions(ctx context.Context) ([]db.Region, error) {
	s.log.Debug(ctx, "listing regions")

	items, err := s.regionRepo.List(ctx)
	if err != nil {
		s.log.Error(ctx, "failed to list regions", "error", err)
		return []db.Region{}, err
	}

	return items, nil
}

func (s *SurveillanceService) GetRegionByID(ctx context.Context, id uuid.UUID) (db.Region, error) {
	s.log.Debug(ctx, "getting region by id", "region_id", id)

	item, err := s.regionRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error(ctx, "failed to get region by id", "region_id", id, "error", err)
		return db.Region{}, err
	}

	return item, nil
}

func (s *SurveillanceService) GetRegionByName(ctx context.Context, name string) (db.Region, error) {
	s.log.Debug(ctx, "getting region by name", "name", name)

	item, err := s.regionRepo.GetByName(ctx, name)
	if err != nil {
		s.log.Error(ctx, "failed to get region by name", "name", name, "error", err)
		return db.Region{}, err
	}

	return item, nil
}

func (s *SurveillanceService) UpsertRegion(ctx context.Context, arg db.UpsertRegionParams) (db.Region, error) {
	s.log.Info(ctx, "upserting region", "name", arg.Name)

	item, err := s.regionRepo.Upsert(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to upsert region", "name", arg.Name, "error", err)
		return db.Region{}, err
	}

	return item, nil
}

func (s *SurveillanceService) ListDistricts(ctx context.Context) ([]db.ListDistrictsRow, error) {
	s.log.Debug(ctx, "listing districts")

	items, err := s.districtRepo.List(ctx)
	if err != nil {
		s.log.Error(ctx, "failed to list districts", "error", err)
		return []db.ListDistrictsRow{}, err
	}

	return items, nil
}

func (s *SurveillanceService) ListDistrictsByRegion(ctx context.Context, regionID uuid.UUID) ([]db.District, error) {
	s.log.Debug(ctx, "listing districts by region", "region_id", regionID)

	items, err := s.districtRepo.ListByRegion(ctx, regionID)
	if err != nil {
		s.log.Error(ctx, "failed to list districts by region", "region_id", regionID, "error", err)
		return []db.District{}, err
	}

	return items, nil
}

func (s *SurveillanceService) GetDistrictByID(ctx context.Context, id uuid.UUID) (db.District, error) {
	s.log.Debug(ctx, "getting district by id", "district_id", id)

	item, err := s.districtRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error(ctx, "failed to get district by id", "district_id", id, "error", err)
		return db.District{}, err
	}

	return item, nil
}

func (s *SurveillanceService) GetDistrictByName(ctx context.Context, name string) (db.District, error) {
	s.log.Debug(ctx, "getting district by name", "name", name)

	item, err := s.districtRepo.GetByName(ctx, name)
	if err != nil {
		s.log.Error(ctx, "failed to get district by name", "name", name, "error", err)
		return db.District{}, err
	}

	return item, nil
}

func (s *SurveillanceService) UpsertDistrict(ctx context.Context, arg db.UpsertDistrictParams) (db.District, error) {
	s.log.Info(ctx, "upserting district", "name", arg.Name)

	item, err := s.districtRepo.Upsert(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to upsert district", "name", arg.Name, "error", err)
		return db.District{}, err
	}

	return item, nil
}

func (s *SurveillanceService) ListFacilities(ctx context.Context) ([]db.ListFacilitiesRow, error) {
	s.log.Debug(ctx, "listing facilities")

	items, err := s.facilityRepo.List(ctx)
	if err != nil {
		s.log.Error(ctx, "failed to list facilities", "error", err)
		return []db.ListFacilitiesRow{}, err
	}

	return items, nil
}

func (s *SurveillanceService) ListFacilitiesByDistrict(ctx context.Context, districtID uuid.UUID) ([]db.Facility, error) {
	s.log.Debug(ctx, "listing facilities by district", "district_id", districtID)

	items, err := s.facilityRepo.ListByDistrict(ctx, districtID)
	if err != nil {
		s.log.Error(ctx, "failed to list facilities by district", "district_id", districtID, "error", err)
		return []db.Facility{}, err
	}

	return items, nil
}

func (s *SurveillanceService) GetFacilityByID(ctx context.Context, id uuid.UUID) (db.Facility, error) {
	s.log.Debug(ctx, "getting facility by id", "facility_id", id)

	item, err := s.facilityRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error(ctx, "failed to get facility by id", "facility_id", id, "error", err)
		return db.Facility{}, err
	}

	return item, nil
}

func (s *SurveillanceService) GetFacilityByExternalID(ctx context.Context, externalID string) (db.Facility, error) {
	s.log.Debug(ctx, "getting facility by external id", "external_id", externalID)

	item, err := s.facilityRepo.GetByExternalID(ctx, externalID)
	if err != nil {
		s.log.Error(ctx, "failed to get facility by external id", "external_id", externalID, "error", err)
		return db.Facility{}, err
	}

	return item, nil
}

func (s *SurveillanceService) UpsertFacilityByExternalID(ctx context.Context, arg db.UpsertFacilityByExternalIDParams) (db.Facility, error) {
	s.log.Info(ctx, "upserting facility by external id", "external_id", arg.ExternalID)

	item, err := s.facilityRepo.UpsertByExternalID(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to upsert facility by external id", "external_id", arg.ExternalID, "error", err)
		return db.Facility{}, err
	}

	return item, nil
}

func (s *SurveillanceService) UpsertFacilityByNameDistrict(ctx context.Context, arg db.UpsertFacilityByNameDistrictParams) (db.Facility, error) {
	s.log.Info(ctx, "upserting facility by name and district", "name", arg.Name)

	item, err := s.facilityRepo.UpsertByNameDistrict(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to upsert facility by name and district", "name", arg.Name, "error", err)
		return db.Facility{}, err
	}

	return item, nil
}

func (s *SurveillanceService) GetEpiWeekByID(ctx context.Context, id uuid.UUID) (db.EpiWeek, error) {
	s.log.Debug(ctx, "getting epi week by id", "week_id", id)

	item, err := s.epiWeekRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error(ctx, "failed to get epi week by id", "week_id", id, "error", err)
		return db.EpiWeek{}, err
	}

	return item, nil
}

func (s *SurveillanceService) GetEpiWeekByYearWeek(ctx context.Context, arg db.GetEpiWeekByYearWeekParams) (db.EpiWeek, error) {
	s.log.Debug(ctx, "getting epi week by year/week", "year", arg.EpiYear, "week", arg.EpiWeek)

	item, err := s.epiWeekRepo.GetByYearWeek(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to get epi week by year/week", "year", arg.EpiYear, "week", arg.EpiWeek, "error", err)
		return db.EpiWeek{}, err
	}

	return item, nil
}

func (s *SurveillanceService) ListEpiWeeksByYear(ctx context.Context, year int32) ([]db.EpiWeek, error) {
	s.log.Debug(ctx, "listing epi weeks by year", "year", year)

	items, err := s.epiWeekRepo.ListByYear(ctx, year)
	if err != nil {
		s.log.Error(ctx, "failed to list epi weeks by year", "year", year, "error", err)
		return []db.EpiWeek{}, err
	}

	return items, nil
}

func (s *SurveillanceService) UpsertEpiWeek(ctx context.Context, arg db.UpsertEpiWeekParams) (db.EpiWeek, error) {
	s.log.Info(ctx, "upserting epi week", "year", arg.EpiYear, "week", arg.EpiWeek)

	item, err := s.epiWeekRepo.Upsert(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to upsert epi week", "year", arg.EpiYear, "week", arg.EpiWeek, "error", err)
		return db.EpiWeek{}, err
	}

	return item, nil
}

func (s *SurveillanceService) UpsertFacilityWeeklyDiseaseMetric(ctx context.Context, arg db.UpsertFacilityWeeklyDiseaseMetricParams) (db.FacilityWeeklyMetric, error) {
	s.log.Debug(ctx, "upserting facility weekly disease metric")

	item, err := s.metricsRepo.UpsertDiseaseMetric(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to upsert facility weekly disease metric", "error", err)
		return db.FacilityWeeklyMetric{}, err
	}

	return item, nil
}

func (s *SurveillanceService) UpsertFacilityWeeklyIndicatorMetric(ctx context.Context, arg db.UpsertFacilityWeeklyIndicatorMetricParams) (db.FacilityWeeklyMetric, error) {
	s.log.Debug(ctx, "upserting facility weekly indicator metric")

	item, err := s.metricsRepo.UpsertIndicatorMetric(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to upsert facility weekly indicator metric", "error", err)
		return db.FacilityWeeklyMetric{}, err
	}

	return item, nil
}

func (s *SurveillanceService) GetFacilityMetricBySourceRecordID(ctx context.Context, sourceRecordID string) (db.FacilityWeeklyMetric, error) {
	s.log.Debug(ctx, "getting facility metric by source record id", "source_record_id", sourceRecordID)

	item, err := s.metricsRepo.GetBySourceRecordID(ctx, sourceRecordID)
	if err != nil {
		s.log.Error(ctx, "failed to get facility metric by source record id", "source_record_id", sourceRecordID, "error", err)
		return db.FacilityWeeklyMetric{}, err
	}

	return item, nil
}

func (s *SurveillanceService) ListFacilityWeeklyDiseaseMetricsByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListFacilityWeeklyDiseaseMetricsByWeekRow, error) {
	s.log.Debug(ctx, "listing facility weekly disease metrics by week")

	items, err := s.metricsRepo.ListDiseaseMetricsByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to list facility weekly disease metrics by week", "error", err)
		return []db.ListFacilityWeeklyDiseaseMetricsByWeekRow{}, err
	}

	return items, nil
}

func (s *SurveillanceService) ListFacilityWeeklyIndicatorMetricsByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListFacilityWeeklyIndicatorMetricsByWeekRow, error) {
	s.log.Debug(ctx, "listing facility weekly indicator metrics by week")

	items, err := s.metricsRepo.ListIndicatorMetricsByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to list facility weekly indicator metrics by week", "error", err)
		return []db.ListFacilityWeeklyIndicatorMetricsByWeekRow{}, err
	}

	return items, nil
}

func (s *SurveillanceService) ListFacilityMetricsByFacility(ctx context.Context, facilityID uuid.UUID) ([]db.ListFacilityMetricsByFacilityRow, error) {
	s.log.Debug(ctx, "listing facility metrics by facility", "facility_id", facilityID)

	items, err := s.metricsRepo.ListByFacility(ctx, facilityID)
	if err != nil {
		s.log.Error(ctx, "failed to list facility metrics by facility", "facility_id", facilityID, "error", err)
		return []db.ListFacilityMetricsByFacilityRow{}, err
	}

	return items, nil
}

func (s *SurveillanceService) ListDistrictWeeklyDiseaseStatusesByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListDistrictWeeklyDiseaseStatusesByWeekRow, error) {
	s.log.Debug(ctx, "listing district weekly disease statuses by week")

	items, err := s.districtStatusRepo.ListDiseaseStatusesByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to list district weekly disease statuses by week", "error", err)
		return []db.ListDistrictWeeklyDiseaseStatusesByWeekRow{}, err
	}

	return items, nil
}

func (s *SurveillanceService) ListDistrictWeeklyIndicatorStatusesByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListDistrictWeeklyIndicatorStatusesByWeekRow, error) {
	s.log.Debug(ctx, "listing district weekly indicator statuses by week")

	items, err := s.districtStatusRepo.ListIndicatorStatusesByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to list district weekly indicator statuses by week", "error", err)
		return []db.ListDistrictWeeklyIndicatorStatusesByWeekRow{}, err
	}

	return items, nil
}

func (s *SurveillanceService) ListRegionWeeklyDiseaseStatusesByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListRegionWeeklyDiseaseStatusesByWeekRow, error) {
	s.log.Debug(ctx, "listing region weekly disease statuses by week")

	items, err := s.regionStatusRepo.ListDiseaseStatusesByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to list region weekly disease statuses by week", "error", err)
		return []db.ListRegionWeeklyDiseaseStatusesByWeekRow{}, err
	}

	return items, nil
}

func (s *SurveillanceService) ListRegionWeeklyIndicatorStatusesByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListRegionWeeklyIndicatorStatusesByWeekRow, error) {
	s.log.Debug(ctx, "listing region weekly indicator statuses by week")

	items, err := s.regionStatusRepo.ListIndicatorStatusesByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to list region weekly indicator statuses by week", "error", err)
		return []db.ListRegionWeeklyIndicatorStatusesByWeekRow{}, err
	}

	return items, nil
}

func (s *SurveillanceService) ListNationalWeeklyDiseaseStatusesByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListNationalWeeklyDiseaseStatusesByWeekRow, error) {
	s.log.Debug(ctx, "listing national weekly disease statuses by week")

	items, err := s.nationalStatusRepo.ListDiseaseStatusesByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to list national weekly disease statuses by week", "error", err)
		return []db.ListNationalWeeklyDiseaseStatusesByWeekRow{}, err
	}

	return items, nil
}

func (s *SurveillanceService) ListNationalWeeklyIndicatorStatusesByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListNationalWeeklyIndicatorStatusesByWeekRow, error) {
	s.log.Debug(ctx, "listing national weekly indicator statuses by week")

	items, err := s.nationalStatusRepo.ListIndicatorStatusesByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to list national weekly indicator statuses by week", "error", err)
		return []db.ListNationalWeeklyIndicatorStatusesByWeekRow{}, err
	}

	return items, nil
}

func (s *SurveillanceService) GetNationalWeeklyStatusSummary(ctx context.Context, epiWeekID uuid.UUID) ([]db.GetNationalWeeklyStatusSummaryRow, error) {
	s.log.Debug(ctx, "getting national weekly status summary")

	items, err := s.dashboardRepo.GetNationalWeeklyStatusSummary(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to get national weekly status summary", "error", err)
		return []db.GetNationalWeeklyStatusSummaryRow{}, err
	}

	return items, nil
}

func (s *SurveillanceService) GetDiseaseDashboardSummaryByWeek(ctx context.Context, arg db.GetDiseaseDashboardSummaryByWeekParams) (db.GetDiseaseDashboardSummaryByWeekRow, error) {
	s.log.Debug(ctx, "getting disease dashboard summary by week")

	item, err := s.dashboardRepo.GetDiseaseDashboardSummaryByWeek(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to get disease dashboard summary by week", "error", err)
		return db.GetDiseaseDashboardSummaryByWeekRow{}, err
	}

	return item, nil
}

func (s *SurveillanceService) GetIndicatorDashboardSummaryByWeek(ctx context.Context, arg db.GetIndicatorDashboardSummaryByWeekParams) (db.GetIndicatorDashboardSummaryByWeekRow, error) {
	s.log.Debug(ctx, "getting indicator dashboard summary by week")

	item, err := s.dashboardRepo.GetIndicatorDashboardSummaryByWeek(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to get indicator dashboard summary by week", "error", err)
		return db.GetIndicatorDashboardSummaryByWeekRow{}, err
	}

	return item, nil
}

func (s *SurveillanceService) GetTopFacilitiesByDiseaseAndWeek(ctx context.Context, arg db.GetTopFacilitiesByDiseaseAndWeekParams) ([]db.GetTopFacilitiesByDiseaseAndWeekRow, error) {
	s.log.Debug(ctx, "getting top facilities by disease and week")

	items, err := s.dashboardRepo.GetTopFacilitiesByDiseaseAndWeek(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to get top facilities by disease and week", "error", err)
		return []db.GetTopFacilitiesByDiseaseAndWeekRow{}, err
	}

	return items, nil
}

func (s *SurveillanceService) GetTopFacilitiesByIndicatorAndWeek(ctx context.Context, arg db.GetTopFacilitiesByIndicatorAndWeekParams) ([]db.GetTopFacilitiesByIndicatorAndWeekRow, error) {
	s.log.Debug(ctx, "getting top facilities by indicator and week")

	items, err := s.dashboardRepo.GetTopFacilitiesByIndicatorAndWeek(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to get top facilities by indicator and week", "error", err)
		return []db.GetTopFacilitiesByIndicatorAndWeekRow{}, err
	}

	return items, nil
}

func (s *SurveillanceService) GetAlertByID(ctx context.Context, id uuid.UUID) (db.Alert, error) {
	s.log.Debug(ctx, "getting alert by id", "alert_id", id)

	item, err := s.alertRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error(ctx, "failed to get alert by id", "alert_id", id, "error", err)
		return db.Alert{}, err
	}

	return item, nil
}

func (s *SurveillanceService) ListAlertsByDisease(ctx context.Context, diseaseID uuid.UUID) ([]db.ListAlertsByDiseaseRow, error) {
	s.log.Debug(ctx, "listing alerts by disease", "disease_id", diseaseID)

	items, err := s.alertRepo.ListByDisease(ctx, diseaseID)
	if err != nil {
		s.log.Error(ctx, "failed to list alerts by disease", "disease_id", diseaseID, "error", err)
		return []db.ListAlertsByDiseaseRow{}, err
	}

	return items, nil
}

func (s *SurveillanceService) ListAlertsByDistrict(ctx context.Context, districtID uuid.UUID) ([]db.ListAlertsByDistrictRow, error) {
	s.log.Debug(ctx, "listing alerts by district")

	items, err := s.alertRepo.ListByDistrict(ctx, districtID)
	if err != nil {
		s.log.Error(ctx, "failed to list alerts by district", "error", err)
		return []db.ListAlertsByDistrictRow{}, err
	}

	return items, nil
}

func (s *SurveillanceService) ListAlertsByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListAlertsByWeekRow, error) {
	s.log.Debug(ctx, "listing alerts by week")

	items, err := s.alertRepo.ListByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to list alerts by week", "error", err)
		return []db.ListAlertsByWeekRow{}, err
	}

	return items, nil
}

func (s *SurveillanceService) ResolveDiseaseByName(ctx context.Context, name string) (db.Disease, error) {
	return s.lookupRepo.ResolveDiseaseByName(ctx, name)
}

func (s *SurveillanceService) ResolveIndicatorByName(ctx context.Context, name string) (db.Indicator, error) {
	return s.lookupRepo.ResolveIndicatorByName(ctx, name)
}

func (s *SurveillanceService) ResolveWeekByYearWeek(ctx context.Context, arg db.ResolveWeekByYearWeekParams) (db.EpiWeek, error) {
	return s.lookupRepo.ResolveWeekByYearWeek(ctx, arg)
}

func (s *SurveillanceService) ResolveDistrictByName(ctx context.Context, name string) (db.District, error) {
	return s.lookupRepo.ResolveDistrictByName(ctx, name)
}

func (s *SurveillanceService) ResolveRegionByName(ctx context.Context, name string) (db.Region, error) {
	return s.lookupRepo.ResolveRegionByName(ctx, name)
}

func (s *SurveillanceService) ResolveSubCountyByNameAndDistrict(ctx context.Context, arg db.ResolveSubCountyByNameAndDistrictParams) (db.SubCounty, error) {
	return s.lookupRepo.ResolveSubCountyByNameAndDistrict(ctx, arg)
}

func (s *SurveillanceService) ResolveFacilityByExternalID(ctx context.Context, externalID string) (db.Facility, error) {
	return s.lookupRepo.ResolveFacilityByExternalID(ctx, externalID)
}

func (s *SurveillanceService) ResolveFacilityByNameAndDistrict(ctx context.Context, arg db.ResolveFacilityByNameAndDistrictParams) (db.Facility, error) {
	return s.lookupRepo.ResolveFacilityByNameAndDistrict(ctx, arg)
}
