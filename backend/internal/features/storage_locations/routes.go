package storage_locations

import "github.com/gin-gonic/gin"

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler) {
	storageLocations := protected.Group("/storage-locations")
	{
		storageLocations.POST("", handler.Create)
		storageLocations.GET("", handler.ListActive)
		storageLocations.GET("/:id", handler.GetByID)
		storageLocations.PUT("/:id", handler.Update)
		storageLocations.DELETE("/:id", handler.Delete)
	}
}
