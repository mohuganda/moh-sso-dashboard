package surveillance

import (
	"github.com/moh-sso-dashboard/internal/config"
	logger "github.com/moh-sso-dashboard/internal/log"
	baseservice "github.com/moh-sso-dashboard/internal/service"
)

type DiseaseService = baseservice.SurveillanceDiseaseService
type EpiWeekService = baseservice.SurveillanceEpiWeekService
type LocationService = baseservice.SurveillanceLocationService
type FacilityService = baseservice.SurveillanceFacilityService
type FacilityWeeklyMetricsService = baseservice.SurveillanceFacilityWeeklyMetricsService
type WeeklyStatusService = baseservice.SurveillanceWeeklyStatusService
type ImportService = baseservice.SurveillanceImportService
type AlertService = baseservice.SurveillanceAlertService

func NewDiseaseService(
	log *logger.Logger,
	repo DiseaseRepository,
	notifications baseservice.NotificationsService,
	cfg ...*config.Config,
) *DiseaseService {
	return baseservice.NewSurveillanceDiseaseService(log, repo, notifications, cfg...)
}

func NewEpiWeekService(
	log *logger.Logger,
	repo EpiWeekRepository,
	notifications ...baseservice.NotificationsService,
) *EpiWeekService {
	return baseservice.NewSurveillanceEpiWeekService(log, repo, notifications...)
}

func NewLocationService(
	log *logger.Logger,
	regionRepo RegionRepository,
	districtRepo DistrictRepository,
	subCountyRepo SubCountyRepository,
	notifications baseservice.NotificationsService,
	cfg ...*config.Config,
) *LocationService {
	return baseservice.NewSurveillanceLocationService(
		log,
		regionRepo,
		districtRepo,
		subCountyRepo,
		notifications,
		cfg...,
	)
}

func NewFacilityService(
	log *logger.Logger,
	repo FacilityRepository,
	notifications ...baseservice.NotificationsService,
) *FacilityService {
	return baseservice.NewSurveillanceFacilityService(log, repo, notifications...)
}

func NewFacilityWeeklyMetricsService(
	log *logger.Logger,
	repo FacilityWeeklyMetricsRepository,
	importRepo ImportRepository,
) *FacilityWeeklyMetricsService {
	return baseservice.NewSurveillanceFacilityWeeklyMetricsService(log, repo, importRepo)
}

func NewWeeklyStatusService(
	log *logger.Logger,
	repo WeeklyStatusRepository,
) *WeeklyStatusService {
	return baseservice.NewSurveillanceWeeklyStatusService(log, repo)
}

func NewImportService(repo ImportRepository) *ImportService {
	return baseservice.NewSurveillanceImportService(repo)
}

func NewAlertService(
	log *logger.Logger,
	repo AlertRepository,
	importRepo ImportRepository,
	notifications baseservice.NotificationsService,
	cfg ...*config.Config,
) *AlertService {
	return baseservice.NewSurveillanceAlertService(log, repo, importRepo, notifications, cfg...)
}
