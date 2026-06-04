package routes

import "github.com/gin-gonic/gin"

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
	email := protected.Group("/emails")
	{
		email.POST("/send", deps.Email.Send)
		email.POST("/queue", deps.Email.Queue)
		email.GET("", deps.Email.List)
		email.GET("/status/:status", deps.Email.ListByStatus)
		email.GET("/:id", deps.Email.GetByID)
		email.POST("/:id/retry", deps.Email.Retry)
		email.DELETE("/:id", deps.Email.Delete)
	}
}

func registerGeoJSONRoutes(protected *gin.RouterGroup, deps Dependencies) {
	geojson := protected.Group("/geojson")
	{
		geojson.GET("/:name", deps.GeoJSON.GetGeoJSON)
	}
}

func registerClientRoutes(protected *gin.RouterGroup, deps Dependencies) {
	clients := protected.Group("/clients")
	{
		clients.GET("", deps.Clients.ListClients)
		clients.GET("/:id", deps.Clients.GetClient)
		clients.POST("", deps.Clients.CreateClient)
		clients.PATCH("/:id/toggle", deps.Clients.ToggleClientEnabled)
		clients.DELETE("/:id", deps.Clients.DeleteClient)
		clients.GET("/:id/roles", deps.Clients.ListClientRoles)
	}
}

func registerUserRoutes(protected *gin.RouterGroup, deps Dependencies) {
	users := protected.Group("/users")
	{
		users.GET("", deps.Users.ListUsers)
		users.GET("/:id", deps.Users.GetUser)
	}
}

func registerStorageLocationRoutes(protected *gin.RouterGroup, deps Dependencies) {
	storageLocations := protected.Group("/storage-locations")
	{
		storageLocations.POST("", deps.StorageLocations.Create)
		storageLocations.GET("", deps.StorageLocations.ListActive)
		storageLocations.GET("/:id", deps.StorageLocations.GetByID)
		storageLocations.PUT("/:id", deps.StorageLocations.Update)
		storageLocations.DELETE("/:id", deps.StorageLocations.Delete)
	}
}

func registerSessionRoutes(protected *gin.RouterGroup, deps Dependencies) {
	sessions := protected.Group("/sessions")
	{
		sessions.GET("", deps.Sessions.GetUserSessions)
		sessions.DELETE("/:id", deps.Sessions.LogoutSession)
	}
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
