package routes

import "github.com/gin-gonic/gin"

func registerSurveillanceRoutes(protected *gin.RouterGroup, deps Dependencies) {
	surveillance := protected.Group("/surveillance")
	{
		surveillance.GET("/weeks", deps.Surveillance.ListEpiWeeksByYear)
		surveillance.GET("/alerts", deps.Surveillance.ListAlerts)
		surveillance.GET("/diseases", deps.Surveillance.ListDiseases)
		surveillance.GET("/regions", deps.Surveillance.ListRegions)

		surveillance.GET("/districts", deps.Surveillance.ListDistricts)
		surveillance.GET("/regions/:regionID/districts", deps.Surveillance.ListDistrictsByRegion)
		surveillance.POST("/districts", deps.Surveillance.UpsertDistrict)

		surveillance.GET("/districts/:districtID/subcounties", deps.Surveillance.ListSubcountiesByDistrict)
		surveillance.GET("/subcounties/:id", deps.Surveillance.GetSubcountyByID)
		surveillance.POST("/subcounties", deps.Surveillance.UpsertSubcounty)
		surveillance.DELETE("/subcounties/:id", deps.Surveillance.DeleteSubcounty)

		surveillance.GET("/facility-weekly-metrics/week/:epiWeekID", deps.Surveillance.ListFacilityWeeklyMetricsByWeek)
		surveillance.GET("/facility-weekly-metrics/facility/:facilityID", deps.Surveillance.ListFacilityWeeklyMetricsByFacility)
		surveillance.GET("/facility-weekly-metrics/facility/:facilityID/disease/:diseaseID/trend", deps.Surveillance.ListFacilityDiseaseMetricsTrend)
		surveillance.GET("/facility-weekly-metrics/facility/:facilityID/indicator/:indicatorID/trend", deps.Surveillance.ListFacilityIndicatorMetricsTrend)
		surveillance.GET("/facility-weekly-metrics/week/:epiWeekID/disease/:diseaseID", deps.Surveillance.ListFacilityDiseaseMetricsByWeekAndDisease)
		surveillance.GET("/facility-weekly-metrics/disease-trend", deps.Surveillance.ListDiseaseWeeklyTrendAggregated)

		surveillance.GET("/weekly-statuses/list", deps.Surveillance.ListWeeklyStatuses)
		surveillance.GET("/weekly-statuses/detailed", deps.Surveillance.ListWeeklyStatusesDetailed)
		surveillance.GET("/weekly-statuses/district/week/:epiWeekID", deps.Surveillance.ListDistrictWeeklyStatusesByWeek)
		surveillance.GET("/weekly-statuses/region/week/:epiWeekID", deps.Surveillance.ListRegionWeeklyStatusesByWeek)
		surveillance.GET("/weekly-statuses/national/week/:epiWeekID", deps.Surveillance.ListNationalWeeklyStatusesByWeek)

		surveillance.GET("/imports", deps.Surveillance.ListImportBatches)
		surveillance.POST("/imports", deps.Surveillance.CreateImportBatch)
		surveillance.GET("/imports/:batchID", deps.Surveillance.GetImportBatchByID)
		surveillance.PATCH("/imports/:batchID/status", deps.Surveillance.UpdateImportBatchStatus)
		surveillance.GET("/imports/:batchID/raw-rows", deps.Surveillance.ListImportRawRowsByBatch)
	}
}
