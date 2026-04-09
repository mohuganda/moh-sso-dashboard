package repository

import (
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/postgres"
)

type Repositories struct {
	Diseases        interfaces.DiseaseRepository
	Indicators      interfaces.IndicatorRepository
	Regions         interfaces.RegionRepository
	Districts       interfaces.DistrictRepository
	SubCounties     interfaces.SubCountyRepository
	Facilities      interfaces.FacilityRepository
	EpiWeeks        interfaces.EpiWeekRepository
	FacilityMetrics interfaces.FacilityWeeklyMetricsRepository
	WeeklyStatus    interfaces.WeeklyStatusRepository
	Dashboard       interfaces.DashboardRepository
	Alerts          interfaces.AlertRepository
	Lookup          interfaces.LookupRepository
	Imports         interfaces.ImportRepository
}

func NewRepositories(db db.Store) *Repositories {
	return &Repositories{
		Diseases:        postgres.NewDiseaseRepository(db),
		Indicators:      postgres.NewIndicatorRepository(db),
		Regions:         postgres.NewRegionRepository(db),
		Districts:       postgres.NewDistrictRepository(db),
		SubCounties:     postgres.NewSubCountyRepository(db),
		Facilities:      postgres.NewFacilityRepository(db),
		EpiWeeks:        postgres.NewEpiWeekRepository(db),
		FacilityMetrics: postgres.NewFacilityWeeklyMetricsRepository(db),
		WeeklyStatus:    postgres.NewWeeklyStatusRepository(db),
		Dashboard:       postgres.NewDashboardRepository(db),
		Alerts:          postgres.NewAlertRepository(db),
		Lookup:          postgres.NewLookupRepository(db),
		Imports:         postgres.NewImportRepository(db),
	}
}
