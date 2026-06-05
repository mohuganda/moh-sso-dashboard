package surveillance

import "github.com/gin-gonic/gin"

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler) {
	surveillance := protected.Group("/surveillance")
	{
		surveillance.GET("/weeks", handler.ListEpiWeeksByYear)
		surveillance.GET("/alerts", handler.ListAlerts)
		surveillance.GET("/diseases", handler.ListDiseases)
		surveillance.GET("/regions", handler.ListRegions)

		surveillance.GET("/districts", handler.ListDistricts)
		surveillance.GET("/regions/:regionID/districts", handler.ListDistrictsByRegion)
		surveillance.POST("/districts", handler.UpsertDistrict)

		surveillance.GET("/districts/:districtID/subcounties", handler.ListSubcountiesByDistrict)
		surveillance.GET("/subcounties/:id", handler.GetSubcountyByID)
		surveillance.POST("/subcounties", handler.UpsertSubcounty)
		surveillance.DELETE("/subcounties/:id", handler.DeleteSubcounty)

		surveillance.GET("/facility-weekly-metrics/week/:epiWeekID", handler.ListFacilityWeeklyMetricsByWeek)
		surveillance.GET("/facility-weekly-metrics/facility/:facilityID", handler.ListFacilityWeeklyMetricsByFacility)
		surveillance.GET("/facility-weekly-metrics/facility/:facilityID/disease/:diseaseID/trend", handler.ListFacilityDiseaseMetricsTrend)
		surveillance.GET("/facility-weekly-metrics/facility/:facilityID/indicator/:indicatorID/trend", handler.ListFacilityIndicatorMetricsTrend)
		surveillance.GET("/facility-weekly-metrics/week/:epiWeekID/disease/:diseaseID", handler.ListFacilityDiseaseMetricsByWeekAndDisease)
		surveillance.GET("/facility-weekly-metrics/disease-trend", handler.ListDiseaseWeeklyTrendAggregated)

		surveillance.GET("/weekly-statuses/list", handler.ListWeeklyStatuses)
		surveillance.GET("/weekly-statuses/detailed", handler.ListWeeklyStatusesDetailed)
		surveillance.GET("/weekly-statuses/district/week/:epiWeekID", handler.ListDistrictWeeklyStatusesByWeek)
		surveillance.GET("/weekly-statuses/region/week/:epiWeekID", handler.ListRegionWeeklyStatusesByWeek)
		surveillance.GET("/weekly-statuses/national/week/:epiWeekID", handler.ListNationalWeeklyStatusesByWeek)

		surveillance.GET("/imports", handler.ListImportBatches)
		surveillance.POST("/imports", handler.CreateImportBatch)
		surveillance.GET("/imports/:batchID", handler.GetImportBatchByID)
		surveillance.PATCH("/imports/:batchID/status", handler.UpdateImportBatchStatus)
		surveillance.GET("/imports/:batchID/raw-rows", handler.ListImportRawRowsByBatch)
	}
}
