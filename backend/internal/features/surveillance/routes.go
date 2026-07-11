package surveillance

import (
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
	"github.com/moh-sso-dashboard/internal/ratelimit"
)

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler, limiter *ratelimit.Limiter) {
	surveillance := protected.Group("/surveillance")
	surveillance.Use(middleware.RequireSystem(authz.SystemDataStatistics))
	{
		read := middleware.RequirePermission(authz.PermissionSurveillanceRead)
		importData := middleware.RequirePermission(authz.PermissionSurveillanceImport)
		manageLocations := middleware.RequirePermission(authz.PermissionSurveillanceManageLocations)
		importLimit := ratelimit.MiddlewareForPolicy(limiter, ratelimit.SurveillanceImportPolicy())
		manageLimit := ratelimit.MiddlewareForPolicy(limiter, ratelimit.SurveillanceManagePolicy())

		surveillance.GET("/weeks", read, handler.ListEpiWeeksByYear)
		surveillance.GET("/alerts", read, handler.ListAlerts)
		surveillance.GET("/diseases", read, handler.ListDiseases)
		surveillance.GET("/regions", read, handler.ListRegions)

		surveillance.GET("/districts", read, handler.ListDistricts)
		surveillance.GET("/regions/:regionID/districts", read, handler.ListDistrictsByRegion)
		surveillance.POST("/districts", manageLocations, manageLimit, handler.UpsertDistrict)

		surveillance.GET("/districts/:districtID/subcounties", read, handler.ListSubcountiesByDistrict)
		surveillance.GET("/subcounties/:id", read, handler.GetSubcountyByID)
		surveillance.POST("/subcounties", manageLocations, manageLimit, handler.UpsertSubcounty)
		surveillance.DELETE("/subcounties/:id", manageLocations, manageLimit, handler.DeleteSubcounty)

		surveillance.GET("/facility-weekly-metrics/week/:epiWeekID", read, handler.ListFacilityWeeklyMetricsByWeek)
		surveillance.GET("/facility-weekly-metrics/facility/:facilityID", read, handler.ListFacilityWeeklyMetricsByFacility)
		surveillance.GET("/facility-weekly-metrics/facility/:facilityID/disease/:diseaseID/trend", read, handler.ListFacilityDiseaseMetricsTrend)
		surveillance.GET("/facility-weekly-metrics/facility/:facilityID/indicator/:indicatorID/trend", read, handler.ListFacilityIndicatorMetricsTrend)
		surveillance.GET("/facility-weekly-metrics/week/:epiWeekID/disease/:diseaseID", read, handler.ListFacilityDiseaseMetricsByWeekAndDisease)
		surveillance.GET("/facility-weekly-metrics/disease-trend", read, handler.ListDiseaseWeeklyTrendAggregated)

		surveillance.GET("/weekly-statuses/list", read, handler.ListWeeklyStatuses)
		surveillance.GET("/weekly-statuses/detailed", read, handler.ListWeeklyStatusesDetailed)
		surveillance.GET("/weekly-statuses/district/week/:epiWeekID", read, handler.ListDistrictWeeklyStatusesByWeek)
		surveillance.GET("/weekly-statuses/region/week/:epiWeekID", read, handler.ListRegionWeeklyStatusesByWeek)
		surveillance.GET("/weekly-statuses/national/week/:epiWeekID", read, handler.ListNationalWeeklyStatusesByWeek)

		surveillance.GET("/imports", read, handler.ListImportBatches)
		surveillance.POST("/imports", importData, importLimit, handler.CreateImportBatch)
		surveillance.GET("/imports/:batchID", read, handler.GetImportBatchByID)
		surveillance.PATCH("/imports/:batchID/status", importData, importLimit, handler.UpdateImportBatchStatus)
		surveillance.GET("/imports/:batchID/raw-rows", read, handler.ListImportRawRowsByBatch)
	}
}
