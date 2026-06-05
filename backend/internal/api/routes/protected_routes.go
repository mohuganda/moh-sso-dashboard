package routes

import (
	"github.com/gin-gonic/gin"

	clientfeature "github.com/moh-sso-dashboard/internal/features/clients"
	emailfeature "github.com/moh-sso-dashboard/internal/features/email"
	sessionfeature "github.com/moh-sso-dashboard/internal/features/sessions"
	storagelocationfeature "github.com/moh-sso-dashboard/internal/features/storage_locations"
	userfeature "github.com/moh-sso-dashboard/internal/features/users"
)

func RegisterProtectedRoutes(protected *gin.RouterGroup, deps Dependencies) {
	registerEmailRoutes(protected, deps)
	registerGeoJSONRoutes(protected, deps)
	registerClientRoutes(protected, deps)
	registerUserRoutes(protected, deps)
	registerDocumentRoutes(protected, deps)
	registerStorageLocationRoutes(protected, deps)
	registerSessionRoutes(protected, deps)
	registerDataQualityRoutes(protected, deps)
	registerVisualiserRoutes(protected, deps)
	registerSurveillanceRoutes(protected, deps)
}

func registerEmailRoutes(protected *gin.RouterGroup, deps Dependencies) {
	emailfeature.RegisterProtectedRoutes(protected, deps.Email)
}

func registerGeoJSONRoutes(protected *gin.RouterGroup, deps Dependencies) {
	geojson := protected.Group("/geojson")
	{
		geojson.GET("/:name", deps.GeoJSON.GetGeoJSON)
	}
}

func registerClientRoutes(protected *gin.RouterGroup, deps Dependencies) {
	clientfeature.RegisterProtectedRoutes(protected, deps.Clients)
}

func registerUserRoutes(protected *gin.RouterGroup, deps Dependencies) {
	userfeature.RegisterProtectedRoutes(protected, deps.Users)
}

func registerStorageLocationRoutes(protected *gin.RouterGroup, deps Dependencies) {
	storagelocationfeature.RegisterProtectedRoutes(protected, deps.StorageLocations)
}

func registerSessionRoutes(protected *gin.RouterGroup, deps Dependencies) {
	sessionfeature.RegisterProtectedRoutes(protected, deps.Sessions)
}

func registerDataQualityRoutes(protected *gin.RouterGroup, deps Dependencies) {
	issues := protected.Group("/issues")
	{
		issues.POST("", deps.DataQuality.CreateIssue)
		issues.GET("", deps.DataQuality.ListIssues)
		issues.PUT("/:issueCode", deps.DataQuality.UpdateIssue)
		issues.POST("/:issueCode/resolveIssue", deps.DataQuality.ResolveIssue)
		issues.GET("/:issueCode/transactions", deps.DataQuality.ListIssueResolutionTransactions)
	}
}
